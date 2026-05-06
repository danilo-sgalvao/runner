# 05/05/25 -- LUIZ AUGUSTO

Hoje foi um dia de entendimento importante sobre o funcionamento das Sprints 1 e 2 do projeto.

Durante a análise do código e da estrutura do sistema, ficou mais claro como o projeto está organizado e como as partes se conectam entre si. O sistema é um CLI desenvolvido em Go, que utiliza a biblioteca Cobra para gerenciar comandos de terminal. A partir do arquivo `main.go`, o programa inicia a execução chamando o `cmd.Execute()`, que por sua vez direciona para o comando base (`rootCmd`) e seus subcomandos, como `sign` e `version`.

Na Sprint 1, o foco principal estava na estrutura inicial do projeto, incluindo a criação do CLI e a organização dos comandos básicos. Foi possível entender como o programa é iniciado e como a arquitetura do Cobra permite separar responsabilidades entre comandos diferentes.

Já na Sprint 2, o entendimento avançou para a parte funcional do sistema, especialmente o comando `sign`. Esse comando realiza a validação de parâmetros, procura o arquivo `assinador.jar` e executa um processo externo em Java para realizar a assinatura digital. Também foi compreendido como o sistema trata diferentes ambientes, verificando caminhos alternativos para localizar o JAR e lidando com possíveis erros de execução.

Além disso, foi possível perceber a integração do projeto com automações no GitHub Actions, onde o sistema é compilado para diferentes sistemas operacionais (Windows, Linux e macOS) e posteriormente publicado em releases com versionamento, checksums e assinaturas digitais.

De forma geral, o entendimento dessas duas sprints ajudou a visualizar melhor o fluxo completo do sistema: desde a execução do comando no terminal até a geração e distribuição dos binários. Isso tornou mais claro como cada parte do projeto contribui para o funcionamento final da aplicação e como a arquitetura foi pensada para suportar automação, distribuição e segurança.
