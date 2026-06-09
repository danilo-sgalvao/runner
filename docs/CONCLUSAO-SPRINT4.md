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
| `internal/simserver/manager.go` | registro de PID gravado pelo **CLI**, `IsResponding`/`Health` via Actuator, `IsPortFree` |
| `internal/simserver/wait.go` | `WaitUntilReady` — polling de `/actuator/health/readiness` |
| `cmd/start.go` | `simulador start [--port] [--source]` — checa porta, obtém jar, sobe em background, aguarda readiness (US-03.1) |
| `cmd/stop.go` | `simulador stop [--port]` — encerra pelo PID registrado (US-03.2) |
| `cmd/status.go` | `simulador status [--port]` — consulta `/actuator/health`; reconcilia registro órfão (US-03.2) |
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

### D1. O "Simulador" é o validador FHIR externo `hubsaude-validador-api`

O artefato gerenciado é um serviço **externo** (Spring Boot 4.0.6, baixado pronto), não construído
neste repositório. O CLI só gerencia seu ciclo de vida; não controla o código Java. O contrato foi
**verificado** (estática via `javap` e ao vivo) sobre `hubsaude-validador-api-0.1.10-exec.jar`, o
que **corrigiu** as suposições iniciais do plano (8443/HTTPS, `/api/info`, `/shutdown`).

### D2. HTTP na porta 8081, sem TLS

O validador expõe HTTP simples (porta default 8080). O `simulador` usa **8081** por padrão para
não colidir com o `assinador.jar` (8080), passando `--server.port=N` ao jar. O cliente Go é um
`http.Client` comum — **sem** `InsecureSkipVerify`, pois não há TLS.

### D3. PID gravado pelo CLI; `stop` por `proc.Kill()`

Diferente do assinador (cujo Java grava o PID), o jar externo não escreve
`~/.hubsaude/simulador.pid` — o **CLI grava** o registro `{pid, port}` logo após `cmd.Start()`.
Como o validador **não expõe `/shutdown`**, o `stop` encerra pelo PID registrado.

### D4. Readiness via Spring Actuator com timeout generoso

A prontidão é verificada em `GET /actuator/health/readiness` (não `/health`). O **cold start é de
~20s** (carga dos pacotes FHIR embutidos) mais warm-up lazy do HAPI; por isso o `WaitUntilReady`
usa timeout de **90s**, não os 30s do `assinatura`.

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
`hubsaude-validador-api-0.1.10-exec.jar`: `start` (subiu e atingiu readiness) → `status` (UP) →
`start` novamente (detectou e reusou a instância) → `stop` (encerrou e limpou o registro) →
`status` (parado), sem processo órfão.

---

## 5. Pendências restantes

Ambas dependem de informação externa, não de implementação:

1. **Owner/repo do `simulador.jar` no `release.json`.** Hoje aponta para `danilo-sgalvao/runner`
   com `version 0.0.0` (placeholder). O artefato real é o `hubsaude-validador-api-<versão>-exec.jar`,
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
