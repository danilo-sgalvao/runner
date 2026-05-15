# Plano de Implementação: Provisionamento Automático do JRE

Este plano descreve como os CLIs em Go (`assinatura` e `simulador`) vão detectar, baixar e configurar o JRE automaticamente, sem que o usuário precise instalar o Java manualmente — atendendo **US-03** e **US-04**.

## Visão geral da estratégia

Em vez de embutir a URL de download do JRE diretamente no código, os CLIs consultam um arquivo `release.json` hospedado neste repositório (branch `main`). Esse arquivo centraliza as URLs de download do JRE por plataforma, permitindo atualizar a versão do Java sem recompilar os binários.

```
CLI inicia
  └─ Java em ~/.hubsaude/jre (versão ok)?     ──► usa esse Java
       └─ não → Java no PATH (versão ok)?     ──► usa esse Java
            └─ não → busca release.json
                      ├─ sucesso → baixa JRE (com barra de progresso)
                      └─ falha (offline) → Java no PATH (qualquer versão)? ──► usa com aviso
                                                └─ não → aborta com mensagem clara
```

## Consequências

> [!WARNING]
> **Tamanho do Download:** O JRE pesa aproximadamente 40–60 MB. O download ocorre apenas na primeira execução em uma máquina sem Java, ou quando a versão local estiver desatualizada.
> **Complexidade adicional:** O Go precisa descompactar `.zip` (Windows) e `.tar.gz` (Linux/macOS) e realizar uma requisição HTTP para buscar o `release.json` antes de qualquer outra coisa.

## Arquivo `release.json`

Criar na raiz do repositório (branch `main`):

```json
{
  "jre": {
    "version": "21",
    "windows_x64": "https://api.adoptium.net/v3/binary/latest/21/ga/windows/x64/jre/hotspot/normal/eclipse",
    "linux_x64":   "https://api.adoptium.net/v3/binary/latest/21/ga/linux/x64/jre/hotspot/normal/eclipse",
    "mac_x64":     "https://api.adoptium.net/v3/binary/latest/21/ga/mac/x64/jre/hotspot/normal/eclipse"
  }
}
```

Para trocar a versão do Java no futuro, basta alterar `"version"` e as URLs neste arquivo — sem recompilar nenhum binário.

## Proposed Changes

### [NEW] `projetos/assinatura/internal/jre/manager.go`

Pacote compartilhável entre os CLIs. Contém toda a lógica de provisionamento do JRE:

1. **`JavaPath() (string, error)`** — ponto de entrada principal; retorna o caminho absoluto do executável `java` pronto para uso.
2. **Detectar local (`~/.hubsaude/jre`):** Se já existe um executável `java` instalado localmente, retorna esse caminho sem fazer rede.
3. **Detectar sistema (PATH):** Se `java` está disponível no PATH, retorna o caminho do sistema.
4. **Buscar `release.json`:** Faz GET em `https://raw.githubusercontent.com/danilo-sgalvao/runner/main/release.json`, desserializa o JSON. Se a requisição falhar (sem rede), verifica se há algum `java` no PATH — qualquer versão — e o usa com um aviso ao usuário; se não houver nenhum, aborta com mensagem clara.
5. **Selecionar URL por plataforma:** Usa `runtime.GOOS` e `runtime.GOARCH` para escolher `windows_x64`, `linux_x64` ou `mac_x64`.
6. **Baixar JRE:** Faz download do arquivo para um temporário, exibindo uma barra de progresso no terminal.
7. **Extrair:** Descompacta `.zip` (Windows) ou `.tar.gz` (Linux/macOS) em `~/.hubsaude/jre/`.
8. **Retornar caminho:** Retorna o caminho absoluto do `java` recém-instalado.

```
~/.hubsaude/
  jre/
    bin/
      java          (Linux/macOS)
      java.exe      (Windows)
    ...
```

### [MODIFY] `projetos/assinatura/cmd/sign.go`

**De:** `exec.Command("java", "-jar", jarPath, ...)`  
**Para:**
```go
javaPath, err := jre.JavaPath()
// tratar erro
exec.Command(javaPath, "-jar", jarPath, ...)
```

### [MODIFY] `projetos/assinatura/cmd/validate.go`

Mesma alteração de `sign.go`.

### [MODIFY] `projetos/simulador/cmd/*.go` *(quando o CLI simulador for criado)*

O CLI `simulador` (US-03) também usará `jre.JavaPath()` antes de invocar `simulador.jar`, sem duplicar a lógica de provisionamento.

## Diferença em relação ao plano anterior

| Aspecto | Plano anterior | Plano atual |
|---|---|---|
| Versão do Java | Hardcoded (`21`) na URL dentro do binário | Lida do `release.json` no repositório |
| URL de download | Hardcoded no código Go | Centralizada no `release.json` |
| Diretório de instalação | `~/.runner/jre` | `~/.hubsaude/jre` (conforme spec) |
| Escopo | Apenas `assinatura` | `assinatura` + `simulador` (via pacote compartilhado) |
| Atualização de versão | Recompila os binários | Atualiza só o `release.json` |

## Plano de verificação

### Testes automatizados
1. **Sem Java no sistema:** Alterar `PATH` na execução para esconder o Java; verificar que o `release.json` é buscado e o JRE é baixado para `~/.hubsaude/jre`.
2. **JRE já instalado localmente:** Verificar que o CLI usa o JRE local sem fazer requisição de rede.
3. **Java no PATH:** Verificar que o CLI usa o Java do sistema sem baixar nada.
4. **`release.json` indisponível + java no PATH:** Verificar que o CLI usa o java do PATH com aviso, sem abortar.
5. **`release.json` indisponível + sem java no PATH:** Verificar mensagem de erro clara ao usuário.

### Verificação manual
- Compilar o binário Windows.
- Executar em ambiente sem Java instalado (VM limpa ou container).
- Confirmar criação de `~/.hubsaude/jre/` e execução correta da assinatura.
