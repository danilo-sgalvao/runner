# CLAUDE.md — Sistema Runner

Guia de contexto para assistentes AI trabalhando neste repositório.

## Visão geral

CLI multiplataforma para execução de aplicações Java do ecossistema HubSaúde (UFG/SES-GO, 2026).
Dois CLIs em Go gerenciam dois serviços Java: `assinatura` (assinador.jar) e `simulador` (validador FHIR).

```
runner/
├── projetos/
│   ├── go.work               ← workspace Go: shared + assinatura + simulador
│   ├── shared/               ← módulo compartilhado (jre, config, release, process)
│   ├── assinatura/           ← CLI Go: sign, validate, start, stop
│   ├── simulador/            ← CLI Go: start, stop, status
│   └── assinador-java/       ← serviço Java (Spring Boot, fat-jar)
├── release.json              ← URLs de download: JRE, assinador.jar, simulador.jar
└── .github/workflows/        ← build.yml (CI), release.yml (Cosign + Releases)
```

## Workspace Go (multi-módulo)

`projetos/go.work` declara três módulos:

| Módulo | Caminho | Propósito |
|--------|---------|-----------|
| `github.com/danilo-sgalvao/runner/shared` | `projetos/shared` | jre, config, release, process — reusados pelos dois CLIs |
| `github.com/danilo-sgalvao/runner/assinatura` | `projetos/assinatura` | CLI do assinador |
| `github.com/danilo-sgalvao/runner/simulador` | `projetos/simulador` | CLI do simulador |

Cada `go.mod` consumidor tem `replace .../shared => ../shared` para builds fora do workspace.

**Convenção de build/test:** use `cd projetos/<módulo>` antes dos comandos `go`:
```bash
cd projetos/simulador && go build .
cd projetos/shared && go test ./...
```

## CLIs Go

### `assinatura`

Gerencia o `assinador.jar` (Spring Boot, porta 8080).

| Comando | Descrição |
|---------|-----------|
| `assinatura sign --content "..."` | Assina conteúdo (HTTP se servidor ativo, senão java -jar) |
| `assinatura validate --content "..." --signature "..."` | Valida assinatura |
| `assinatura start [--port 8080] [--timeout 30]` | Inicia servidor em background |
| `assinatura stop [--port 8080]` | Encerra servidor pelo PID |
| `assinatura version` | Exibe versão |

Registro: `~/.hubsaude/assinador.pid` (gravado pelo **Java**, `ServerStartupHandler`).
Health check: `GET /health`.

### `simulador`

Gerencia o `hubsaude-validador-api` (validador FHIR, Spring Boot, porta padrão **8081**).

| Comando | Descrição |
|---------|-----------|
| `simulador start [--port 8081] [--source <url>]` | Inicia validador em background |
| `simulador stop [--port 8081]` | Encerra pelo PID registrado |
| `simulador status [--port 8081]` | Consulta `/actuator/health`; reconcilia pid file órfão |
| `simulador version` | Exibe versão |

Registro: `~/.hubsaude/simulador.pid` (gravado pelo **CLI Go**, pois o jar externo não o faz).
Health/readiness: `GET /actuator/health/readiness` (Spring Actuator).
Sem endpoint de shutdown — encerramento via `proc.Kill()` no PID.

**Cold start ~20 s** (carrega 7 pacotes FHIR embutidos). `WaitUntilReady` aguarda até **90 s**.

Arquivos gerenciados em `~/.hubsaude/`:
- `simulador.jar` — jar baixado e cacheado
- `simulador.version` — versão do jar em cache (controla invalidação)
- `simulador.pid` — registro `{pid, port}` gravado pelo CLI

## Módulo `shared`

Pacotes reutilizados pelos dois CLIs:

| Pacote | Função |
|--------|--------|
| `shared/config` | `HubSaudeDir()`, `JREDir()`, `ReleaseURL` |
| `shared/release` | `Fetch()` + struct `File` (inclui `Jar` e `Simulador{URL,Version}`) |
| `shared/jre` | `JavaPath()` — detecta Java 21+ no PATH ou em `~/.hubsaude/jre` |
| `shared/process` | `Detach(cmd)` — detach multiplataforma (`Setsid` Unix / `CREATE_NEW_PROCESS_GROUP` Windows) |

## Serviço Java (`assinador-java`)

Spring Boot 3.3.5, Java 21, fat-jar via `spring-boot-maven-plugin`.

**Modos:**
- `java -jar assinador.jar sign/validate ...` — modo CLI (sem Spring)
- `java -jar assinador.jar serve [--port N]` — modo servidor HTTP

**Arquitetura em camadas:** `domain` → `application` → `presentation/{cli,http}` + `infrastructure`.
O núcleo (`domain`, `application`) é livre de framework.

**PKCS#11:** ativado por `HUBSAUDE_PKCS11_LIBRARY`; sem a variável usa `FakeSignatureService`.
**Timeout de inatividade:** controlado por `HUBSAUDE_TIMEOUT_MINUTES` (var de ambiente).

Build: `cd projetos/assinador-java && mvn package` → `target/assinador.jar`

## CI/CD

| Workflow | Gatilho | Ações |
|----------|---------|-------|
| `build.yml` | Push na `main` | Compila `assinatura` + `simulador` (3 plataformas); testes unitários `shared`, `assinatura`, `simulador`; testes de integração Go+Java |
| `release.yml` | Tag `v*` | 6 binários compilados, checksums SHA256, assinatura Cosign (`.bundle`), publicação no GitHub Releases |

Para publicar release: `git tag vX.Y.Z && git push origin vX.Y.Z`

## `release.json`

```json
{
  "jre":  { "version": "21", "windows_x64": "...", "linux_x64": "...", "mac_x64": "..." },
  "jar":  { "url": "...assinador.jar" },
  "simulador": { "version": "0.0.0", "url": "...simulador.jar" }
}
```

A `version` em `simulador` controla a invalidação de cache do jar local (`~/.hubsaude/simulador.version`).

## Convenções

- Paths específicos de cada CLI ficam em `internal/config/paths.go` de cada módulo (não em `shared/config`).
- Testes usam `t.TempDir()` + `t.Setenv("HOME", ...)` para isolamento; nunca tocam o `~/.hubsaude` real.
- `PidFilePath` e `fetchRelease` são vars exportadas/injetáveis para testes (padrão dos dois módulos).
- Testes de integração Go requerem `-tags integration` e Java instalado; testes unitários não precisam de Java.
