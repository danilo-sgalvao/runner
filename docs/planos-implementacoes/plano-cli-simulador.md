
# Plano de Implementação — Sprint 4: CLI `simulador` e Segurança Final

Plano de execução das histórias **US-03.1, US-03.2, US-03.3 e US-03.4** (mais a extensão de
US-05.3 aos novos binários). Referências:
[`plano-sprints.md`](../plano-sprints.md) (Sprint 4) e [`especificacao.md`](../especificacao.md) (US-03).

## Estado da implementação (para quem continuar)

Progresso pela [ordem sugerida](#ordem-sugerida-de-implementação) abaixo:

| Passo | Descrição | Status |
|-------|-----------|--------|
| 1 | Workspace Go + módulo `shared` + rename do `assinatura` | ✅ **concluído** |
| 2 | `release.json` + `shared/release` estendidos com `simulador` | ✅ **concluído** |
| 3 | `simulador/internal/simjar` (download dinâmico) + testes | ✅ **concluído** |
| 4 | `simulador/internal/simserver` (HTTP, `/actuator/health`, stop por PID) + testes | ✅ **concluído** |
| 5 | `simulador/cmd` (`root`, `version`, `start`, `stop`, `status`) + testes | ✅ **concluído** |
| 6 | CI/CD (`build.yml`, `release.yml`) com binários do `simulador` | ✅ **concluído** |
| 7 | Documentação (`CLAUDE.md`, `README.md`, `CONCLUSAO-SPRINT4.md`) | ✅ **concluído** |

**O que já existe no repositório (passos 1–2):**

- Workspace `projetos/go.work` com `use ./assinatura ./shared`. **Ao criar o módulo `simulador`
  no passo 5, adicione `./simulador` a esse `use`.**
- Módulo `projetos/shared` (`github.com/danilo-sgalvao/runner/shared`) com os pacotes
  `config`, `release`, `jre`, `process`. Sem dependências externas.
  - `shared/config`: `HubSaudeDir()`, `JREDir()`, `ReleaseURL` e a const exportada `DirName`
    (`.hubsaude`). Caminhos específicos de CLI **não** ficam aqui.
  - `shared/release`: `Fetch()` + struct `File`, **já com o campo `Simulador{URL, Version}`**
    (tag `json:"simulador"`). Reuse direto no `simjar` do passo 3.
- Módulo `projetos/assinatura` renomeado para `github.com/danilo-sgalvao/runner/assinatura`
  (era `.../runner`), com `require` + `replace … => ../shared`. Mantém `internal/{jar,server}`
  e ganhou `internal/config` próprio (`JarPath()`/`PidPath()` → `assinador.jar`/`assinador.pid`).
- `release.json` (raiz do repo) com a seção `simulador` (`version: "0.0.0"` como sentinela até o
  primeiro release real; `url` → `…/releases/latest/download/simulador.jar`).

**Verificado:** `go vet ./...` e `go test ./...` verdes em `shared` e `assinatura`; teste de
integração do `assinatura` compila com `-tags integration`. Nenhuma mudança de comportamento.

**Convenção para rodar comandos Go neste repo multi-módulo:** use a flag `-C` para apontar o
módulo, ex.: `go -C projetos/shared test ./...`, `go -C projetos/assinatura build ./...`. O
`go.work` é descoberto automaticamente a partir de `projetos/`.

**Pendência herdada para o passo 3 (refinada após análise do jar):** o artefato real é o
**`hubsaude-validador-api`** (validador FHIR da SES-GO), distribuído como
`hubsaude-validador-api-<versão>-exec.jar` — **não** `simulador.jar`. A `url`/`version` em
`release.json` (hoje `danilo-sgalvao/runner` + `simulador.jar`, `version: 0.0.0`) precisam
apontar para o **owner/repo externo** e para o nome de arquivo real. Isso é correção de **dados**
no `release.json`; a struct `Simulador{URL, Version}` em `shared/release` (passo 2) permanece
inalterada. O nome do cache local pode continuar `~/.hubsaude/simulador.jar` (escolha do CLI).

## Objetivo

Entregar um segundo CLI multiplataforma, `simulador`, que gerencia o ciclo de vida do
**Simulador do HubSaúde** (`simulador.jar`) — iniciar, parar e monitorar — sem que o usuário
precise conhecer os comandos Java subjacentes. Ao final da Sprint, o Sistema Runner está
completo: ambos os CLIs publicados no GitHub Releases com checksums e assinatura Cosign.

## Premissa que muda tudo: o `simulador.jar` é externo

Diferente do `assinador.jar` (construído neste repositório), o **Simulador do HubSaúde é um
artefato pronto**, baixado do GitHub Releases da disciplina. O CLI não controla o código Java
do Simulador. Isso gera diferenças estruturais em relação ao `assinatura` que **não podem ser
copiadas cegamente**.

### Contrato confirmado do jar externo (v0.1.10)

O artefato é o **`hubsaude-validador-api`** — um **validador de recursos FHIR** da SES-GO
(`br.gov.go.saude.hubsaude.validador.api`), Spring Boot 4.0.6 sobre Tomcat 11, Java 21. Fat-jar
(~177 MB) que embute 7 pacotes FHIR (`.tgz`). O contrato abaixo foi **verificado** sobre
`hubsaude-validador-api-0.1.10-exec.jar`, estática (`javap`) e ao vivo (servidor de pé), e
**corrige** as suposições iniciais deste plano (que assumiam 8443/HTTPS, `/api/info`, `/shutdown`):

| Aspecto | `assinatura` (assinador.jar) | `simulador` (hubsaude-validador-api) |
|---------|------------------------------|------------------------------|
| Origem do JAR | construído no repo (`mvn package`) | **externo**, só baixado |
| Protocolo / porta | HTTP 8080 | **HTTP 8080** (sem TLS; `--server.port=N` p/ trocar) |
| Quem grava o PID file | o **Java** (`ServerStartupHandler`) | o **próprio CLI Go**, ao spawnar |
| Parar | `proc.Kill()` pelo PID | **`proc.Kill()` pelo PID** (não há `/shutdown`) |
| Status / readiness | `GET /health` | **`GET /actuator/health/readiness`** (Spring Actuator) |
| Checar porta antes de subir | (não faz) | **verifica a porta livre antes de iniciar** |

Endpoints reais (confirmados):

- Negócio: `POST /fhir/$validate` (consumes/produces `application/fhir+json`; resposta
  `OperationOutcome` FHIR), `GET /fhir/profiles` (json), `GET /fhir/terminology-status?url=` (json).
- Actuator (exposição `health,info,prometheus`): `GET /actuator/health`
  (`{"status":"UP","groups":["liveness","readiness"]}`), `/actuator/health/liveness`,
  `/actuator/health/readiness`, `/actuator/info` (`{}`), `/actuator/prometheus`.
  **`POST /actuator/shutdown` → 404** (não exposto).
- Arg opcional para trocar os pacotes FHIR: `--packages=<lista>` (ou `-p=<lista>`).

Consequências de projeto:

- O cliente do `simulador` fala **HTTP comum** — `http.Client` padrão, **sem**
  `InsecureSkipVerify` (não há TLS). Readiness/health checados via `/actuator/health/readiness`.
- **Não há endpoint de shutdown.** O `stop` encerra **por `proc.Kill()` no PID registrado**;
  como o jar externo não escreve `~/.hubsaude/simulador.pid`, o **CLI Go grava o registro**
  `{pid, port}` logo após `cmd.Start()`.
- **Cold start ~20s** (carrega os 7 pacotes FHIR embutidos); a primeira chamada `$validate`
  dispara warm-up lazy adicional do HAPI. O `WaitUntilReady` precisa de **timeout generoso
  (≥60s)**, não os 30s do `assinatura`.
- **Colisão de porta:** o padrão 8080 é o mesmo do `assinador.jar`. Escolher outra default para
  o `simulador` (ex.: **8081**) ou documentar que não devem rodar juntos na mesma porta.

---

## Decisão estrutural: workspace Go + módulo compartilhado

O módulo atual é `github.com/danilo-sgalvao/runner`, com raiz em `projetos/assinatura`. As
regras de `internal/` do Go **proíbem** um segundo módulo (`simulador`) de importar esses
pacotes. Para reusar a infraestrutura (objetivo registrado da refatoração da Sprint 3),
adota-se um **Go workspace com um módulo compartilhado**.

### Layout final

```
projetos/
├── go.work                       # use ./shared ./assinatura ./simulador
│
├── shared/                       # módulo: github.com/danilo-sgalvao/runner/shared
│   ├── go.mod
│   ├── config/   paths.go        # HubSaudeDir, JREDir, ReleaseURL (partes genéricas)
│   ├── release/  release.go      # Fetch() + struct File (estendida com Simulador)
│   ├── jre/      manager.go      # JavaPath() — detecção/download do Java 21+
│   └── process/  detach_*.go     # detach multiplataforma
│
├── assinatura/                   # módulo: .../runner/assinatura  (RENOMEADO)
│   ├── go.mod                    # require + replace .../shared => ../shared
│   ├── cmd/                      # inalterado na lógica; só reaponta imports
│   └── internal/{jar,server}/    # específicos do assinador (ficam aqui)
│
└── simulador/                    # módulo: .../runner/simulador  (NOVO)
    ├── go.mod                    # require + replace .../shared => ../shared
    ├── main.go
    ├── cmd/      root.go, version.go, start.go, stop.go, status.go
    └── internal/
        ├── simjar/   manager.go      # Find()/download do simulador.jar via release.json
        └── simserver/ manager.go, client.go, wait.go   # PID/registro, /actuator/health, stop por PID (HTTP)
```

> **Nota sobre o module path.** `shared` não pode ser `.../runner/shared` enquanto o
> `assinatura` ocupar `.../runner` (seria interpretado como subpacote, não módulo). Por isso o
> `assinatura` é renomeado para `.../runner/assinatura`. É um refactor mecânico (reapontar
> imports `.../runner/internal/...` → `.../runner/assinatura/internal/...` e
> `.../runner/shared/...`).

> **go.work × CI.** O `go.work` serve ao desenvolvimento local. Para que o build por módulo no
> CI seja determinístico sem depender do workspace, cada `go.mod` consumidor inclui uma diretiva
> `replace ... => ../shared`. Assim `cd projetos/simulador && go build .` funciona com ou sem
> `GOWORK`.

### O que migra para `shared` (sem mudança de comportamento)

- `config/paths.go`: ficam `HubSaudeDir()`, `JREDir()`, `ReleaseURL`. Os helpers específicos do
  assinador (`JarPath()`, `PidPath()` → `assinador.jar`/`assinador.pid`) **saem do shared** e
  passam para o módulo `assinatura`. O `simulador` define os seus próprios
  (`simulador.jar`/`simulador.pid`).
- `release/release.go`: `Fetch()` e a struct `File` — esta ganha a seção `Simulador` (ver
  abaixo).
- `jre/manager.go`: reusado integralmente (ambos os CLIs precisam do Java 21+).
- `process/detach_{unix,windows}.go`: reusado integralmente.

Os testes que acompanham esses pacotes migram junto, apenas reapontando imports.

---

## Mudanças por artefato

### 1. `release.json` — nova seção `simulador`

```json
{
  "jre":  { ... inalterado ... },
  "jar":  { "url": "...assinador.jar" },
  "simulador": {
    "version": "1.0.0",
    "url": "https://github.com/<owner>/<repo>/releases/latest/download/simulador.jar"
  }
}
```

A `version` controla a invalidação de cache (US-03.4: não rebaixar se a versão local já é a
mais recente). Espelha o padrão de `internal/jar` proposto em
[`plano-download-jar-assinador.md`](./plano-download-jar-assinador.md): gravar
`~/.hubsaude/simulador.version` ao baixar e comparar antes do próximo download.

### 2. `shared/release/release.go` — struct estendida

```go
type File struct {
    JRE       struct{ ... } `json:"jre"`
    Jar       struct{ URL string `json:"url"` } `json:"jar"`
    Simulador struct {
        URL     string `json:"url"`
        Version string `json:"version"`
    } `json:"simulador"`
}
```

### 3. `simulador/internal/simjar/manager.go` — obtenção dinâmica (US-03.4)

`Find() (string, error)` espelhando `assinatura/internal/jar`:

1. Atalho de desenvolvimento: jar ao lado do executável / caminho local conhecido.
2. `~/.hubsaude/simulador.jar` presente **e** `simulador.version` == versão remota → usa cache.
3. Senão → baixa de `release.Simulador.URL`, valida (checksum quando disponível), grava o jar e
   `simulador.version`.
4. Flag **`--source <url>`** (critério de US-03.4) sobrepõe a URL do `release.json`.
5. Offline + sem cache → erro claro (não há fallback de sistema, igual ao assinador.jar).

### 4. `simulador/internal/simserver/` — ciclo de vida via HTTP

- `manager.go`: `ProcessInfo{PID, Port}`, leitura/escrita/limpeza de
  `~/.hubsaude/simulador.pid` (o **CLI** grava, não o jar). `Info(port)` faz
  `GET http://localhost:<port>/actuator/info` (e/ou `/actuator/health`); `Stop` encerra por
  **`proc.Kill()`** no PID registrado (não há `/shutdown`). Cliente **HTTP comum**, sem TLS.
- `wait.go`: `WaitUntilReady(port, timeout)` sondando `GET /actuator/health/readiness` até
  `status: UP` (análogo ao `WaitUntilReady` do assinatura, trocando `/health` →
  `/actuator/health/readiness`). **Timeout ≥60s** por causa do cold start (~20s + warm-up lazy).
- `portfree.go` (ou helper no manager): `IsPortFree(port)` via `net.Listen("tcp", :port)` —
  usado pelo `start` para checar a porta antes de subir (critério de US-03.1/US-03).

### 5. `simulador/cmd/` — comandos (US-03.1, US-03.2, US-03.3)

- `root.go` / `version.go`: espelham o `assinatura` (`Use: "simulador"`, mesmo padrão de
  `version`).
- `start.go` (US-03.1): checa porta livre → `simjar.Find()` → `jre.JavaPath()` →
  `process.Detach` + `cmd.Start()` (passando `--server.port=<porta>` ao jar) → **grava PID/porta**
  → `simserver.WaitUntilReady` (timeout ≥60s). Flag `--port` (default **8081**, para não colidir
  com o `assinador.jar` em 8080).
- `stop.go` (US-03.2): encerra por **`proc.Kill()`** no PID registrado (não há endpoint
  `/shutdown`); limpa o registro. Flag `--port`.
- `status.go` (US-03.2, **novo, sem equivalente no assinatura**): consulta
  `/actuator/health` (e/ou `/actuator/info`); imprime "em execução" + status retornado, ou
  "não está em execução". Reconcilia com o pid file (registro órfão → reporta parado e limpa).

### 6. CI/CD — `build.yml` e `release.yml` (US-03.3, US-05.3)

- **`build.yml`**: adicionar passos de compilação cross-platform do `simulador`
  (`cd projetos/simulador && GOOS=... go build`), nos mesmos 3 alvos, e subir os binários como
  artifacts. O job de teste roda `go test ./...` em `shared/`, `assinatura/` e `simulador/`.
- **`release.yml`**: compilar os 3 binários do `simulador`
  (`simulador-<tag>-<os>-<arch>`), incluí-los no `checksums.txt`, assiná-los com Cosign
  (`.bundle`) e publicá-los no mesmo release — exatamente o padrão já aplicado ao `assinatura`.
- Build per-módulo continua funcionando graças às diretivas `replace`; se necessário, fixar
  `GOWORK=off` nesses passos para isolar do workspace.

### 7. Testes

- `shared/**`: testes migrados (release com `httptest`, jre, etc.) — devem permanecer verdes só
  reapontando imports.
- `simulador/internal/simjar/manager_test.go`: download fresh, cache válido, cache
  desatualizado, `--source`, offline sem cache.
- `simulador/internal/simserver/*_test.go`: `Info`/`WaitUntilReady` contra um
  `httptest.NewServer` (HTTP comum) expondo `/actuator/health/readiness` e `/actuator/info`;
  `IsPortFree` com porta ocupada.
- `simulador/cmd/*_test.go`: `start` aborta com porta ocupada; `stop` encerra o PID registrado;
  `status` formata execução vs. parado. Usar `var` injetáveis / `PidFilePath` sobrescrevível
  (mesmo padrão de `internal/server` do assinatura) e `HOME` apontando para tempdir.

### 8. Documentação

- `CLAUDE.md`: novo subprojeto `projetos/simulador`; seção de layout do workspace/módulo
  compartilhado; tabela de comandos `start`/`stop`/`status`; remover o `simulador` de "What Is
  Not Yet Implemented".
- `README.md`: instruções de uso do `simulador` (incl. `~/.hubsaude/simulador.*` e porta 8081).
- `docs/CONCLUSAO-SPRINT4.md`: registrar entrega e decisões (espelha as conclusões anteriores).

---

## Ordem sugerida de implementação

1. ✅ **Refactor do workspace (sem mudança de comportamento):** criar `go.work`, módulo `shared`,
   mover `config`/`release`/`jre`/`process`, renomear o módulo do `assinatura` e reapontar
   imports. Rodar a suíte inteira do `assinatura` — tem de continuar 100% verde antes de seguir.
2. ✅ **`release.json` + `shared/release` estendidos** com a seção `simulador`.
3. ✅ **`simulador/internal/simjar`** (download dinâmico) + testes.
4. ✅ **`simulador/internal/simserver`** (HTTP, `/actuator/health/readiness`, stop por PID gravado
   pelo CLI, `IsPortFree`) + testes.
5. ✅ **`simulador/cmd`** (`root`, `version`, `start`, `stop`, `status`) + testes.
6. ✅ **CI/CD** (`build.yml`, `release.yml`) com os binários do `simulador`.
7. ✅ **Documentação** (`CLAUDE.md`, `README.md`, `CONCLUSAO-SPRINT4.md`).

## Mapa US → entregas

| US | Entrega principal |
|----|-------------------|
| US-03.3 | Passos 1 e 5 — estrutura do CLI espelhando o `assinatura` + comandos `start`/`stop`/`status` |
| US-03.1 | `start` (checa porta livre, baixa jar se preciso, sobe em background, feedback) |
| US-03.2 | `stop` (kill por PID) + `status` (`/actuator/health`) + registro PID/porta |
| US-03.4 | `simjar.Find()` + seção `simulador` no `release.json` + `--source` + cache por versão |
| US-05.3 | Passo 6 — checksums + Cosign para os binários do `simulador` |

## Riscos e pontos de atenção

- **Cold start ~20s + warm-up lazy:** `WaitUntilReady` precisa de timeout ≥60s. A primeira
  chamada `$validate` ainda dispara carregamento adicional do HAPI — o `start` pode reportar
  "pronto" (readiness UP) antes desse warm-up; aceitável, pois é o contrato do próprio serviço.
- **PID gravado pelo CLI, não pelo Java:** se o processo morrer fora do CLI, o pid file fica
  órfão — `status`/`stop` precisam reconciliar via `/actuator/health` e limpar o registro.
- **Sem endpoint de shutdown:** `stop` depende de `proc.Kill()` no PID. Se o PID estiver órfão/
  reciclado, validar via `/actuator/health` antes de matar para não encerrar processo alheio.
- **Colisão de porta com o `assinador.jar` (ambos 8080):** default do `simulador` em 8081 e/ou
  checagem de porta livre no `start` antes de subir.
- **Contrato preso à v0.1.10:** endpoints e exposição do actuator confirmados sobre
  `hubsaude-validador-api-0.1.10-exec.jar`. Reconfirmar se a versão de release mudar (sobretudo
  se `shutdown` passar a ser exposto ou a porta default mudar).
- **Renomeação do módulo do `assinatura`:** mecânica, porém ampla — é o passo de maior risco de
  regressão. Isolar no passo 1 e validar com a suíte completa antes de avançar.
- **`go.work` no CI:** garantir builds determinísticos via `replace` (e `GOWORK=off` se preciso),
  para o workspace não vazar dependências locais nos artefatos de release.
