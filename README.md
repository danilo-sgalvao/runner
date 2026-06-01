# Sistema Runner

CLI multiplataforma para execução de aplicações Java do ecossistema HubSaúde, desenvolvido como trabalho prático da disciplina de Implementação e Integração — Bacharelado em Engenharia de Software (UFG, 2026).

---

## Sobre o projeto

O **Sistema Runner** facilita o acesso à funcionalidade de execução de aplicações Java via linha de comandos, permitindo que usuários utilizem as ferramentas do HubSaúde sem precisar configurar ou instalar o Java manualmente.

O projeto é composto por:

- **`assinatura`** — CLI multiplataforma (Go) para criação e validação de assinaturas digitais
- **`assinador.jar`** — Aplicação Java que realiza (de forma simulada) as operações de assinatura
- **`simulador`** — CLI multiplataforma (Go) para gerenciamento do Simulador do HubSaúde *(previsto para próximas sprints)*

---

## Download

Baixe o binário mais recente para sua plataforma na página de [Releases](https://github.com/danilo-sgalvao/runner/releases):

| Plataforma | Arquivo |
|---|---|
| Windows | `assinatura-<versão>-windows-amd64.exe` |
| Linux | `assinatura-<versão>-linux-amd64` |
| macOS | `assinatura-<versão>-darwin-amd64` |

---

## Uso (binário baixado)

> **Java não precisa estar instalado.** Na primeira execução, o `assinatura` detecta automaticamente o Java 21 do sistema; se não houver, baixa um JRE compatível e o instala em `~/.hubsaude/jre`. Tudo sem intervenção do usuário.

O binário publicado nos Releases vem com o nome completo da plataforma (ex.: `assinatura-v0.2.0-windows-amd64.exe`). O nome `assinatura` usado nos exemplos abaixo **não está disponível automaticamente** — para usá-lo você precisa de uma das opções:

- **Chamar pelo caminho/nome completo** do arquivo baixado, ou
- **Renomear** o binário para `assinatura` (Linux/macOS) / `assinatura.exe` (Windows) e **adicioná-lo ao `PATH`**.

Os exemplos a seguir assumem que isso já foi feito. Caso contrário, substitua `assinatura` pelo caminho real do executável (ex.: `.\assinatura-v0.2.0-windows-amd64.exe`).

### Exibir a versão

```bash
assinatura version
```

### Criar uma assinatura digital

```bash
assinatura sign --content "conteudo a ser assinado"
```

Exemplo de saída:

```json
{"signature":"MOCKED_SIGNATURE_BASE64_==","valid":true,"message":"Assinatura criada com sucesso"}
```

### Validar uma assinatura digital

```bash
assinatura validate --content "conteudo a ser assinado" --signature "MOCKED_SIGNATURE_BASE64_=="
```

Exemplo de saída:

```json
{"signature":"MOCKED_SIGNATURE_BASE64_==","valid":true,"message":"Assinatura é válida"}
```

### Ajuda

```bash
assinatura --help
assinatura sign --help
assinatura validate --help
```

---

## Compilar e rodar a partir do código-fonte

Para quem clonou o repositório, a ordem importa: **o `assinador.jar` precisa ser compilado antes** de executar o CLI, porque o CLI o localiza e o invoca como subprocesso.

### Pré-requisitos

- [Go 1.24+](https://go.dev/dl/)
- [Java JDK 21+](https://adoptium.net/)
- [Maven 3.9+](https://maven.apache.org/download.cgi)

### 1. Clonar o repositório

```bash
git clone https://github.com/danilo-sgalvao/runner.git
cd runner
```

### 2. Compilar o assinador.jar (passo obrigatório, vem primeiro)

```bash
cd projetos/assinador-java
mvn package          # gera target/assinador.jar
cd ../..
```

### 3. Executar o CLI

Em modo de desenvolvimento, direto pelo Go (não requer renomear binário nem PATH):

```bash
cd projetos/assinatura
go run . sign --content "teste"
```

Ou gere o binário nativo e chame-o pelo caminho local (no Windows o `.exe` exige o prefixo `.\`; em Linux/macOS, `./`):

```bash
# Windows
go build -o assinatura.exe .
.\assinatura.exe sign --content "teste"

# Linux / macOS
go build -o assinatura .
./assinatura sign --content "teste"
```

### 4. Executar os testes

```bash
# Testes Go (na pasta projetos/assinatura)
go test ./...

# Testes Java, incl. integração HTTP (na pasta projetos/assinador-java)
mvn test
```

### 5. Gerar binários para as três plataformas

```bash
# Ainda na pasta projetos/assinatura
GOOS=linux  GOARCH=amd64 go build -o assinatura-linux .
GOOS=darwin GOARCH=amd64 go build -o assinatura-macos .
GOOS=windows GOARCH=amd64 go build -o assinatura-windows.exe .
cd ../..
```

---

## Modo servidor (HTTP)

Além do modo CLI (uma invocação por comando), o `assinador.jar` pode rodar como **servidor HTTP** permanente, expondo os mesmos casos de uso via REST — útil para menor latência em chamadas repetidas. Hoje o servidor é iniciado diretamente pelo `java -jar`; a integração pelo CLI Go (`assinatura serve`, reuso e parada do processo) está prevista para uma próxima sprint.

### Iniciar o servidor

```bash
# porta padrão 8080
java -jar projetos/assinador-java/target/assinador.jar serve

# porta personalizada
java -jar projetos/assinador-java/target/assinador.jar serve --port 9090
```

Ao subir, o servidor registra `{"pid":...,"port":...}` em `~/.hubsaude/assinador.pid`.

### Chamar os endpoints

**Linux / macOS (bash) — `curl`:**

```bash
# POST /sign
curl -X POST http://localhost:8080/sign \
  -H "Content-Type: application/json" \
  -d '{"content":"documento"}'
# → {"signature":"MOCKED_SIGNATURE_BASE64_==","valid":true,"message":"Assinatura criada com sucesso"}

# POST /validate
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"content":"documento","signature":"MOCKED_SIGNATURE_BASE64_=="}'
# → {"signature":"MOCKED_SIGNATURE_BASE64_==","valid":true,"message":"Assinatura é válida"}
```

**Windows (PowerShell) — `Invoke-RestMethod`:**

No PowerShell, `curl` é um alias de `Invoke-WebRequest` e não aceita a mesma sintaxe; use `Invoke-RestMethod`, que ainda desserializa a resposta JSON em um objeto.

```powershell
# POST /sign
Invoke-RestMethod -Uri http://localhost:8080/sign -Method Post `
  -ContentType "application/json" `
  -Body '{"content":"documento"}'
# → signature                  valid message
#   ---------                  ----- -------
#   MOCKED_SIGNATURE_BASE64_==   True Assinatura criada com sucesso

# POST /validate
Invoke-RestMethod -Uri http://localhost:8080/validate -Method Post `
  -ContentType "application/json" `
  -Body '{"content":"documento","signature":"MOCKED_SIGNATURE_BASE64_=="}'
# → valid=True, message="Assinatura é válida"
```

> Em respostas **HTTP 400** (parâmetro ausente/vazio), `Invoke-RestMethod` lança um erro de terminação. No PowerShell 7+ o corpo fica em `$_.ErrorDetails.Message`; no Windows PowerShell 5.1 esse campo vem vazio e é preciso ler o stream da resposta:
>
> ```powershell
> try {
>   Invoke-RestMethod -Uri http://localhost:8080/sign -Method Post `
>     -ContentType "application/json" -Body '{"content":""}'
> } catch {
>   $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
>   $reader.ReadToEnd()
>   # → {"signature":null,"valid":false,"message":"Parâmetro 'content' inválido ou ausente"}
> }
> ```

Parâmetros ausentes/vazios retornam **HTTP 400**; uma assinatura que não confere retorna **HTTP 200** com `"valid":false` (é resultado de negócio, não erro de entrada).

---

## Verificando a autenticidade dos artefatos

Todos os artefatos são assinados com [Cosign](https://docs.sigstore.dev/cosign/overview/) via Sigstore. Para verificar a autenticidade de um binário baixado:

### Linux / macOS

```bash
cosign verify-blob \
  --bundle assinatura-v0.2.0-linux-amd64.bundle \
  assinatura-v0.2.0-linux-amd64
```

### Windows

```powershell
cosign verify-blob `
  --bundle assinatura-v0.2.0-windows-amd64.exe.bundle `
  assinatura-v0.2.0-windows-amd64.exe
```

Se a verificação for bem-sucedida, o Cosign exibirá:

```
Verified OK
```

### Verificando checksums SHA256

Cada release inclui um arquivo `checksums.txt` com os hashes SHA256 de todos os binários:

```bash
sha256sum --check checksums.txt
```

---

## Instalar o Cosign

Para instalar o Cosign, acesse: https://docs.sigstore.dev/cosign/system_config/installation/

---

## CI/CD

O projeto utiliza **GitHub Actions** com dois workflows:

| Workflow | Gatilho | O que faz |
|---|---|---|
| `build.yml` | Push na `main` | Compila para as 3 plataformas e salva os artefatos |
| `release.yml` | Criação de tag `v*` | Compila, gera checksums, assina com Cosign e publica no GitHub Releases |

Para publicar uma nova release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

---

## Estrutura do projeto

```
runner/
├── .github/
│   └── workflows/
│       ├── build.yml                       # Pipeline de build contínuo
│       └── release.yml                     # Pipeline de release com Cosign
├── docs/                                   # Especificação, planos e relatórios
├── projetos/
│   ├── assinador-java/                     # Serviço Java (Maven, fat-jar)
│   │   ├── pom.xml
│   │   └── src/
│   │       ├── main/java/com/hubsaude/assinador/
│   │       │   ├── AssinadorApplication.java   # Composition root; dispatcher CLI / serve (--port)
│   │       │   ├── WebApplication.java     # @SpringBootApplication (raiz do contexto, modo serve)
│   │       │   ├── domain/
│   │       │   │   ├── model/              # DTOs: SignRequest, ValidateRequest, SignatureResult
│   │       │   │   └── service/            # SignatureService (interface) + FakeSignatureService
│   │       │   ├── application/
│   │       │   │   ├── SignUseCase.java
│   │       │   │   ├── ValidateUseCase.java
│   │       │   │   └── validation/         # RequestValidator + ValidationException
│   │       │   ├── presentation/
│   │       │   │   ├── cli/
│   │       │   │   │   ├── CliRunner.java      # Parsing de args
│   │       │   │   │   └── CliPresenter.java   # Formatação JSON + exit codes
│   │       │   │   └── http/               # modo serve: POST /sign, POST /validate
│   │       │   │       ├── SignatureController.java
│   │       │   │       ├── GlobalExceptionHandler.java
│   │       │   │       └── dto/            # SignHttpRequest, ValidateHttpRequest, SignatureHttpResponse
│   │       │   └── infrastructure/
│   │       │       ├── json/
│   │       │       │   └── JsonMapper.java     # Serialização Jackson (modo CLI)
│   │       │       ├── config/
│   │       │       │   └── AppConfig.java  # @Configuration: núcleo como beans Spring
│   │       │       └── ServerStartupHandler.java  # registra PID/porta em ~/.hubsaude/
│   │       └── test/java/com/hubsaude/assinador/
│   │           ├── FakeSignatureServiceTest.java
│   │           ├── application/
│   │           │   ├── UseCasesTest.java
│   │           │   └── validation/
│   │           │       └── RequestValidatorTest.java
│   │           ├── infrastructure/json/
│   │           │   └── JsonMapperTest.java
│   │           └── presentation/http/
│   │               └── SignatureControllerTest.java
│   └── assinatura/                         # CLI Go (Cobra)
│       ├── cmd/
│       │   ├── root.go                     # Comando raiz
│       │   ├── version.go                  # Subcomando version
│       │   ├── sign.go                     # Subcomando sign
│       │   ├── validate.go                 # Subcomando validate
│       │   ├── jar.go                      # Localização do assinador.jar
│       │   └── *_test.go                   # Testes unitários
│       ├── internal/jre/
│       │   ├── manager.go                  # Detecção e auto-download do JRE
│       │   └── manager_test.go
│       ├── main.go                         # Ponto de entrada do CLI
│       └── go.mod
├── release.json                            # Metadados/URLs do JRE para download
└── README.md
```

---

## Releases

| Versão | O que tem |
|---|---|
| [v0.1.0](https://github.com/danilo-sgalvao/runner/releases/tag/v0.1.0) | CLI base com comando `version`, pipelines CI/CD, binários para 3 plataformas |
| [v0.2.0](https://github.com/danilo-sgalvao/runner/releases/tag/v0.2.0) | Comandos `sign` e `validate`, integração com `assinador.jar` |

---

## Versionamento

O projeto segue [Versionamento Semântico (SemVer)](https://semver.org/lang/pt-BR/):

- `MAJOR`: mudanças incompatíveis com versões anteriores
- `MINOR`: novas funcionalidades compatíveis com versões anteriores
- `PATCH`: correções de bugs

---

## Contexto acadêmico

Este projeto é desenvolvido como trabalho prático da disciplina de **Implementação e Integração** do Bacharelado em Engenharia de Software da **Universidade Federal de Goiás (UFG)**, em parceria com a **Secretaria de Estado de Saúde de Goiás (SES)** no âmbito da plataforma **HubSaúde**.

---

## Licença

Distribuído sob a licença MIT. Consulte o arquivo `LICENSE` para mais informações.
