Plano: download automático do assinador.jar

  Objetivo

  O cliente final recebe apenas o binário assinatura. O assinador.jar é resolvido em runtime via HTTP —
  mesmo modelo já adotado para o JRE.

  Estrutura final esperada em ~/.hubsaude/

  ~/.hubsaude/
  ├── jre/                # JRE gerenciado (já existe)
  │   ├── bin/java[.exe]
  │   └── ...
  └── jar/                # NOVO
      ├── assinador.jar
      └── version.txt     # versão instalada (para invalidar cache)

  Mudanças

  1. release.json — estender com seção do JAR

  JAR é cross-platform → uma URL só. Versão controla invalidação de cache.
  {
    "jre": { ... já existe ... },
    "assinador": {
      "version": "0.3.0",
      "url": "https://github.com/danilo-sgalvao/runner/releases/download/v0.3.0/assinador.jar",
      "sha256": "<hash do jar publicado>"
    }
  }
  Inclusão do sha256 é opcional, mas alinha com o padrão Cosign/Sigstore que já é usado para os binários
  — verifica integridade do download.

  2. Refatorar internal/jre/manager.go para extrair partes reutilizáveis

  Antes de criar o pacote novo, mover para internal/release/:
  - a struct releaseFile (estendida com Assinador)
  - a função fetchRelease()
  - a função downloadWithProgress() (renomeada e exportada como Download)

  Isso evita duplicação entre internal/jre e o novo internal/jar. Mantém os testes existentes verdes
  (apenas reaponta imports).

  3. Novo pacote internal/jar

  Espelhar a estrutura de internal/jre. API pública:
  func JarPath() (string, error)
  Lógica:
  1. Ler ~/.hubsaude/jar/version.txt.
  2. Buscar release.json via release.Fetch().
  3. Se versão local == versão remota e assinador.jar existe → retorna path local.
  4. Senão → baixa, valida sha256 (se presente), salva como assinador.jar, escreve version.txt.
  5. Offline + cache existente → usa cache com aviso (igual ao fallback do JRE).
  6. Offline + sem cache → erro claro (diferente do JRE, não há fallback de sistema).

  4. Simplificar cmd/jar.go::encontrarJar()

  Substituir a busca atual por:
  func encontrarJar() (string, error) {
      // Atalho de desenvolvimento: se o repo está montado, prioriza o JAR local
      local := filepath.Join("..", "assinador-java", "target", "assinador.jar")
      if _, err := os.Stat(local); err == nil {
          return local, nil
      }
      return jar.JarPath()
  }
  Remove o <exe-dir>/assinador.jar (cliente final não tem mais JAR ao lado do binário). Mantém o fallback
   de target/ para go run durante desenvolvimento.

  5. CI/CD — .github/workflows/release.yml

  Adicionar passos:
  - actions/setup-java@v4 (JDK 21) + mvn -B package em projetos/assinador-java/
  - Renomear target/assinador-*.jar → assinador.jar
  - Calcular sha256sum assinador.jar (entra no checksums.txt)
  - Assinar com Cosign (mesmo padrão dos binários)
  - Publicar assinador.jar e .bundle como assets do release

  Fluxo manual de release passa a ser:
  1. Atualizar release.json com assinador.version e assinador.url apontando para a tag prevista (e sha256
   após primeiro build, se quiser ser rigoroso — alternativa: pular sha256 na primeira iteração).
  2. Commitar release.json no main.
  3. git tag vX.Y.Z && git push origin vX.Y.Z.

  6. Testes

  - internal/release/release_test.go — fetch com servidor httptest.
  - internal/jar/manager_test.go — fluxos: download fresh, cache válido, cache desatualizado, offline com
   cache, offline sem cache, sha256 inválido.
  - cmd/jar_test.go — adaptar para nova ordem (dev fallback → JarPath()); usar interface/var injetável
  para mockar jar.JarPath ou apontar HOME para tempdir.

  7. Documentação

  - CLAUDE.md: reescrever seção "JAR discovery order" + diagrama do "Command flow" (incluir jar.JarPath()
   entre JavaPath e spawns).
  - README.md: explicar que o JAR baixa automaticamente; documentar pasta ~/.hubsaude/jar/.
  - docs/plano-download-jar.md: novo, espelhando o plano-download-java.md.

  Ordem sugerida de implementação

  1. Extrair internal/release (refactor sem mudança de comportamento) + atualizar testes do JRE.
  2. Criar internal/jar + testes.
  3. Atualizar cmd/jar.go para usar jar.JarPath().
  4. Atualizar release.json no repositório.
  5. Atualizar release.yml (build/publicação do JAR).
  6. Atualizar CLAUDE.md e README.md.

  Riscos

  - Ovo-galinha do release.json: o CLI lê do main via raw.githubusercontent. A entrada do assinador
  precisa existir no JSON antes do binário tentar usá-la — o primeiro release que ativar o recurso
  precisa de coordenação (atualizar JSON → publicar release → só então binários antigos param de
  funcionar... considerar feature flag ou versão 0.0.0 inicial).
  - Invalidação por versão, não por hash: se você sobrescrever a URL sem subir a versão, clientes com
  cache não atualizam. Documentar regra: toda mudança no JAR exige bump de versão em release.json.
  - Sem fallback de sistema: diferente do Java, não há "JAR do sistema". Sem rede + sem cache = erro
  fatal. A mensagem precisa ser explícita.