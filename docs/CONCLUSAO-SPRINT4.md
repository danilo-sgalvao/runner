# Conclusão da Sprint 4

Resumo do estado em relação ao [`plano-sprints.md`](./plano-sprints.md) (Sprint 4) e às decisões
de arquitetura registradas. Detalhes de execução em
[`planos-implementacoes/plano-cli-simulador.md`](./planos-implementacoes/plano-cli-simulador.md).

## 1. Status por User Story

| US | Cumpridos | Pendentes |
|----|-----------|-----------|
| US-03.1 — Iniciar o Simulador via CLI | 4/4 | — |
| US-03.2 — Parar e monitorar (`stop`/`status`) | 4/4 | — |
| US-03.3 — Estrutura base do CLI `simulador` + CI/CD | 4/4 | — |
| US-03.4 — Obter `simulador.jar` dinamicamente | 4/5 | verificação de checksum do download (ver §5) |
| US-05.3 — Checksums + Cosign nos binários do `simulador` | 4/4 | — |

Sprint 4: **funcionalidade concluída.** Pendências restantes são de **dado externo** e
**fonte de integridade**, não de código (ver §5).

---

## 2. O que foi entregue

### Refatoração habilitadora (passos 1–2, pré-requisito da Sprint)

- **Go workspace** (`projetos/go.work`) com módulo **`shared`**
  (`github.com/danilo-sgalvao/runner/shared`): `config`, `release`, `jre`, `process` migrados
  para reuso pelos dois CLIs. O módulo `assinatura` foi renomeado para `.../runner/assinatura` e
  passou a ter seu próprio `internal/config`.
- `release.json` e `shared/release` estendidos com a seção `simulador` (`url`, `version`).

### CLI `simulador` (passos 3–5)

| Arquivo | Descrição |
|---------|-----------|
| `internal/config/paths.go` | caminhos `simulador.jar` / `.pid` / `.version` sob `~/.hubsaude` |
| `internal/simjar/manager.go` | `Find(source)` — download dinâmico com cache por versão, flag `--source`, fallback offline (US-03.4) |
| `internal/simserver/manager.go` | registro de PID gravado pelo **CLI**, `IsResponding`/`Probe` via HTTPS `GET /api/info` (`InsecureSkipVerify`), `IsPortFree` |
| `internal/simserver/wait.go` | `WaitUntilReady` — polling de `GET /api/info` (HTTPS) |
| `cmd/start.go` | `simulador start [--port] [--source]` — checa porta, obtém jar, sobe em background, aguarda readiness (US-03.1) |
| `cmd/stop.go` | `simulador stop [--port]` — encerra pelo PID registrado (US-03.2) |
| `cmd/status.go` | `simulador status [--port]` — consulta `GET /api/info`; reconcilia registro órfão (US-03.2) |
| `cmd/root.go`, `cmd/version.go`, `main.go` | estrutura do CLI espelhando o `assinatura` (US-03.3) |

### CI/CD (passo 6)

- `build.yml`: testes unitários (`shared` + `simulador`) + cross-compile dos binários do
  `simulador` nas 3 plataformas; `go-version` alinhado a 1.26.
- `release.yml`: compila, gera `checksums.txt`, assina com **Cosign** e publica os 3 binários do
  `simulador` (+ `.bundle`) junto com os do `assinatura`.

### Documentação (passo 7)

`CLAUDE.md`, `README.md` (seção de uso do `simulador`) e este documento atualizados; o `simulador`
saiu de "What Is Not Yet Implemented".

---

## 3. Decisões de arquitetura

### D1. O "Simulador" é o `hubsaude-simulador` externo (SMART on FHIR / OAuth2 com mTLS)

O artefato gerenciado é um serviço **externo** (Spring Boot 4, Tomcat 11, Java 21, baixado pronto),
não construído neste repositório. O CLI só gerencia seu ciclo de vida; não controla o código Java.
O contrato foi **verificado ao vivo** (2026-06-15) sobre `hubsaude-simulador-0.0.0-SNAPSHOT.jar`:
HTTPS na porta 8443 com mTLS, readiness/status via `GET /api/info` e `POST /shutdown` gracioso.

### D2. HTTPS na porta 8443 com mTLS (cliente Go com `InsecureSkipVerify`)

O `hubsaude-simulador` expõe **HTTPS na porta 8443** (= `server.port`/`server.base-url` do jar),
com certificado self-signed (keystore PKCS12 embutido) e `client-auth: want`. O `simulador` usa
**8443** por padrão (difere do `assinador.jar` em 8080), passando `--server.port=N` ao jar. O
cliente Go usa `tls.Config{InsecureSkipVerify: true}` — é probe local de ciclo de vida, não canal
de dados sensível; GETs de probe passam sem certificado de cliente.

### D3. PID gravado pelo CLI; `stop` por `proc.Kill()`

Diferente do assinador (cujo Java grava o PID), o jar externo não escreve
`~/.hubsaude/simulador.pid` — o **CLI grava** o registro `{pid, port}` logo após `cmd.Start()`.
O jar expõe `POST /shutdown` (graceful), mas o `stop` encerra pelo PID registrado por ser
independente do estado HTTP do servidor.

### D4. Readiness via `GET /api/info` (não há Actuator)

A prontidão é verificada em `GET /api/info` (200 = no ar); o jar **não tem Spring Actuator**
(`/actuator/**` responde 500, não 404). O startup é rápido (**~3s**), mas o `WaitUntilReady` usa
timeout de **60s** para dar folga ao download/preparo do JRE no primeiro start.

### D5. Download versionado do jar (US-03.4)

`simjar.Find` usa o jar ao lado do executável; senão compara `~/.hubsaude/simulador.version` com a
versão do `release.json` e só rebaixa se diferir; `--source` sobrepõe a URL; offline com cache
presente degrada para o cache. Grava em arquivo temporário e renomeia, evitando jar parcial.

### D6. Reuso via módulo `shared`, não cópia

Em vez de clonar a infraestrutura do `assinatura`, `config`/`release`/`jre`/`process` foram
extraídos para o módulo `shared`, consumido pelos dois CLIs via `replace … => ../shared` (builds
por módulo determinísticos no CI, com ou sem o workspace).

---

## 4. Resultado dos testes

| Suite | Testes | Resultado |
|-------|--------|-----------|
| Go `simulador/internal/simjar` (novos) | 6 | ✅ |
| Go `simulador/internal/simserver` (novos) | 6 | ✅ |
| Go `simulador/cmd` (novos) | 8 | ✅ |
| `shared` + `assinatura` (anteriores) | inalterados | ✅ |
| **Total novo (simulador)** | **20** | **✅** |

Além dos unitários, foi executado um **smoke test ponta-a-ponta real** contra o jar
`hubsaude-simulador-0.0.0-SNAPSHOT.jar`: `start` (subiu via HTTPS e atingiu readiness em
`/api/info`) → `status` (em execução, nome+versão) → `start` novamente (detectou e reusou a
instância) → `stop` (encerrou e limpou o registro) → `status` (parado), sem processo órfão.

---

## 5. Pendências restantes

Ambas dependem de informação externa, não de implementação:

1. **Owner/repo do `simulador.jar` no `release.json`.** Hoje aponta para `danilo-sgalvao/runner`
   com `version 0.0.0` (placeholder). O artefato real é o `hubsaude-simulador-<versão>.jar`,
   publicado em um **repositório externo da disciplina/SES**. Basta ajustar `url`/`version` —
   correção de **dados**, sem mudança de código (a struct `Simulador{URL, Version}` e a lógica do
   `simjar` permanecem). Enquanto não confirmado, o download é validável com `--source`.

2. **Verificação de checksum do download (US-03.4).** O `simjar` baixa o jar de forma atômica
   (arquivo temporário + rename), mas **não valida checksum**, porque o `release.json` ainda não
   expõe um hash para o `simulador.jar`. Quando a fonte externa publicar um checksum (ou ao
   adicioná-lo ao `release.json`), basta estender `release.File.Simulador` com o campo e comparar
   após o download — o ponto de extensão já está isolado em `simjar.download`.

---

## 6. Estado final do Sistema Runner

Com a Sprint 4, o Sistema Runner está completo conforme a especificação: dois CLIs
multiplataforma (`assinatura` e `simulador`) que executam aplicações Java do HubSaúde sem
configuração manual de Java, ambos publicados no GitHub Releases com checksums e assinatura
Cosign. Todos os épicos US-01 a US-05 têm suas histórias entregues.

---

## Adendo (2026-06-16) — Correção do jar/contrato do Simulador

O CLI `simulador` foi originalmente implementado contra o contrato do jar **errado**
(`hubsaude-validador-api` — validador FHIR, HTTP 8080, Spring Actuator). O artefato correto do
Simulador é o **`hubsaude-simulador`** (servidor SMART on FHIR / OAuth2 com mTLS): **HTTPS na
porta 8443** com certificado self-signed, readiness/status via **`GET /api/info`** (não há
Actuator) e **`POST /shutdown`** gracioso. O ciclo de vida (`start`/`stop`/`status`) foi
reapontado para esse contrato — detalhes e verificação ao vivo em
[`planos-implementacoes/plano-correcao-jar-simulador.md`](./planos-implementacoes/plano-correcao-jar-simulador.md).
A correção é contida em `internal/simserver` + ajustes em `cmd` (porta 8443, cliente TLS com
`InsecureSkipVerify`, probe `/api/info`); o `stop` segue encerrando por PID. Pendente: fixar
`release.json` com owner/repo/versão reais do `hubsaude-simulador` quando o release oficial sair.
