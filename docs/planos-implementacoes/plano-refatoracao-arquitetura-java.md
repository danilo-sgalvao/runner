# Plano: refatoração da arquitetura do assinador.jar (preparação para a Sprint 3)

## Objetivo

Reorganizar o `projetos/assinador-java` em camadas bem definidas (domain → application → presentation → infrastructure) **antes** de iniciar a Sprint 3. A Sprint 3 introduz o modo servidor HTTP (US-02.4) e o material criptográfico via PKCS#11 (US-02.5); ambos precisam **reaproveitar** a mesma lógica de validação e simulação já usada no modo CLI. A arquitetura atual, plana, não comporta isso sem duplicação.

A refatoração estrutural (Fase 1) **não muda o comportamento externo**: o contrato de saída do modo CLI (JSON em stdout, mensagens em stderr, exit codes 0/1) permanece idêntico, pois é consumido pelo CLI Go `assinatura`.

## Estado atual

```
com.hubsaude.assinador
├── Main.java                 # parsing de args + roteamento + serialização JSON + I/O + exit codes
├── SignatureService.java     # interface (porta)
├── FakeSignatureService.java # implementação simulada + validação de parâmetros (misturadas)
└── domain/
    ├── SignRequest.java       # DTO (content, token)
    ├── ValidateRequest.java   # DTO (content, signature)
    └── SignatureResponse.java # DTO (signature, valid, message)
```

### Problemas identificados

1. **`Main` faz coisas demais.** Concentra parsing de argumentos, roteamento de comandos, serialização JSON (`toJson`/`escapeJson` à mão), escrita em stdout/stderr e controle de exit code. São cinco responsabilidades em uma classe.
2. **Serialização JSON na camada errada e frágil.** O `toJson` manual com `escapeJson` é fácil de quebrar (não trata `\n`, `\t`, unicode, etc.) e está dentro do entrypoint. Deve ser delegado a uma biblioteca (Jackson) na infraestrutura.
3. **Validação acoplada à implementação "fake".** A validação de parâmetros (presença/formato) vive dentro de `FakeSignatureService`. Quando entrar uma implementação real (PKCS#11, Sprint 3), a validação teria de ser reescrita ou duplicada. Validação é regra de fronteira, independente de a assinatura ser fake ou real.
4. **Só existe a camada `domain` — e ela só contém DTOs.** Não há separação entre o núcleo de negócio, a orquestração de casos de uso e os adaptadores de entrada/saída. Não há onde encaixar um `SignatureController` (Spring) sem misturá-lo com o núcleo.
5. **Um único ponto de entrada (CLI).** A Sprint 3 exige dois adaptadores de entrada (CLI + HTTP) servindo o mesmo núcleo.

## Arquitetura-alvo

Organização **por camadas** (em linha com o pedido: presentation / controllers / application / domain). O núcleo (`domain` + `application`) é **livre de framework** — sem anotações Spring, sem I/O, sem JSON. Os detalhes técnicos (Spring, Jackson, PKCS#11) ficam nas bordas (`presentation` e `infrastructure`).

```
com.hubsaude.assinador
│
├── AssinadorApplication.java         # dispatcher: decide modo CLI vs. servidor a partir de args[0]
│
├── domain/                           # NÚCLEO — regras puras, sem framework/IO/JSON
│   ├── model/
│   │   ├── SignRequest.java
│   │   ├── ValidateRequest.java
│   │   └── SignatureResult.java      # renomeado de SignatureResponse (é resultado de domínio, não DTO de transporte)
│   └── service/
│       ├── SignatureService.java     # porta (interface)
│       └── FakeSignatureService.java # impl. simulada — assume entrada já validada
│
├── application/                      # CASOS DE USO + validação de entrada
│   ├── SignUseCase.java              # valida → delega ao SignatureService
│   ├── ValidateUseCase.java          # valida → delega ao SignatureService
│   └── validation/
│       ├── RequestValidator.java     # única fonte de verdade da validação de parâmetros
│       └── ValidationException.java
│
├── presentation/                     # ADAPTADORES DE ENTRADA
│   ├── cli/
│   │   ├── CliRunner.java            # parsing de argumentos → caso de uso
│   │   └── CliPresenter.java         # formata SignatureResult p/ stdout/stderr + exit code
│   └── http/                         # SPRINT 3 (US-02.4)
│       ├── SignatureController.java  # @RestController: POST /sign, POST /validate
│       ├── GlobalExceptionHandler.java
│       └── dto/
│           ├── SignHttpRequest.java
│           ├── ValidateHttpRequest.java
│           └── SignatureHttpResponse.java
│
└── infrastructure/                   # ADAPTADORES DE SAÍDA / detalhes técnicos
    ├── json/
    │   └── JsonMapper.java           # Jackson — substitui o toJson/escapeJson manual
    └── crypto/                       # SPRINT 3 (US-02.5)
        └── Pkcs11SignatureService.java  # impl. alternativa do SignatureService via SunPKCS11
```

### Regra de dependência

As dependências apontam **para dentro**: `presentation` e `infrastructure` dependem de `application`, que depende de `domain`. O `domain` não depende de ninguém. Resultado: o mesmo `SignUseCase` é chamado tanto pelo `CliRunner` quanto pelo `SignatureController`, sem duplicar validação nem simulação.

### Dois "composition roots", um único núcleo

O ponto-chave para conciliar a distinção *cold start* (CLI) × *warm start* (servidor) da especificação:

- **Modo CLI** (`java -jar assinador.jar sign --content ...`): `AssinadorApplication` instancia manualmente o núcleo (sem subir Spring) → execução leve, adequada a scripts e uso esporádico.
- **Modo servidor** (`java -jar assinador.jar serve [--port N]`): `AssinadorApplication` chama `SpringApplication.run(...)`; uma `@Configuration` declara os mesmos objetos do núcleo como beans → menor latência nas chamadas subsequentes.

Assim o overhead do Spring Boot só é pago quando o modo servidor é realmente usado.

## Fases de implementação

### Fase 1 — Refatoração estrutural (Sprint 2, sem mudar comportamento externo)

Reorganiza o código e introduz Jackson, mantendo a saída do CLI idêntica e todos os testes verdes.

1. Adicionar dependência Jackson (`com.fasterxml.jackson.core:jackson-databind`) ao `pom.xml`.
2. Criar os pacotes `domain.model`, `domain.service`, `application`, `application.validation`, `presentation.cli`, `infrastructure.json`.
3. Mover os DTOs para `domain/model`; renomear `SignatureResponse` → `SignatureResult`.
4. Mover `SignatureService` + `FakeSignatureService` para `domain/service`; **remover a validação** de dentro do `FakeSignatureService` (passa a assumir entrada válida).
5. Criar `application/validation/RequestValidator` + `ValidationException` com a lógica de validação extraída.
6. Criar `application/SignUseCase` e `ValidateUseCase` (validam via `RequestValidator`, depois chamam o `SignatureService`).
7. Criar `presentation/cli/CliRunner` (parsing de args, hoje no `Main`) e `CliPresenter` (formatação de saída + exit codes, hoje no `Main`).
8. Criar `infrastructure/json/JsonMapper` (Jackson) — remove `toJson`/`escapeJson` manuais.
9. Reduzir `Main`/`AssinadorApplication` a um *composition root* enxuto do modo CLI.
10. Migrar os testes: `FakeSignatureServiceTest` foca no núcleo; novos testes de `RequestValidator` (cenários de validação) e de `JsonMapper`; remover os testes de `Main.toJson` (a serialização agora é do Jackson).

**Critério de pronto da Fase 1:** `mvn test` verde, e a saída de `sign`/`validate` byte-a-byte equivalente à atual (validar manualmente o contrato JSON consumido pelo CLI Go).

### Fase 2 — Dispatcher dual-mode (ponte para a Sprint 3, ainda na Sprint 2)

11. Introduzir o despacho por modo em `AssinadorApplication`: `args[0] == "serve"` reservado para o modo servidor (ainda sem implementação — pode retornar "não implementado"); qualquer outro comando segue pelo `CliRunner`. Estabelece o ponto de extensão sem ativar Spring ainda.

### Fase 3 — Sprint 3: modo servidor HTTP (US-02.4 e suporte a US-01.5–01.9)

12. Adicionar `spring-boot-starter-web` (e o parent/BOM do Spring Boot) ao `pom.xml`; ajustar o `spring-boot-maven-plugin` para repackage mantendo `AssinadorApplication` como classe principal.
13. Criar `presentation/http/SignatureController` com `POST /sign` e `POST /validate`, reusando **os mesmos** `SignUseCase`/`ValidateUseCase`.
14. Criar os DTOs HTTP + `GlobalExceptionHandler` para estrutura de erro consistente (sucesso e falha).
15. Criar a `@Configuration` que declara o núcleo como beans.
16. Ativar o modo `serve` (porta padrão + flag `--port`); registrar PID/porta em `~/.hubsaude/` conforme US-01.5.
17. Testes de integração dos endpoints (`@SpringBootTest`/MockMvc).

### Fase 4 — Sprint 3: PKCS#11 (US-02.5)

18. Criar `infrastructure/crypto/Pkcs11SignatureService` (implementação alternativa de `SignatureService` via `SunPKCS11`); o campo `token` já existente em `SignRequest` passa a ser usado.
19. Selecionar a implementação (fake vs. PKCS#11) por profile/flag; mensagem clara quando o dispositivo não está disponível.
20. Testes de integração com SoftHSM2 (ou simulador equivalente).

## Impacto fora do Java

- **CLI Go (`assinatura`)**: não muda, **desde que** o contrato do modo CLI (argumentos, JSON em stdout, exit codes) seja preservado na Fase 1. Na Sprint 3, o CLI Go ganhará o caminho HTTP (US-01.6) apontando para os endpoints da Fase 3.
- **Documentação**: ao concluir a Fase 1, atualizar `CLAUDE.md` (seções "Java service architecture" e "Key Files") e o `README.md` para refletir a nova estrutura de pacotes. Atualizar também o diagrama C4 de contêiner (o assinador.jar passa a ter dois modos de execução).

## Riscos e cuidados

- **Quebra do contrato com o CLI Go.** A troca do `toJson` manual pelo Jackson pode alterar a ordem/formatação das chaves do JSON. Garantir saída equivalente (ou ajustar o parsing do lado Go conjuntamente). Bloqueante na Fase 1.
- **Cold start do Spring no modo servidor.** Aceitável porque só ocorre no `serve`; o modo CLI permanece sem Spring. Não regredir isso transformando todo o núcleo em beans Spring.
- **Repackage do fat-jar.** O `spring-boot-maven-plugin` usa seu próprio launcher; é preciso configurar a `start-class`/`mainClass` para o dispatcher `AssinadorApplication`, senão o modo CLI quebra. Validar na Fase 3.
- **Escopo da refatoração.** Fase 1 é refatoração pura (sem novas features); resistir à tentação de já implementar HTTP/PKCS#11 antes de o núcleo estar limpo e testado.

## Ordem sugerida

Fase 1 (estrutural) → Fase 2 (dispatcher) ainda dentro da Sprint 2; Fases 3 e 4 abrem a Sprint 3.

## Decisões em aberto

1. **Framework HTTP: Spring Boot** (recomendado, alinhado à nota original "controllers (spring)" e ao `SignatureController` da especificação). Alternativa mais leve: `com.sun.net.httpserver.HttpServer` (zero dependências) ou Javalin — descartam o cold start, mas fogem do que a especificação sugere e do objetivo didático.
2. **Validação na camada application** (`RequestValidator`) como fonte única, em vez de Bean Validation (`@NotBlank`) só nos DTOs HTTP — assim CLI e HTTP compartilham exatamente as mesmas regras. (Pode-se somar `@Valid` nos DTOs HTTP como primeira barreira, delegando a regra de negócio ao mesmo validador.)
3. **Renomear `SignatureResponse` → `SignatureResult`**: separa o resultado de domínio do DTO de transporte HTTP (`SignatureHttpResponse`). Confirmar se vale o churn nos testes agora ou manter o nome.
