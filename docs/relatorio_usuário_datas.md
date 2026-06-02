# 05/05/25 -- LUIZ AUGUSTO

Hoje foi um dia de entendimento importante sobre o funcionamento das Sprints 1 e 2 do projeto.

Durante a análise do código e da estrutura do sistema, ficou mais claro como o projeto está organizado e como as partes se conectam entre si. O sistema é um CLI desenvolvido em Go, que utiliza a biblioteca Cobra para gerenciar comandos de terminal. A partir do arquivo `main.go`, o programa inicia a execução chamando o `cmd.Execute()`, que por sua vez direciona para o comando base (`rootCmd`) e seus subcomandos, como `sign` e `version`.

Na Sprint 1, o foco principal estava na estrutura inicial do projeto, incluindo a criação do CLI e a organização dos comandos básicos. Foi possível entender como o programa é iniciado e como a arquitetura do Cobra permite separar responsabilidades entre comandos diferentes.

Já na Sprint 2, o entendimento avançou para a parte funcional do sistema, especialmente o comando `sign`. Esse comando realiza a validação de parâmetros, procura o arquivo `assinador.jar` e executa um processo externo em Java para realizar a assinatura digital. Também foi compreendido como o sistema trata diferentes ambientes, verificando caminhos alternativos para localizar o JAR e lidando com possíveis erros de execução.

Além disso, foi possível perceber a integração do projeto com automações no GitHub Actions, onde o sistema é compilado para diferentes sistemas operacionais (Windows, Linux e macOS) e posteriormente publicado em releases com versionamento, checksums e assinaturas digitais.

De forma geral, o entendimento dessas duas sprints ajudou a visualizar melhor o fluxo completo do sistema: desde a execução do comando no terminal até a geração e distribuição dos binários. Isso tornou mais claro como cada parte do projeto contribui para o funcionamento final da aplicação e como a arquitetura foi pensada para suportar automação, distribuição e segurança.

# 05/05/25 -- Danilo Galvão

Consolidação do plano para implementação da capacidade de verificar e, se necessário, baixar e instalar localmente o Java na máquina do usuário. Validação com o professor: uma abordagem melhor seria tornar a escolha da versão flexível e externa ao sistema. Plano disponível em: [plano-download-java.md](plano-download-java.md).

# 12/05/26 -- Danilo Galvão

Implementação do provisionamento automático de JRE para a Sprint 2 do Sistema Runner.

**O que foi feito:**

- **Provisionamento automático do JRE (`internal/jre/manager.go`)**: implementação do fluxo definido em `docs/plano-download-java.md`. Detecta Java gerenciado localmente em `~/.hubsaude/jre`, depois verifica o `java` no PATH do sistema (exige versão 21+), e se necessário baixa e extrai o JRE via URLs definidas em `release.json`. Suporta extração de `.zip` (Windows) e `.tar.gz` (Linux/macOS).

- **`release.json`**: criado na raiz do repositório com as URLs do JRE por plataforma (Eclipse Temurin 21), permitindo atualizar a versão do Java sem recompilar os binários.

- **Atualização do `sign.go` e `validate.go`**: migração para `RunE` (retorna erro em vez de chamar `os.Exit()`), uso de `jre.JavaPath()` no lugar de `"java"` hardcoded, e `MarkFlagRequired` para validação automática de flags obrigatórias pelo Cobra.

# 12/05/26 -- LUIZ AUGUSTO

Implementação e testes da Sprint 2 do Sistema Runner.

A sprint foi focada em entregar o fluxo ponta-a-ponta de assinatura e validação digital simulada, com qualidade de código e cobertura de testes.

**O que foi feito:**

- **Refatoração do AssinadorService.java**: a classe foi reestruturada para separar a lógica de negócio (validação e simulação) do I/O (impressão no terminal e System.exit). Agora lança `IllegalArgumentException` para parâmetros inválidos, tornando o código testável com JUnit sem necessidade de interceptar a JVM.

- **Atualização do Main.java**: passou a ser responsável pelo parse de argumentos, formatação da saída (`status=sucesso`, `assinatura=...`, etc.) e pelos códigos de saída. Captura a exceção lançada pelo serviço e exibe mensagem de erro limpa ao usuário.

- **Testes JUnit 5 (17 testes)**: cobertura completa dos cenários de sucesso, falha e validação de parâmetros de `sign` e `validate`, incluindo um teste de fluxo completo (sign → validate). Todos passam com `mvn test`.

- **Correção do root.go**: nome do CLI corrigido de "runner" para "assinatura", com descrições e exemplos de uso adequados.

- **Criação do jar.go**: função `encontrarJar()` extraída para arquivo próprio, retornando `(string, error)` em vez de chamar `os.Exit()`, permitindo tratamento de erro limpo e testabilidade.

- **Testes Go (17 testes + 8 testes do jre)**: validação de registro de comandos, presença e configuração correta de todas as flags, e lógica de seleção de URL do JRE por plataforma. Todos passam.

**Resultado dos testes:**
- Java: 17/17 ✅
- Go cmd: 17/17 ✅
- Go jre: 8 pass / 2 skip (skips de plataforma: Linux e macOS pulados corretamente no Windows) ✅

# 19/05/26 -- LUIZ AUGUSTO

Implementação da identificação genérica do `assinador.jar` independente da máquina.

Anteriormente, a função `encontrarJar()` localizava o arquivo `.jar` em apenas dois locais fixos: ao lado do executável (modo produção) ou em `../assinador-java/target/` (modo desenvolvimento local). Isso causava falha em qualquer máquina que não tivesse o projeto clonado localmente ou o binário distribuído com o jar ao lado.

**O que foi feito:**

- **Atualização do `jar.go`**: a função `encontrarJar()` foi expandida com uma nova ordem de busca:
  1. Ao lado do executável (modo distribuído/produção)
  2. `~/.hubsaude/assinador.jar` (cache local gerenciado automaticamente)
  3. `../assinador-java/target/assinador.jar` (modo desenvolvimento)
  4. Download automático via `release.json` do repositório, salvando em `~/.hubsaude/assinador.jar`

- **Atualização do `release.json`**: adicionada a chave `"jar"` com a URL do artefato publicado no GitHub Releases (`releases/latest/download/assinador.jar`), permitindo atualizar o jar sem recompilar os binários Go.

**Resultado:** qualquer máquina que execute o CLI pela primeira vez sem o jar presente irá baixá-lo automaticamente da internet e armazená-lo em cache local, sem necessidade de intervenção manual. O padrão segue a mesma arquitetura já usada pelo gerenciador do JRE (`internal/jre/manager.go`).

# 19/05/26 -- Danilo Galvão

Refatoração do `assinador-java` com introdução da interface `SignatureService`.

**O que foi feito:**

- **Criação da interface `SignatureService`**: definição dos contratos `sign(SignRequest)` e `validate(ValidateRequest)`, desacoplando a lógica de negócio da implementação concreta.

- **Implementação `FakeSignatureService`**: classe que implementa `SignatureService` com assinaturas simuladas (retorna `MOCKED_SIGNATURE_BASE64_==`), isolando o comportamento fake atrás da interface.

- **Atualização do `Main.java`**: passou a depender da interface `SignatureService` em vez da implementação direta, tornando o sistema preparado para substituição futura por uma implementação real de criptografia sem alteração no ponto de entrada.

# 26/05/26 -- Danilo Galvão

Refatoração completa da arquitetura do `assinador-java` (Fases 1 e 2 do plano de refatoração — Sprint 2).

O objetivo era reorganizar a estrutura plana existente em camadas bem definidas antes da Sprint 3, que exige dois adaptadores de entrada (CLI e HTTP) compartilhando o mesmo núcleo de negócio. A refatoração foi executada em cinco checkpoints atômicos, cada um terminando com build verde e `mvn test` passando.

**O que foi feito:**

- **CP1 — Jackson + fat-jar**: substituição do `toJson`/`escapeJson` manual do `Main` por `infrastructure/json/JsonMapper` usando Jackson. Adicionada dependência `jackson-databind` ao `pom.xml` e configurado o `maven-shade-plugin` para gerar fat-jar self-contained (`assinador.jar` com todas as dependências embutidas), necessário porque o CLI Go distribui um único arquivo.

- **CP2 — Reorganização do domínio**: criação dos pacotes `domain/model/` e `domain/service/`. DTOs movidos para `domain/model/`; `SignatureService` e `FakeSignatureService` movidos para `domain/service/`. `SignatureResponse` renomeada para `SignatureResult` (resultado de domínio, não DTO de transporte).

- **CP3 — Camada `application`**: extração da validação de parâmetros do `FakeSignatureService` para `application/validation/RequestValidator` (fonte única das regras) e `ValidationException`. Criação de `SignUseCase` e `ValidateUseCase`, que validam via `RequestValidator` e delegam ao `SignatureService`. `FakeSignatureService` passa a assumir entrada já validada. Entrypoint religado nos use cases.

- **CP4 — Camada `presentation/cli`**: extração do parsing de argumentos para `CliRunner` e da formatação de saída (JSON, stdout/stderr, exit codes) para `CliPresenter`. `Main.java` renomeado para `AssinadorApplication` — composition root enxuto. `<mainClass>` atualizado nos dois plugins do `pom.xml`.

- **CP5 — Dispatcher dual-mode**: `AssinadorApplication` passou a detectar `args[0] == "serve"` e desviar para mensagem `"Modo servidor (serve) ainda não implementado."` + exit 1, estabelecendo o ponto de extensão da Sprint 3 sem ativar Spring.

- **Documentação**: `CLAUDE.md` e `README.md` atualizados para refletir a nova estrutura de pacotes; `plano-refatoracao-arquitetura-java.md` atualizado com status das fases; `.gitignore` ajustado para ignorar `dependency-reduced-pom.xml` gerado pelo shade.

**Resultado dos testes:**
- Java: 22/22 ✅ (`FakeSignatureServiceTest` 3 + `RequestValidatorTest` 10 + `UseCasesTest` 5 + `JsonMapperTest` 3 + `SignatureControllerTest` 1 — contrato externo preservado em todos os cenários)
- Contrato CLI Go (JSON, streams, exit codes, mensagens literais): verificado manualmente nos 4 cenários baseline ✅

# 26/05/26 -- LUIZ AUGUSTO

Implementação da Fase 3 do plano de refatoração da arquitetura Java: modo servidor HTTP (US-02.4).

A fase adiciona ao `assinador.jar` a capacidade de funcionar como servidor HTTP permanente, expondo os endpoints `/sign` e `/validate`. O mesmo núcleo de negócio — use cases, validação e serviço fake — é reutilizado pelos dois adaptadores de entrada (CLI e HTTP), sem duplicação.

**O que foi feito:**

- **Atualização do `pom.xml`**: substituição do `maven-shade-plugin` pelo `spring-boot-maven-plugin` com goal `repackage`. Adicionadas as dependências `spring-boot-starter-web` e `spring-boot-starter-test` via BOM do Spring Boot 3.3.5. O fat-jar final (`assinador.jar`, ~20 MB) continua sendo um executável self-contained com `Main-Class` apontando para o launcher do Spring Boot e `Start-Class` para `AssinadorApplication`.

- **`WebApplication.java`**: classe anotada com `@SpringBootApplication`, usada como raiz do contexto Spring. Ativada apenas quando `AssinadorApplication` entra no branch `serve`; o modo CLI não sobe o Spring.

- **`infrastructure/config/AppConfig.java`**: `@Configuration` que declara como beans Spring os mesmos objetos do núcleo instanciados manualmente no modo CLI (`FakeSignatureService`, `RequestValidator`, `SignUseCase`, `ValidateUseCase`).

- **`presentation/http/SignatureController.java`**: `@RestController` com `POST /sign` e `POST /validate`. Recebe DTOs HTTP, converte para os modelos de domínio, delega aos use cases existentes e retorna `SignatureHttpResponse`.

- **`presentation/http/GlobalExceptionHandler.java`**: `@RestControllerAdvice` que trata `ValidationException` com HTTP 400 e exceções genéricas com HTTP 500, sempre retornando a mesma estrutura de resposta (`signature`, `valid`, `message`).

- **DTOs HTTP** (`presentation/http/dto/`): `SignHttpRequest`, `ValidateHttpRequest` e `SignatureHttpResponse` — tipos de transporte independentes do domínio.

- **`infrastructure/ServerStartupHandler.java`**: `ApplicationListener<WebServerInitializedEvent>` que, ao servidor iniciar, escreve `{"pid":N,"port":N}` em `~/.hubsaude/assinador.pid` para uso futuro pelo CLI Go (US-01.7, US-01.8). Exibe no stderr a confirmação da porta e PID.

- **Atualização do `AssinadorApplication.java`**: o branch `serve` agora parseia o argumento `--port N` (padrão 8080), configura a propriedade `server.port` via `SpringApplication.setDefaultProperties` e chama `SpringApplication.run(WebApplication.class, ...)`. Qualquer outro comando segue o fluxo CLI inalterado.

- **`SignatureControllerTest.java`**: 7 testes de integração com `@SpringBootTest` + `MockMvc`, cobrindo sign com conteúdo válido (200), sign com conteúdo vazio (400), validate com assinatura correta (200 `valid=true`), validate com assinatura errada (200 `valid=false`), validate sem content (400) e validate sem signature (400).

**Resultado dos testes:**
- Java (total): 29/29 ✅ (22 anteriores + 7 novos de integração HTTP)

# 02/06/26 -- LUIZ AUGUSTO

Implementação completa da Sprint 3 do Sistema Runner: modo servidor HTTP, gerenciamento de ciclo de vida via CLI e integração com dispositivo criptográfico via PKCS#11.

**O que foi feito:**

- **`GET /health` (`HealthController.java`)**: endpoint adicionado ao servidor Spring Boot. Retorna `{"status":"UP"}` e é usado pelo CLI Go para aguardar a inicialização do servidor antes de confirmar ao usuário.

- **Integração PKCS#11 (US-02.5)**:
  - `infrastructure/pkcs11/Pkcs11Config.java`: lê as variáveis de ambiente `HUBSAUDE_PKCS11_LIBRARY`, `HUBSAUDE_PKCS11_NAME` e `HUBSAUDE_PKCS11_PIN`. Retorna `null` quando a biblioteca não está configurada, sinalizando uso do serviço simulado.
  - `infrastructure/pkcs11/Pkcs11ServiceFactory.java`: configura o provider `SunPKCS11` com a biblioteca nativa indicada e abre o `KeyStore` do dispositivo. Inclui documentação do setup com SoftHSM2.
  - `domain/service/Pkcs11SignatureService.java`: implementa `SignatureService` usando chave privada PKCS#11 para assinar (`SHA256withRSA`) e tentando todos os certificados do dispositivo para validar.
  - `infrastructure/config/AppConfig.java`: atualizado para usar PKCS#11 quando configurado, com fallback automático para `FakeSignatureService` e log de aviso ao usuário.

- **Shutdown por inatividade (US-01.9)**:
  - `infrastructure/RequestTimestamp.java`: componente Spring com `AtomicLong` que registra o instante da última requisição HTTP recebida.
  - `infrastructure/InactivityFilter.java`: filtro `OncePerRequestFilter` que chama `touch()` a cada requisição, mantendo o timestamp atualizado.
  - `infrastructure/InactivityShutdown.java`: lê a variável de ambiente `HUBSAUDE_TIMEOUT_MINUTES` e inicia uma thread daemon que chama `System.exit(0)` caso nenhuma requisição seja recebida dentro do período configurado.

- **Testes Java**:
  - `Pkcs11SignatureServiceTest.java` (5 testes): cobre chave não encontrada, exceção no KeyStore, KeyStore vazio, encoding Base64 inválido e erro geral — todos com KeyStore mockado via Mockito, sem necessidade de hardware.
  - `SignatureServerSmokeTest.java`: adicionado `getHealth_retorna200ComStatusUp()` ao smoke test existente.

- **Pacote `internal/process` (Go)**: `detach_unix.go` (build tag `!windows`) usa `Setsid: true`; `detach_windows.go` usa `CREATE_NEW_PROCESS_GROUP`. Garante que o servidor Java sobrevive ao encerramento do CLI em todas as plataformas.

- **Pacote `internal/server` (Go)**:
  - `manager.go`: lê, limpa e verifica o arquivo `~/.hubsaude/assinador.pid` gravado pelo `ServerStartupHandler`. `IsResponding` faz health check HTTP com timeout de 2 segundos. `PidFilePath` é exportado para sobrescrita em testes.
  - `client.go`: `Sign()` e `Validate()` fazem `POST /sign` e `POST /validate` no servidor ativo, desserializando a resposta em `SignatureResponse`.

- **`cmd/start.go` (US-01.5)**: comando `assinatura start [--port 8080] [--timeout 0]`. Verifica se já há instância ativa na mesma porta antes de iniciar. Inicia o JAR com `java -jar ... serve --port N` em background (processo detachado), passa `HUBSAUDE_TIMEOUT_MINUTES` via ambiente e aguarda o `/health` responder por até 30 segundos.

- **`cmd/stop.go` (US-01.8)**: comando `assinatura stop [--port 0]`. Lê o PID de `~/.hubsaude/assinador.pid`, verifica se o processo responde, encerra via `os.FindProcess` + `Kill()` e limpa o arquivo de registro.

- **`cmd/sign.go` e `cmd/validate.go` (US-01.6 + US-01.7)**: ambos verificam `~/.hubsaude/assinador.pid` e fazem health check antes de cada operação. Se o servidor estiver ativo, usam HTTP (menor latência). Caso contrário, caem automaticamente para `java -jar`. Flag `--local` força o modo local independentemente do estado do servidor.

- **Testes Go**:
  - `internal/server/manager_test.go` (4 testes): leitura de arquivo válido, arquivo inexistente, JSON corrompido e remoção via `ClearProcessInfo`.
  - `internal/server/client_test.go` (3 testes): sign com servidor fake (`httptest`), validate com servidor fake e erro quando servidor indisponível.
  - `cmd/start_test.go` (6 testes) e `cmd/stop_test.go` (5 testes): registro dos comandos no root, presença e valores padrão das flags.

**Resultado dos testes:**
- Java: 34/34 ✅ (29 anteriores + 5 novos de PKCS#11)
- Go cmd: 28/28 ✅ (17 anteriores + 6 start + 5 stop)
- Go server: 7/7 ✅ (4 manager + 3 client)
