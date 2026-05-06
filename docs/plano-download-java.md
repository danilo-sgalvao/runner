# Plano de Implementação: Instalação Automática do Java (JRE)

Este plano descreve como modificaremos o CLI em Go (`assinatura`) para baixar e instalar o Java (JRE) localmente caso o usuário não tenha o Java instalado em sua máquina.

## User Review Required

> [!WARNING]
> **Tamanho do Download:** Baixar o JRE implicará em um download de aproximadamente 40MB a 60MB na primeira execução em uma máquina sem Java.
> **Complexidade do Go:** A extração de arquivos `.zip` (Windows) e `.tar.gz` (Linux/macOS) exige código adicional no Go. 
> Por favor, aprove este plano para prosseguirmos com a implementação.

## Open Questions

> [!IMPORTANT]
> 1. **Diretório de Instalação:** O plano atual sugere instalar o Java no diretório do usuário (ex: `~/.runner/jre`). Você prefere que ele seja instalado nesta pasta global do usuário ou na mesma pasta onde o executável `assinatura.exe` está localizado?
> 2. **Versão do Java:** O projeto requer Java 21+. Usaremos a API do Eclipse Temurin (Adoptium) para baixar o JRE 21 LTS mais recente. Você está de acordo?
> 3. **Feedback Visual:** Deseja que um progresso de download (como uma barra de progresso) seja exibido no terminal enquanto o JRE é baixado?

## Proposed Changes

### `projetos/assinatura/cmd`

Adicionaremos um novo pacote ou arquivo responsável pelo gerenciamento do Java e atualizaremos os comandos existentes para usá-lo.

#### [NEW] `projetos/assinatura/cmd/java_manager.go`
Criaremos um arquivo contendo a lógica para:
1. **Detectar o Java:** Tentar rodar `java -version` usando o executável do sistema (via variável `PATH`).
2. **Definir o Diretório Local:** Determinar o caminho local de instalação (ex: `~/.runner/jre`).
3. **Fazer o Download:** Caso o Java não seja encontrado nem no sistema nem na pasta local, fazer uma requisição HTTP para a API do Eclipse Temurin:
   `https://api.adoptium.net/v3/binary/latest/21/ga/{os}/{arch}/jre/hotspot/normal/eclipse?project=jdk`
4. **Extrair:** Descompactar o arquivo `.zip` (Windows) ou `.tar.gz` (Linux/macOS) no diretório local.
5. **Retornar o Caminho:** Retornar o caminho absoluto do executável `java` (seja o do sistema ou o baixado localmente).

#### [MODIFY] `projetos/assinatura/cmd/sign.go`
Alterar a chamada de execução do Java.
**De:** `exec.Command("java", "-jar", jarPath, ...)`
**Para:** 
```go
javaExecutable := obterCaminhoJava() // Chama a função do java_manager.go
exec.Command(javaExecutable, "-jar", jarPath, ...)
```

#### [MODIFY] `projetos/assinatura/cmd/validate.go`
Fazer a mesma alteração descrita acima para garantir que o comando `validate` também utilize o Java local se necessário.

## Verification Plan

### Automated Tests
1. **Simular ausência do Java:** Alteraremos temporariamente a variável `PATH` na execução para "esconder" o Java do sistema.
2. **Testar Download:** Rodaremos o CLI (`go run . sign --content "teste"`) e verificaremos se o download do JRE acontece.
3. **Testar Execução:** Verificaremos se, após o download, o comando de assinatura/validação é concluído com sucesso usando o Java recém-baixado.

### Manual Verification
- Compilar o binário Windows `.exe`.
- Executar em uma máquina virtual ou ambiente Windows sem o Java previamente instalado.
- Verificar a criação da pasta e a execução correta da assinatura.
