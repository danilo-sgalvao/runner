# Conclusão da Sprint 2

Resumo do estado atual em relação ao [`plano-sprints.md`](./plano-sprints.md) (Sprint 2) e plano de ação para fechar os critérios pendentes.

## 1. Status por User Story

| US | Cumpridos | Pendentes |
|----|-----------|-----------|
| US-02.1 — Simulação de `sign` | 5/5 | — |
| US-02.2 — Validação de parâmetros de `sign` | 4/4 | — |
| US-02.3 — Simulação/validação de `validate` | 4/4 | — |
| US-01.2 — Parsing CLI | 5/5 | — |
| US-01.3 — Invocação local do JAR | 5/5 | — |
| US-01.4 — Exibição legível | 4/4 | — |
| US-04.1 — Provisionamento de JRE | 5/5 | — |

## 2. Pontos para fechar 100% da Sprint 2

Todos resolvidos:

1. ✅ **Refatoração** `SignatureService`/`FakeSignatureService` implementada (PR 1).
2. ✅ **Testes de integração** end-to-end adicionados com build tag `integration` (PR 2).
3. ✅ **Critério US-04.1** atualizado de JDK para JRE em `plano-sprints.md` (PR 3).

---

## 3. Planejamento — Refatoração para `SignatureService` / `FakeSignatureService`

**Origem:** US-02.1 (critérios pendentes) + decisão registrada de seguir o desenho de `referencia/assinador-java/`.

### 3.1. Comparação entre estado atual e referência

| Aspecto | Atual (`projetos/assinador-java`) | Referência (`referencia/assinador-java`) |
|---------|-----------------------------------|-------------------------------------------|
| Pacote raiz | `com.hubsaude.assinador` | `com.kyriosdata.assinador` |
| API de serviço | Classe `AssinadorService` com métodos **estáticos** `sign(String, String)`, `validate(String, String)` | Interface `SignatureService` com `sign(SignRequest)`, `validate(ValidateRequest)` |
| Implementação | A própria `AssinadorService` (sem polimorfismo) | Classe `FakeSignatureService implements SignatureService` |
| DTOs | Nenhum — parâmetros primitivos | `SignRequest { content, token }`, `ValidateRequest { content, signature }` em `com.kyriosdata.assinador.domain` |
| Retorno | `sign` → `String`; `validate` → `boolean`; erros via `IllegalArgumentException` | Ambos retornam `SignatureResponse { signature, valid, message }`; erros como campo `valid=false` + `message` |
| Lógica de assinatura | Hash determinístico: `ASSINATURA-SIMULADA-<algoritmo>-<hexHash>` | Constante fixa: `MOCKED_SIGNATURE_BASE64_==` |
| Algoritmo | Parâmetro `--algorithm` aceita `SHA256withRSA` ou `SHA512withRSA` | **Não existe** — sem suporte a algoritmo |
| Token | Não suportado | Campo opcional em `SignRequest` (não usado na simulação) |
| Testes | `AssinadorServiceTest` (17 testes) | `FakeSignatureServiceTest` (4 testes) |

### 3.2. Decisões tomadas

- **D1. Pacote raiz:** **manter `com.hubsaude.assinador`** — preserva a consistência com o restante do repositório. O pacote da referência (`com.kyriosdata.assinador`) é apenas inspirador da estrutura/nomes de classe.
- **D2. Parâmetro `--algorithm`:** **remover** do serviço Java (`AssinadorService`) e do flag do CLI Go (`sign.go`) — é simulação sem cripto real e a referência o descarta. Documentar a mudança como breaking change no README.
- **D3. Lógica de simulação:** **adotar a constante** `FAKE_SIGNATURE = "MOCKED_SIGNATURE_BASE64_=="` da referência. Testes que dependiam do hash determinístico variável por conteúdo (ex.: "conteúdos distintos → assinaturas distintas") serão removidos; os de presença/validação de campos continuam aplicáveis.
- **D4. Formato de saída do `Main`:** **serializar `SignatureResponse` como JSON** simples montado via `String.format` (sem dependência externa). Substitui o atual `chave=valor` e antecipa o futuro modo HTTP previsto em `CLAUDE.md`.

### 3.3. Plano de implementação

Sequência sugerida; cada item entrega um estado compilável e testável.

#### Passo 1 — Criar DTOs em `domain`
Novos arquivos em `projetos/assinador-java/src/main/java/com/hubsaude/assinador/domain/`:
- `SignRequest.java` — campos `content`, `token` (getters/setters).
- `ValidateRequest.java` — campos `content`, `signature` (getters/setters).
- `SignatureResponse.java` — campos `signature`, `valid` (boolean), `message`.

Cópias adaptadas da referência (somente troca do `package`). Sem lógica.

#### Passo 2 — Criar interface `SignatureService`
```java
public interface SignatureService {
    SignatureResponse sign(SignRequest request);
    SignatureResponse validate(ValidateRequest request);
}
```

#### Passo 3 — Criar `FakeSignatureService implements SignatureService`
- Replicar a lógica da referência: validação de presença de `content` e (em `validate`) `signature`, retornando `SignatureResponse` com `valid=false` + `message` quando inválido.
- Constante `FAKE_SIGNATURE = "MOCKED_SIGNATURE_BASE64_=="` (D3).

#### Passo 4 — Remover `AssinadorService.java`
Deletar após `Main.java` e testes migrarem (passos 5 e 6).

#### Passo 5 — Atualizar `Main.java`
- Trocar chamadas estáticas por instância: `SignatureService service = new FakeSignatureService();`
- Construir `SignRequest`/`ValidateRequest` a partir dos flags.
- Remover parsing de `--algorithm` (D2).
- Tratar `SignatureResponse`: se `response.isValid()`, imprimir o JSON em stdout e `exit(0)`; senão imprimir o JSON em stderr e `exit(1)`.
- Formato de saída JSON via `String.format` (D4), p. ex.:
  ```
  {"signature":"MOCKED_SIGNATURE_BASE64_==","valid":true,"message":"Assinatura criada com sucesso"}
  ```
  Escapar `"` e `\` no campo `message` antes da interpolação para evitar JSON inválido.

#### Passo 6 — Migrar testes
- Renomear `AssinadorServiceTest.java` → `FakeSignatureServiceTest.java` (ou criar o novo e remover o antigo).
- Reescrever asserções para o novo contrato (DTOs + `SignatureResponse`).
- **Cobertura mínima:** os 4 cenários da referência **mais** os casos preservados que ela não cobre: `content` vazio e apenas espaços-em-branco (`isBlank()`), fluxo `sign → validate` end-to-end. Os testes de `--algorithm` e os de "conteúdos distintos → assinaturas distintas" são descartados (D2 e D3).

#### Passo 7 — Ajustar o CLI Go
- Em `projetos/assinatura/cmd/sign.go`: remover flag `--algorithm` e a passagem do parâmetro ao subprocesso Java (D2).
- Em `projetos/assinatura/cmd/sign_test.go`: remover assertivas que dependem do flag.
- Saída do JAR é repassada via `os.Stdout`/`os.Stderr`, então o JSON do Passo 5 chega ao usuário sem mudança adicional no Go.

> **Nota sobre `pom.xml`:** nenhuma alteração necessária. `groupId=com.hubsaude` (D1), `mainClass=com.hubsaude.assinador.Main` e `finalName=assinador` permanecem iguais.

#### Passo 8 — Documentação
Conforme convenção do projeto (atualizar docs após mudanças de código):
- `CLAUDE.md`: atualizar tabela "Key Files" (novas classes `SignatureService`, `FakeSignatureService`, DTOs em `domain/`; remoção de `AssinadorService`).
- `README.md`: remover `--algorithm` dos exemplos; mencionar a constante de assinatura simulada se aplicável.
- `docs/especificacao.md`: revisar US-02 para refletir o contrato baseado em DTOs e saída JSON.
- Marcar como `[x]` os critérios pendentes de US-02.1 em `plano-sprints.md`.

### 3.4. Critérios de aceitação da refatoração

- [ ] Interface `SignatureService` existe com os métodos `sign(SignRequest)` e `validate(ValidateRequest)`.
- [ ] Classe `FakeSignatureService` implementa `SignatureService` e contém a lógica de simulação.
- [ ] DTOs `SignRequest`, `ValidateRequest`, `SignatureResponse` existem no subpacote `domain`.
- [ ] `Main` instancia `FakeSignatureService` e delega para a interface (sem chamadas a métodos estáticos de simulação).
- [ ] Suite de testes verde após migração — cobertura mínima dos 4 cenários da referência **+** os casos preservados (espaços-em-branco, fluxo end-to-end).
- [ ] `mvn package` produz `target/assinador.jar` consumível pelo CLI Go sem alteração de `encontrarJar()`.
- [ ] CLI Go executa `assinatura sign --content "x"` e `assinatura validate --content "x" --signature "MOCKED_SIGNATURE_BASE64_=="` com saída coerente.

### 3.5. Riscos e mitigações

| Risco | Mitigação |
|-------|-----------|
| Remoção de `--algorithm` quebra usuários/scripts existentes | Projeto em fase de coursework, sem usuários externos — comunicar a mudança em README; remover sem deprecation. |
| CLI Go continua passando `--algorithm` para um JAR que não aceita mais | Mudar Java e Go no **mesmo PR** (PR 1, ver seção 4); rodar `go test ./...` e `mvn test` antes de commitar. |
| Saída JSON do `Main` malformada por aspas/barras na `message` | Função utilitária `escapeJson(String)` aplicada antes da interpolação; teste unitário cobrindo `message` com `"` e `\`. |
| Teste de integração end-to-end (US-01.3, PR 2) falha no CI por ausência de Java/JRE | Reutilizar o auto-download de JRE já implementado, ou marcar o teste com build tag `//go:build integration` e rodá-lo só em jobs específicos do CI. |

### 3.6. Estimativa

Trabalho contido — entre 2 e 4 horas para o PR 1, considerando: criação dos 4 arquivos novos, edição de `Main.java`, migração dos testes Java, ajuste dos testes Go e atualização da documentação. PR 2 (~1h) e PR 3 (~10min) somam pouco.

---

## 4. Divisão em PRs

Os três pontos pendentes são independentes; ficam mais limpos como PRs separados.

### PR 1 — Refatoração `SignatureService` + remoção de `--algorithm` (US-02.1)

**Escopo:** Passos 1 a 8 da seção 3.3. Java e Go alterados juntos porque a flag `--algorithm` precisa sumir dos dois lados no mesmo commit, senão o CLI passa um argumento que o JAR ignora.

**Arquivos tocados:**
- `projetos/assinador-java/src/main/java/com/hubsaude/assinador/domain/SignRequest.java` (novo)
- `projetos/assinador-java/src/main/java/com/hubsaude/assinador/domain/ValidateRequest.java` (novo)
- `projetos/assinador-java/src/main/java/com/hubsaude/assinador/domain/SignatureResponse.java` (novo)
- `projetos/assinador-java/src/main/java/com/hubsaude/assinador/SignatureService.java` (novo)
- `projetos/assinador-java/src/main/java/com/hubsaude/assinador/FakeSignatureService.java` (novo)
- `projetos/assinador-java/src/main/java/com/hubsaude/assinador/AssinadorService.java` (removido)
- `projetos/assinador-java/src/main/java/com/hubsaude/assinador/Main.java` (editado)
- `projetos/assinador-java/src/test/java/com/hubsaude/assinador/AssinadorServiceTest.java` → `FakeSignatureServiceTest.java`
- `projetos/assinatura/cmd/sign.go` (editado — remover flag)
- `projetos/assinatura/cmd/sign_test.go` (editado)
- `CLAUDE.md`, `README.md`, `docs/especificacao.md`, `docs/plano-sprints.md` (docs)

**Critério de pronto:** `mvn test` e `go test ./...` verdes; `mvn package` produz `target/assinador.jar`; CLI executa `sign` e `validate` com saída JSON; US-02.1 marcada `[x]` em `plano-sprints.md`.

### PR 2 — Teste de integração end-to-end (US-01.3)

**Escopo:** Acrescentar 1 teste Go que invoque o `assinador.jar` real (não apenas registra flags do Cobra).

**Estratégia:** Test com build tag `//go:build integration`. Pre-condição: o teste constrói o JAR via `mvn -f ../assinador-java/pom.xml package` num `TestMain` ou pula com `t.Skip` se Java/Maven indisponíveis. Usa `jre.JavaPath()` para garantir Java. Roda o subprocesso e verifica o JSON de saída.

**Arquivos tocados:**
- `projetos/assinatura/cmd/integration_test.go` (novo, com build tag)
- `.github/workflows/build.yml` (opcional — adicionar passo `-tags integration` em job dedicado)

**Critério de pronto:** `go test -tags integration ./...` passa local; US-01.3 marcada `[x]`.

### PR 3 — Correção textual US-04.1 (JDK → JRE)

**Escopo:** Atualizar o critério de US-04.1 em `docs/plano-sprints.md` para refletir a decisão registrada em `plano-download-java.md` (JRE em `~/.hubsaude/jre/`). Atualizar a tabela de status na seção 1 deste documento para retirar a divergência.

**Arquivos tocados:**
- `docs/plano-sprints.md`
- `docs/CONCLUSAO-SPRINT2.MD`

**Critério de pronto:** texto do plano coerente com a implementação; US-04.1 sem ressalvas.

### Ordem de execução

PR 3 pode ir primeiro (independente, trivial). PR 1 em seguida. PR 2 por último, pois é mais simples de escrever depois que o contrato JSON do Passo 5 está consolidado.
