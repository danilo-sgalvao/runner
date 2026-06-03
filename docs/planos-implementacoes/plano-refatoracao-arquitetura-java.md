# Plano: refatoração da arquitetura do assinador.jar (preparação para a Sprint 3)

## Objetivo

Reorganizar o `projetos/assinador-java` em camadas bem definidas (domain → application → presentation → infrastructure) **antes** de iniciar a Sprint 3. A Sprint 3 introduz o modo servidor HTTP (US-02.4) e o material criptográfico via PKCS#11 (US-02.5); ambos precisam **reaproveitar** a mesma lógica de validação e simulação já usada no modo CLI. A arquitetura atual, plana, não comporta isso sem duplicação.

A refatoração estrutural (Fase 1) **não muda o comportamento externo**: o contrato de saída do modo CLI (JSON em stdout, mensagens em stderr, exit codes 0/1) permanece idêntico, pois é consumido pelo CLI Go `assinatura`.

## Status

| Fase | Escopo | Status |
|------|--------|--------|
| Fase 1 — Refatoração estrutural | Sprint 2 | ✅ Concluída |
| Fase 2 — Dispatcher dual-mode | Sprint 2 | ✅ Concluída |
| Fase 3 — Modo servidor HTTP (US-02.4) | Sprint 3 | ✅ Concluída |
| Fase 4 — PKCS#11 (US-02.5) | Sprint 3 | ✅ Concluída |

Todas as quatro fases foram implementadas e commitadas. `CLAUDE.md` e `README.md` atualizados. O modo `serve` sobe um servidor HTTP Spring Boot (`POST /sign`, `POST /validate`) reusando o mesmo núcleo do modo CLI; `mvn test` 34/34 verde. A Fase 4 (PKCS#11) introduziu `Pkcs11Config`, `Pkcs11ServiceFactory` e `Pkcs11SignatureService` — seleção automática via `HUBSAUDE_PKCS11_LIBRARY`, com fallback para `FakeSignatureService`. Do lado do CLI Go, o caminho HTTP de gestão do servidor (US-01.5–01.9) também foi concluído na Sprint 3.

## Estado anterior à refatoração (início da Sprint 2)

Estrutura plana sem separação de responsabilidades — contexto para entender os problemas que motivaram a refatoração:

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

### Fase 1 — Refatoração estrutural ✅ Concluída (Sprint 2)

Reorganiza o código e introduz Jackson, mantendo a saída do CLI idêntica e todos os testes verdes.

1. ✅ Adicionar dependência Jackson (`com.fasterxml.jackson.core:jackson-databind:2.18.2`) ao `pom.xml`; adicionar `maven-shade-plugin` para gerar fat-jar self-contained.
2. ✅ Criar os pacotes `domain.model`, `domain.service`, `application`, `application.validation`, `presentation.cli`, `infrastructure.json`.
3. ✅ Mover os DTOs para `domain/model`; renomear `SignatureResponse` → `SignatureResult`.
4. ✅ Mover `SignatureService` + `FakeSignatureService` para `domain/service`; remover a validação de dentro do `FakeSignatureService` (passa a assumir entrada válida).
5. ✅ Criar `application/validation/RequestValidator` + `ValidationException` com a lógica de validação extraída.
6. ✅ Criar `application/SignUseCase` e `ValidateUseCase` (validam via `RequestValidator`, depois chamam o `SignatureService`).
7. ✅ Criar `presentation/cli/CliRunner` (parsing de args) e `CliPresenter` (formatação de saída + exit codes).
8. ✅ Criar `infrastructure/json/JsonMapper` (Jackson) — remove `toJson`/`escapeJson` manuais.
9. ✅ Reduzir `Main` → `AssinadorApplication`: composition root enxuto do modo CLI; `<mainClass>` atualizado nos dois plugins do `pom.xml`.
10. ✅ Migrar os testes: `FakeSignatureServiceTest` foca no núcleo (4 testes); `RequestValidatorTest` (10 testes); `JsonMapperTest` (3 testes); `UseCasesTest` (5 testes). Total: 22 testes verdes.

**Critério de pronto da Fase 1:** ✅ Atingido — `mvn test` 22/22 verdes; contrato JSON verificado manualmente contra baseline (4 cenários: sign válido, sign vazio, validate correto, validate errado).

### Fase 2 — Dispatcher dual-mode ✅ Concluída (Sprint 2)

11. ✅ Introduzir o despacho por modo em `AssinadorApplication`: `args[0] == "serve"` → stderr "Modo servidor (serve) ainda não implementado." + exit 1; qualquer outro comando segue pelo `CliRunner`. Ponto de extensão estabelecido sem ativar Spring.

### Fase 3 — Sprint 3: modo servidor HTTP ✅ Concluída (US-02.4 e suporte a US-01.5–01.9)

12. ✅ Substituído o `maven-shade-plugin` pelo `spring-boot-maven-plugin` (goal `repackage`); adicionados `spring-boot-starter-web` e `spring-boot-starter-test` via BOM do Spring Boot 3.3.5, mantendo `AssinadorApplication` como `Start-Class`.
13. ✅ Criado `presentation/http/SignatureController` com `POST /sign` e `POST /validate`, reusando **os mesmos** `SignUseCase`/`ValidateUseCase`.
14. ✅ Criados os DTOs HTTP (`SignHttpRequest`, `ValidateHttpRequest`, `SignatureHttpResponse`) + `GlobalExceptionHandler` (`ValidationException` → 400; `Exception` → 500).
15. ✅ Criado `infrastructure/config/AppConfig` (`@Configuration`) que declara o núcleo como beans, e `WebApplication` (`@SpringBootApplication`) como raiz do contexto.
16. ✅ Modo `serve` ativado (porta padrão 8080 + flag `--port`); `ServerStartupHandler` registra PID/porta em `~/.hubsaude/assinador.pid` (suporte a US-01.5, US-01.7, US-01.8).
17. ✅ `SignatureControllerTest`: 7 testes de integração (`@SpringBootTest` + MockMvc). Total do módulo: 29/29 verdes.

> Pendente do lado do CLI Go (Sprint 3, fora deste plano de refatoração Java): US-01.5–01.9 — iniciar/reusar/parar o servidor e invocá-lo por HTTP.

### Fase 4 — Sprint 3: PKCS#11 ✅ Concluída (US-02.5)

18. ✅ `domain/service/Pkcs11SignatureService` implementado via `SunPKCS11`; o campo `token` de `SignRequest` é usado como alias da chave privada no KeyStore. Nota: alocado em `domain/service/` em vez de `infrastructure/crypto/` previsto no diagrama (decisão de implementação — a classe depende apenas de `java.security`, sem infraestrutura).
19. ✅ Seleção automática via variável de ambiente `HUBSAUDE_PKCS11_LIBRARY`; `Pkcs11Config.fromEnvironment()` retorna `null` quando ausente; `AppConfig` instancia `Pkcs11ServiceFactory.create()` com fallback para `FakeSignatureService` e mensagem de aviso em stderr.
20. ✅ 5 testes unitários com Mockito em `Pkcs11SignatureServiceTest` cobrem: chave não encontrada, exceção no KeyStore (sign/validate), KeyStore vazio e encoding Base64 inválido. Testes com SoftHSM2 real requerem setup de SO e são documentados no Javadoc de `Pkcs11ServiceFactory` (fora do repositório).

## Impacto fora do Java

- **CLI Go (`assinatura`)**: ✅ não alterado — contrato do modo CLI preservado na Fase 1. Na Sprint 3, o CLI Go ganhará o caminho HTTP (US-01.6) apontando para os endpoints da Fase 3.
- **Documentação**: ✅ `CLAUDE.md` (seções "Java service architecture" e "Key Files") e `README.md` atualizados para refletir a nova estrutura de pacotes. Diagrama C4 de contêiner: 📋 pendente (atualizar na Sprint 3, quando o segundo modo de execução estiver ativo).

## Riscos e cuidados

- **Quebra do contrato com o CLI Go.** A troca do `toJson` manual pelo Jackson pode alterar a ordem/formatação das chaves do JSON. Garantir saída equivalente (ou ajustar o parsing do lado Go conjuntamente). Bloqueante na Fase 1.
- **Cold start do Spring no modo servidor.** Aceitável porque só ocorre no `serve`; o modo CLI permanece sem Spring. Não regredir isso transformando todo o núcleo em beans Spring.
- **Repackage do fat-jar.** O `spring-boot-maven-plugin` usa seu próprio launcher; é preciso configurar a `start-class`/`mainClass` para o dispatcher `AssinadorApplication`, senão o modo CLI quebra. Validar na Fase 3.
- **Escopo da refatoração.** Fase 1 é refatoração pura (sem novas features); resistir à tentação de já implementar HTTP/PKCS#11 antes de o núcleo estar limpo e testado.

## Ordem sugerida

Fase 1 (estrutural) → Fase 2 (dispatcher) ainda dentro da Sprint 2; Fases 3 e 4 abrem a Sprint 3.

## Decisões tomadas durante a refatoração

1. ✅ **Framework HTTP: Spring Boot** — confirmado para a Fase 3 (US-02.4). Alternativas mais leves (Javalin, `HttpServer`) foram descartadas por fugir do objetivo didático e do que a especificação sugere.
2. ✅ **Validação na camada `application`** — `RequestValidator` implementado como fonte única; CLI e HTTP compartilharão as mesmas regras. Pode-se somar `@Valid` nos DTOs HTTP como barreira extra na Fase 3, delegando a regra de negócio ao mesmo validador.
3. ✅ **Renomear `SignatureResponse` → `SignatureResult`** — renomeado no CP2, separando o resultado de domínio do futuro DTO de transporte HTTP (`SignatureHttpResponse`).

---

## Execução em checkpoints (Fases 1 e 2) — ✅ Concluída

> Todos os checkpoints foram executados e commitados. `mvn test`: 22/22 verdes. Contrato externo preservado em todos os cenários. Esta seção é mantida como **registro histórico** da execução.

Esta seção operacionalizou as Fases 1 e 2 acima em **unidades commitáveis**, cada uma terminando com o build verde. O agrupamento segue dois princípios:

- **Build sempre verde:** cada checkpoint compila e passa `mvn test` ao final. Nenhum estado intermediário deixa o código quebrado.
- **Atomicidade onde o comportamento depende:** alguns sub-passos do plano **não** podem ser commitados isolados sem regredir o comportamento externo. Em especial, o sub-passo 4 (remover validação do `FakeSignatureService`), o 6 (criar use cases) e o 9 (religar o entrypoint) **têm de cair juntos** — entre remover a validação e religar o entrypoint nos use cases, o CLI perderia a validação, mudando mensagem de erro e exit code, e quebrando os testes de integração Go `Sign_ContentVazio` e `Validate_AssinaturaErrada`. Por isso formam um único checkpoint (CP3).

> **Workflow de commit:** o usuário faz os próprios commits. Ao final de cada checkpoint, **pare**, deixe o build verde e forneça o comando de commit pronto (não execute `git commit`). Sugestões de mensagem em cada CP abaixo.

### Decisões em aberto — resolvidas para esta execução

1. **Framework HTTP:** Spring Boot — mas **fora do escopo** das Fases 1–2. Não criar `presentation/http/` nem `infrastructure/crypto/` agora, nem diretórios/placeholders vazios.
2. **Validação na `application`** (`RequestValidator`) como fonte única. Confirmado.
3. **Renomear `SignatureResponse` → `SignatureResult`: SIM, agora** (no CP2). O codebase é minúsculo; adiar só aumenta o churn depois.
4. **Ordem das chaves no JSON:** o critério "byte-a-byte" da Fase 1 é **relaxado para equivalência de campos**. Justificativa: o único consumidor (CLI Go) faz `json.Unmarshal` num `map[string]interface{}` (ver `projetos/assinatura/cmd/integration_test.go`), logo depende dos **nomes** dos campos (`signature`, `valid`, `message`), do tratamento de `null`, do stream (stdout/stderr) e do exit code — **não** da ordem das chaves. Manter o `domain` livre de anotações Jackson (sem `@JsonPropertyOrder`) vale mais que ordem idêntica. Se no futuro a ordem precisar ser fixada, usar um mixin no `JsonMapper`, nunca anotação no domínio.

### Contrato externo a preservar (invariante de todas as fases)

Verificado contra `projetos/assinatura/cmd/integration_test.go`:

| Comando | Entrada | Saída | Stream | Exit |
|---------|---------|-------|--------|------|
| `sign` | `--content` não-vazio | `{"signature":"MOCKED_SIGNATURE_BASE64_==","valid":true,"message":"Assinatura criada com sucesso"}` | stdout | 0 |
| `sign` | `--content` vazio/ausente | `{"signature":null,"valid":false,"message":"Parâmetro 'content' inválido ou ausente"}` | stderr | 1 |
| `validate` | content + signature correta | `{...,"valid":true,"message":"Assinatura é válida"}` | stdout | 0 |
| `validate` | content + signature errada | `{...,"valid":false,"message":"Assinatura é inválida"}` | stderr | 1 |
| `validate` | content/signature ausente | `{"signature":null,"valid":false,"message":"Parâmetro '<campo>' inválido ou ausente"}` | stderr | 1 |
| (qualquer) | comando desconhecido / sem args | mensagem de uso | stderr | 1 |

Constante simulada: `MOCKED_SIGNATURE_BASE64_==`. Mensagens de erro **literais** (extraídas hoje de `FakeSignatureService`) devem ser preservadas exatamente.

### CP0 — Baseline do contrato (sem commit)

Antes de tocar em nada, capturar a saída atual para comparar depois do CP1.

```bash
cd projetos/assinador-java && mvn -q package
JAR=target/assinador.jar
java -jar $JAR sign --content "doc"                 ; echo "exit=$?"   # stdout, exit 0
java -jar $JAR sign --content ""                     ; echo "exit=$?"   # stderr, exit 1
java -jar $JAR validate --content d --signature MOCKED_SIGNATURE_BASE64_== ; echo "exit=$?"  # stdout, exit 0
java -jar $JAR validate --content d --signature errada ; echo "exit=$?"  # stderr, exit 1
```

Guardar as 4 saídas (mentalmente ou num arquivo temporário fora do repo). Servem de gabarito do CP1.

---

### CP1 — Jackson + JsonMapper + fat-jar (contrato JSON isolado)

**Objetivo:** trocar o `toJson`/`escapeJson` manual do `Main` por Jackson, **sem** mover a estrutura de pacotes ainda. É o único checkpoint que pode alterar o contrato externo — por isso fica isolado e é verificado contra o baseline do CP0.

**Ponto crítico (não óbvio):** o jar atual é *thin* (a `maven-jar-plugin` não empacota dependências) e hoje isso funciona porque o runtime tem **zero** dependências. Ao adicionar Jackson, `java -jar assinador.jar` passaria a lançar `NoClassDefFoundError`. O CLI Go distribui **apenas** `assinador.jar` (bundle/auto-download de arquivo único). Logo, o CP1 **tem** de produzir um fat-jar self-contained via `maven-shade-plugin`.

**`pom.xml` — adicionar dependência:**
```xml
<dependency>
    <groupId>com.fasterxml.jackson.core</groupId>
    <artifactId>jackson-databind</artifactId>
    <version>2.18.2</version>
</dependency>
```

**`pom.xml` — adicionar shade (empacota Jackson em `target/assinador.jar`):**
```xml
<plugin>
    <groupId>org.apache.maven.plugins</groupId>
    <artifactId>maven-shade-plugin</artifactId>
    <version>3.6.0</version>
    <executions>
        <execution>
            <phase>package</phase>
            <goals><goal>shade</goal></goals>
            <configuration>
                <finalName>assinador</finalName>
                <transformers>
                    <transformer implementation="org.apache.maven.plugins.shade.resource.ManifestResourceTransformer">
                        <mainClass>com.hubsaude.assinador.Main</mainClass>
                    </transformer>
                </transformers>
            </configuration>
        </execution>
    </executions>
</plugin>
```
Manter a `maven-jar-plugin` como está (mainClass + `finalName=assinador`). A shade roda depois, na mesma fase `package`, e sobrescreve `target/assinador.jar` com o fat-jar. (A `<mainClass>` aparece em **dois** lugares — jar-plugin e shade; ambos serão atualizados no CP4 ao renomear `Main`.)

**Criar `src/main/java/com/hubsaude/assinador/infrastructure/json/JsonMapper.java`:**
```java
package com.hubsaude.assinador.infrastructure.json;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

public final class JsonMapper {
    private static final ObjectMapper MAPPER = new ObjectMapper();
    private JsonMapper() {}
    public static String toJson(Object value) {
        try {
            return MAPPER.writeValueAsString(value);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Falha ao serializar resposta em JSON", e);
        }
    }
}
```
`toJson(Object)` é deliberadamente genérico — não cita o tipo da resposta, então sobrevive ao rename do CP2 sem mudança. O `ObjectMapper` default reproduz o contrato: `isValid()`→`valid`, `getSignature()`→`signature` (inclui `"signature":null` quando nulo, pois nulos não são omitidos por default), `getMessage()`→`message`.

**Editar `Main.java`:** trocar `String json = toJson(response);` por `String json = JsonMapper.toJson(response);` em `printResponse`; **remover** os métodos `toJson` e `escapeJson`.

**Testes:** os 3 testes `toJson_*` em `FakeSignatureServiceTest` referenciam `Main.toJson` (que deixa de existir) → movê-los para um novo `src/test/java/com/hubsaude/assinador/infrastructure/json/JsonMapperTest.java`, asserindo via `contains` (ordem livre):
- `JsonMapper.toJson(new SignatureResponse("SIG==", true, "ok"))` contém `"signature":"SIG=="`, `"valid":true`, `"message":"ok"`.
- resposta com signature nula contém `"signature":null`.
- message com aspas (`erro "especial"`) → JSON contém `\"especial\"` (Jackson escapa).

**Verificação / Definition of done:**
- `mvn package` gera `target/assinador.jar`; `unzip -l target/assinador.jar | grep jackson` mostra as classes do Jackson embutidas.
- Repetir os 4 comandos do CP0 — campos/streams/exit codes equivalentes ao baseline.
- `mvn test` verde.

**Commit sugerido (usuário executa):**
```bash
git add projetos/assinador-java/pom.xml projetos/assinador-java/src
git commit -m "refatoracao: serializacao JSON via Jackson + fat-jar (shade)"
```

---

### CP2 — Mover e renomear o domínio (move/rename mecânico)

**Objetivo:** criar a estrutura de pacotes do núcleo e mover os tipos. **Sem mudança de comportamento** — a validação ainda permanece dentro do `FakeSignatureService` neste checkpoint (sai só no CP3).

**Movimentações:**
- `domain/SignRequest.java` → `domain/model/SignRequest.java` (package `...domain.model`).
- `domain/ValidateRequest.java` → `domain/model/ValidateRequest.java`.
- `domain/SignatureResponse.java` → `domain/model/SignatureResult.java` — **renomear classe e arquivo** `SignatureResponse`→`SignatureResult`.
- `SignatureService.java` → `domain/service/SignatureService.java` (package `...domain.service`); retorno passa a `SignatureResult`.
- `FakeSignatureService.java` → `domain/service/FakeSignatureService.java`; manter a `FAKE_SIGNATURE` e (por ora) a validação; retornos `SignatureResult`.

**Atualizar referências/imports:** `Main.java` (imports + tipo `SignatureResult`), `FakeSignatureServiceTest.java` (imports + tipo + chamadas). O `JsonMapper` não cita o tipo, então não muda.

**Verificação / DoD:** `mvn test` verde; os 4 comandos do CP0 inalterados. Nenhuma busca por `SignatureResponse` deve restar (`grep -r SignatureResponse src` vazio).

**Commit sugerido:**
```bash
git add projetos/assinador-java/src
git commit -m "refatoracao: move DTOs/servico para domain.model e domain.service; renomeia SignatureResponse -> SignatureResult"
```

---

### CP3 — Camada application: validação + use cases (trio atômico)

**Objetivo:** extrair a validação do `FakeSignatureService` para uma fonte única na `application`, introduzir os use cases e religar o entrypoint — **tudo junto**, preservando mensagens e exit codes exatos.

**Criar `application/validation/ValidationException.java`:**
```java
package com.hubsaude.assinador.application.validation;

public class ValidationException extends RuntimeException {
    public ValidationException(String message) { super(message); }
}
```

**Criar `application/validation/RequestValidator.java`** — fonte única de verdade, com as mensagens **literais** de hoje:
```java
package com.hubsaude.assinador.application.validation;

import com.hubsaude.assinador.domain.model.SignRequest;
import com.hubsaude.assinador.domain.model.ValidateRequest;

public class RequestValidator {
    public void validateSign(SignRequest r) {
        if (r == null || isBlank(r.getContent()))
            throw new ValidationException("Parâmetro 'content' inválido ou ausente");
    }
    public void validateValidate(ValidateRequest r) {
        if (r == null || isBlank(r.getContent()))
            throw new ValidationException("Parâmetro 'content' inválido ou ausente");
        if (isBlank(r.getSignature()))
            throw new ValidationException("Parâmetro 'signature' inválido ou ausente");
    }
    private boolean isBlank(String s) { return s == null || s.isBlank(); }
}
```

**Criar `application/SignUseCase.java` e `application/ValidateUseCase.java`:** cada um recebe `SignatureService` + `RequestValidator` no construtor; valida e então delega (`return service.sign(request)` / `service.validate(request)`).

**Remover a validação de `FakeSignatureService`:** `sign` passa a retornar sempre o sucesso simulado (assume entrada válida); `validate` mantém **apenas** a lógica de match (`FAKE_SIGNATURE.equals(signature)` → "Assinatura é válida"/"Assinatura é inválida"), sem mais checagens de null/blank.

**Religar o entrypoint (`Main`):** instanciar `FakeSignatureService`, `RequestValidator`, e os use cases; `handleSign`/`handleValidate` chamam os use cases dentro de `try/catch (ValidationException e)`. No catch, construir o resultado de erro equivalente ao de hoje — `new SignatureResult(null, false, e.getMessage())` — e passá-lo ao mesmo `printResponse` (stderr + exit 1). Assim mensagem e exit code ficam idênticos ao comportamento atual.

**Migração de testes:**
- `FakeSignatureServiceTest`: **remover** os testes de erro de parâmetro (`sign_conteudoNulo/Vazio/Espacos`, `validate_conteudoNulo/Espacos`, `validate_assinaturaNula/Vazia`) — a validação não vive mais ali. Manter `sign_conteudoValido`, `validate_assinaturaCorreta`, `validate_assinaturaErrada`.
- Novo `RequestValidatorTest`: cobrir todos os casos null/blank de `validateSign`/`validateValidate`, asserindo a mensagem da `ValidationException` (`contains("content")` / `contains("signature")`).
- Novo `SignUseCaseTest`/`ValidateUseCaseTest` (ou um `UseCaseFluxoTest`): fluxo feliz e propagação da `ValidationException` em entrada inválida; e o fluxo completo sign→validate migra para o nível de use case.

**Verificação / DoD:** `mvn test` verde; os 4 comandos do CP0 inalterados (em especial os de erro: stderr + exit 1 + mensagem literal). `grep -rn "inválido ou ausente" src/main` deve aparecer **só** em `RequestValidator`.

**Commit sugerido:**
```bash
git add projetos/assinador-java/src
git commit -m "refatoracao: extrai validacao para application (RequestValidator) e introduz use cases"
```

---

### CP4 — Presentation/CLI + composition root enxuto

**Objetivo:** tirar parsing e formatação do `Main`, e reduzir o entrypoint a um composition root.

- **`presentation/cli/CliRunner.java`:** parsing de argumentos (hoje `handleSign`/`handleValidate`), montando os requests e chamando os use cases. Recebe os use cases + `CliPresenter` por construtor.
- **`presentation/cli/CliPresenter.java`:** formatação de saída (hoje `printResponse`): se `valid` → `System.out.println(JsonMapper.toJson(result))`; senão → `System.err.println(...)` + `System.exit(1)`. Também formata o resultado de erro de `ValidationException` e os erros de "comando desconhecido"/"sem args".
- **Renomear `Main.java` → `AssinadorApplication.java`** (composition root): monta o núcleo (`FakeSignatureService`, `RequestValidator`, use cases), o `CliPresenter`, o `CliRunner`, e despacha. **Atualizar `<mainClass>` em DOIS lugares do `pom.xml`** (maven-jar-plugin e maven-shade-plugin) para `com.hubsaude.assinador.AssinadorApplication` — se esquecer, o manifest aponta para classe inexistente e `java -jar` quebra.

**Verificação / DoD:** `mvn package` + repetir os 4 comandos do CP0 (agora exercitam o manifest renomeado) — equivalentes. `mvn test` verde. `unzip -p target/assinador.jar META-INF/MANIFEST.MF | grep Main-Class` mostra `AssinadorApplication`.

**Commit sugerido:**
```bash
git add projetos/assinador-java/pom.xml projetos/assinador-java/src
git commit -m "refatoracao: extrai presentation/cli (CliRunner, CliPresenter) e renomeia Main -> AssinadorApplication"
```

---

### CP5 — Dispatcher dual-mode (Fase 2, ponte para a Sprint 3)

**Objetivo:** estabelecer o ponto de extensão do modo servidor **sem** subir Spring.

Em `AssinadorApplication`, antes de despachar ao `CliRunner`: se `args[0].equals("serve")`, imprimir em stderr algo como `"Modo servidor (serve) ainda não implementado."` e `System.exit(1)`; qualquer outro comando segue o fluxo CLI normal. Adicionar 1–2 testes do dispatcher se viável (ou validar manualmente).

**Verificação / DoD:** `java -jar target/assinador.jar serve` → mensagem "não implementado" + exit 1; `sign`/`validate` inalterados; `mvn test` verde.

**Commit sugerido:**
```bash
git add projetos/assinador-java/src
git commit -m "feat: dispatcher reserva o modo 'serve' (ainda nao implementado)"
```

---

### Pós-execução ✅ Concluída

- ✅ `CLAUDE.md` — seções "Java service architecture" e "Key Files" atualizadas.
- ✅ `README.md` — árvore de estrutura do projeto atualizada.
- 📋 Diagrama C4 de contêiner — atualizar na Sprint 3, quando o modo servidor estiver ativo.

> `presentation/http/` e `infrastructure/crypto/` **não foram criados** neste ciclo — são Fases 3 e 4 (Sprint 3).

### Mapa checkpoint → sub-passos do plano

| Checkpoint | Sub-passos cobertos |
|------------|---------------------|
| CP1 | 1, 8 (+ fat-jar, não listado originalmente) |
| CP2 | 2, 3, parte do 4 (move, sem remover validação) |
| CP3 | 4 (remoção), 5, 6, 9, 10 (migração de testes) |
| CP4 | 7, 9 (composition root) |
| CP5 | 11 (Fase 2) |
