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

## Uso

> **Java não precisa estar instalado.** Na primeira execução, o `assinatura` detecta automaticamente o Java 21 do sistema; se não houver, baixa um JRE compatível e o instala em `~/.hubsaude/jre`. Tudo sem intervenção do usuário.

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

## Como compilar o projeto

### Pré-requisitos

- [Go 1.24+](https://go.dev/dl/)
- [Java JDK 21+](https://adoptium.net/)
- [Maven 3.9+](https://maven.apache.org/download.cgi)

### 1. Clonar o repositório

```bash
git clone https://github.com/danilo-sgalvao/runner.git
cd runner
```

### 2. Compilar o assinador.jar

```bash
cd projetos/assinador-java
mvn package
cd ../..
```

### 3. Executar o CLI em modo de desenvolvimento

```bash
cd projetos/assinatura
go run . sign --content "teste"
```

### 4. Executar os testes

```bash
# Testes Go (na pasta projetos/assinatura)
go test ./...

# Testes Java (na pasta projetos/assinador-java)
mvn test
```

### 5. Gerar binário nativo

```bash
# Ainda na pasta projetos/assinatura
# Windows
go build -o assinatura.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o assinatura-linux .

# macOS
GOOS=darwin GOARCH=amd64 go build -o assinatura-macos .
cd ../..
```

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
│   │       │   ├── AssinadorApplication.java   # Composition root; dispatcher CLI / serve
│   │       │   ├── domain/
│   │       │   │   ├── model/              # DTOs: SignRequest, ValidateRequest, SignatureResult
│   │       │   │   └── service/            # SignatureService (interface) + FakeSignatureService
│   │       │   ├── application/
│   │       │   │   ├── SignUseCase.java
│   │       │   │   ├── ValidateUseCase.java
│   │       │   │   └── validation/         # RequestValidator + ValidationException
│   │       │   ├── presentation/cli/
│   │       │   │   ├── CliRunner.java      # Parsing de args
│   │       │   │   └── CliPresenter.java   # Formatação JSON + exit codes
│   │       │   └── infrastructure/json/
│   │       │       └── JsonMapper.java     # Serialização Jackson
│   │       └── test/java/com/hubsaude/assinador/
│   │           ├── FakeSignatureServiceTest.java
│   │           ├── application/
│   │           │   ├── UseCasesTest.java
│   │           │   └── validation/
│   │           │       └── RequestValidatorTest.java
│   │           └── infrastructure/json/
│   │               └── JsonMapperTest.java
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
