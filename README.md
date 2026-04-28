# Sistema Runner

CLI multiplataforma para execução de aplicações Java do ecossistema HubSaúde, desenvolvido como trabalho prático da disciplina de Implementação e Integração — Bacharelado em Engenharia de Software (UFG, 2026).

---

## Sobre o projeto

O **Sistema Runner** facilita o acesso à funcionalidade de execução de aplicações Java via linha de comandos, permitindo que usuários utilizem as ferramentas do HubSaúde sem precisar configurar ou instalar o Java manualmente.

O projeto é composto por:

- **`assinatura`** — CLI multiplataforma (Go) para criação e validação de assinaturas digitais
- **`assinador.jar`** — Aplicação Java que realiza (de forma simulada) as operações de assinatura
- **`simulador`** — CLI multiplataforma (Go) para gerenciamento do Simulador do HubSaúde

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

### Exibir a versão

```bash
assinatura version
```

### Criar uma assinatura digital

```bash
assinatura sign --file documento.xml --cert certificado.pem
```

### Validar uma assinatura digital

```bash
assinatura validate --file documento.xml --signature assinatura.xml
```

### Ajuda

```bash
assinatura --help
assinatura sign --help
assinatura validate --help
```

---

## Verificando a autenticidade dos artefatos

Todos os artefatos são assinados com [Cosign](https://docs.sigstore.dev/cosign/overview/) via Sigstore. Para verificar a autenticidade de um binário baixado:

### Linux / macOS

```bash
cosign verify-blob \
  --bundle assinatura-v0.1.0-linux-amd64.bundle \
  assinatura-v0.1.0-linux-amd64
```

### Windows

```powershell
cosign verify-blob `
  --bundle assinatura-v0.1.0-windows-amd64.exe.bundle `
  assinatura-v0.1.0-windows-amd64.exe
```

Se a verificação for bem-sucedida, o Cosign exibirá:

```
Verified OK
```

### Verificando checksums SHA256

Cada release inclui um arquivo `checksums.txt` com os hashes SHA256 de todos os binários. Para verificar:

```bash
sha256sum --check checksums.txt
```

---

## Instalação do Cosign

Para instalar o Cosign, acesse: https://docs.sigstore.dev/cosign/system_config/installation/

---

## Como compilar o projeto

### Pré-requisitos

- [Go 1.24+](https://go.dev/dl/)

### Compilar para sua plataforma atual

```bash
go build -o assinatura .
```

### Compilar para todas as plataformas

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o assinatura-windows-amd64.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o assinatura-linux-amd64 .

# macOS
GOOS=darwin GOARCH=amd64 go build -o assinatura-darwin-amd64 .
```

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
│       ├── build.yml       # Pipeline de build contínuo
│       └── release.yml     # Pipeline de release com Cosign
├── cmd/
│   ├── root.go             # Comando raiz do CLI
│   └── version.go          # Subcomando version
├── main.go                 # Ponto de entrada
├── go.mod
└── README.md
```

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
