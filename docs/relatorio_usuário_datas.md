# 05/05/25 -- LUIZ AUGUSTO

Hoje foi um dia de entendimento importante sobre o funcionamento das Sprints 1 e 2 do projeto.

Durante a análise do código e da estrutura do sistema, ficou mais claro como o projeto está organizado e como as partes se conectam entre si. O sistema é um CLI desenvolvido em Go, que utiliza a biblioteca Cobra para gerenciar comandos de terminal. A partir do arquivo `main.go`, o programa inicia a execução chamando o `cmd.Execute()`, que por sua vez direciona para o comando base (`rootCmd`) e seus subcomandos, como `sign` e `version`.

Na Sprint 1, o foco principal estava na estrutura inicial do projeto, incluindo a criação do CLI e a organização dos comandos básicos. Foi possível entender como o programa é iniciado e como a arquitetura do Cobra permite separar responsabilidades entre comandos diferentes.

Já na Sprint 2, o entendimento avançou para a parte funcional do sistema, especialmente o comando `sign`. Esse comando realiza a validação de parâmetros, procura o arquivo `assinador.jar` e executa um processo externo em Java para realizar a assinatura digital. Também foi compreendido como o sistema trata diferentes ambientes, verificando caminhos alternativos para localizar o JAR e lidando com possíveis erros de execução.

Além disso, foi possível perceber a integração do projeto com automações no GitHub Actions, onde o sistema é compilado para diferentes sistemas operacionais (Windows, Linux e macOS) e posteriormente publicado em releases com versionamento, checksums e assinaturas digitais.

De forma geral, o entendimento dessas duas sprints ajudou a visualizar melhor o fluxo completo do sistema: desde a execução do comando no terminal até a geração e distribuição dos binários. Isso tornou mais claro como cada parte do projeto contribui para o funcionamento final da aplicação e como a arquitetura foi pensada para suportar automação, distribuição e segurança.

# 12/05/26 -- LUIZ AUGUSTO

Implementação completa da Sprint 2 do Sistema Runner.

A sprint foi focada em entregar o fluxo ponta-a-ponta de assinatura e validação digital simulada, com qualidade de código e cobertura de testes.

**O que foi feito:**

- **Refatoração do AssinadorService.java**: a classe foi reestruturada para separar a lógica de negócio (validação e simulação) do I/O (impressão no terminal e System.exit). Agora lança `IllegalArgumentException` para parâmetros inválidos, tornando o código testável com JUnit sem necessidade de interceptar a JVM.

- **Atualização do Main.java**: passou a ser responsável pelo parse de argumentos, formatação da saída (`status=sucesso`, `assinatura=...`, etc.) e pelos códigos de saída. Captura a exceção lançada pelo serviço e exibe mensagem de erro limpa ao usuário.

- **Testes JUnit 5 (17 testes)**: cobertura completa dos cenários de sucesso, falha e validação de parâmetros de `sign` e `validate`, incluindo um teste de fluxo completo (sign → validate). Todos passam com `mvn test`.

- **Correção do root.go**: nome do CLI corrigido de "runner" para "assinatura", com descrições e exemplos de uso adequados.

- **Criação do jar.go**: função `encontrarJar()` extraída para arquivo próprio, retornando `(string, error)` em vez de chamar `os.Exit()`, permitindo tratamento de erro limpo e testabilidade.

- **Provisionamento automático do JRE (`internal/jre/manager.go`)**: implementação do fluxo definido em `docs/plano-download-java.md`. Detecta Java local (`~/.hubsaude/jre`), depois sistema (PATH), e se necessário baixa o JRE via `release.json`. Suporta extração de `.zip` (Windows) e `.tar.gz` (Linux/macOS).

- **Atualização do sign.go e validate.go**: migração para `RunE` (retorna erro em vez de chamar `os.Exit()`), uso de `jre.JavaPath()` no lugar de `"java"` hardcoded, e `MarkFlagRequired` para validação automática de flags obrigatórias pelo Cobra.

- **Testes Go (17 testes + 8 testes do jre)**: validação de registro de comandos, presença e configuração correta de todas as flags, e lógica de seleção de URL do JRE por plataforma. Todos passam.

- **release.json**: criado na raiz do repositório com as URLs do JRE por plataforma (Eclipse Temurin 21), permitindo atualizar a versão do Java sem recompilar os binários.

- **Guia técnico (`docs/guia-tecnico.md`)**: documento explicando o problema resolvido, a estrutura do projeto, as ferramentas utilizadas (Go, Cobra, Java, Maven, JUnit 5, GitHub Actions, Cosign), o fluxo lógico de cada componente e como compilar/testar localmente.

**Resultado dos testes:**
- Java: 17/17 ✅
- Go cmd: 17/17 ✅
- Go jre: 8 pass / 2 skip (skips de plataforma: Linux e macOS pulados corretamente no Windows) ✅

# 05/05/25 -- Danilo Galvão

Consolidação do plano para implementação da capacidade de verificar e, se necessário, baixar e instalar localmente o Java na máquina do usuário. Validação com o professor: uma abordagem melhor seria tornar a escolha da versão flexível e externa ao sistema. Plano disponível em: [plano-download-java.md](plano-download-java.md).
