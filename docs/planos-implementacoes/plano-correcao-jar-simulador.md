# Plano de Correção — Reapontar o CLI `simulador` para o jar correto (`hubsaude-simulador`)

> **Status:** ✅ implementado (2026-06-16) — correção aplicada ao CLI (`internal/simserver`
> + `cmd/{root,start,status,stop}`) e suíte do módulo `simulador` verde; validado ponta-a-ponta
> contra o jar real. As seções abaixo descrevem o plano que foi seguido.
> **Origem:** o CLI `projetos/simulador` foi implementado contra o contrato do jar **errado**
> (`hubsaude-validador-api`). O jar correto do Simulador é
> `hubsaude-simulador-<versão>.jar` (`Start-Class br.gov.go.saude.hubsaude.simulador.SimuladorApplication`).
> **Supersede** a seção "Contrato confirmado do jar externo (v0.1.10)" de
> [`plano-cli-simulador.md`](./plano-cli-simulador.md), que descreve o validador-api.

---

## 1. Contexto: existem dois jars distintos

| Jar | O que é | Pacote / Start-Class |
|-----|---------|----------------------|
| `hubsaude-validador-api-0.1.10-exec.jar` (~177 MB) | **Validador FHIR** da SES-GO (HAPI). Controllers `Validate/Profile/TerminologyStatus`. HTTP 8080, Spring Actuator, sem `/shutdown`. | `br.gov.go.saude.hubsaude.validador.api` |
| `hubsaude-simulador-0.0.0-SNAPSHOT.jar` (~92 MB) | **O Simulador correto.** Servidor de autorização **SMART on FHIR / OAuth2 com mTLS**. | `br.gov.go.saude.hubsaude.simulador.SimuladorApplication` |

O CLI atual (`simserver`, `cmd/start|stop|status`, `cmd/root`) embute as premissas do **validador-api**
(HTTP 8080→8081, `/actuator/health[/readiness]`, sem shutdown). Apontado para o jar correto, o ciclo de
vida **não funciona**: `start` sempre estoura timeout e `status` sempre reporta "parado", porque (a) o
servidor é HTTPS e o cliente Go usa `http://`, e (b) os endpoints `/actuator/**` não existem.

## 2. Contrato real do jar correto (verificado ao vivo em 2026-06-15)

Subido com `java -jar hubsaude-simulador-0.0.0-SNAPSHOT.jar --server.port=18443` e sondado com `curl`:

- **HTTPS + mTLS na porta configurada.** `application.yml`: `server.port: 8443`,
  `server.ssl.enabled: true`, keystore PKCS12 embutido (`classpath:keystore/simulador.p12`,
  senha `simulador`, alias `simulador`), `server.ssl.client-auth: want`,
  `server.base-url: https://localhost:8443`. Spring Boot 4.0.1, Tomcat 11, Java 21.
- **Certificado self-signed** → o cliente Go precisa de `tls.Config{InsecureSkipVerify: true}`.
- **`client-auth: want` (não `need`)** + warning observado no log
  (`JSSE TLS 1.3 ... does not support post handshake authentication ... incompatible with optional
  certificate authentication`) → **GETs sem certificado de cliente passam**. Confirmado: `/api/info` → 200.
- **Não há Spring Actuator** (0 entradas no jar; `BOOT-INF/lib` sem `spring-boot-actuator`).
  `GET /actuator/**` retorna **HTTP 500** (capturado pelo `GlobalExceptionHandler`, não 404).
  > **Atualização 2026-06-23:** isto vale para o `0.0.0-SNAPSHOT` analisado aqui. No `0.1.11` o jar
  > **passou a incluir o Actuator** (`/actuator/health` → 200 UP). A escolha de sondar `/api/info`
  > permanece (probe estável entre versões); o CLI não muda.
- **Startup rápido (~3,2 s)** — não há carga pesada de pacotes FHIR. O timeout de 90 s e o texto
  "~1 min" do `start` atual estão muito superdimensionados.
- **`POST /shutdown` existe** (`ShutdownController`): retorna 200
  `{"message":"Shutdown iniciado..."}` e faz **graceful shutdown** (processo encerra com exit 0,
  após ~0,5 s de delay interno).
- **O jar não grava PID file** — o CLI Go continua gravando `~/.hubsaude/simulador.pid` (igual hoje).

### Endpoints (probe de liveness recomendado em **negrito**)

| Método | Rota | Resposta | Uso no CLI |
|--------|------|----------|------------|
| GET | **`/api/info`** | 200 `{"version","name"}` | **readiness / status / IsResponding** |
| GET | `/.well-known/smart-configuration` | 200 (estático) | alternativa de probe |
| GET | `/metadata`, `/cds-services` | 200 | — |
| GET | `/certs` (JWKS) + POST `/rotate` | 200 | — |
| POST | `/auth/token` | mTLS via `MtlsTokenEndpointFilter` | — |
| POST | `/fhir/**`, GET `/fhir/{rt}/{id}` | FhirFakeController | — |
| POST | **`/shutdown`** | 200, graceful | **stop (recomendado)** |

## 3. Diferença entre o que o código assume e o jar real

| Aspecto | Código atual (`simserver`/`cmd`) | Jar correto | Ação |
|---------|----------------------------------|-------------|------|
| Esquema HTTP | `http://` (`actuatorURL`) | **HTTPS self-signed** | cliente TLS + `InsecureSkipVerify` |
| Porta padrão | `defaultPort = 8081` (`cmd/root.go`) | **8443** (= `server.port`/`base-url`) | `defaultPort = 8443` |
| Probe readiness | `GET /actuator/health/readiness` (`wait.go`) | inexistente (500) | `GET /api/info` → 200 |
| Probe health/status | `GET /actuator/health` (`manager.go`) | inexistente (500) | `GET /api/info` |
| Modelo de resposta | `HealthStatus{Status}` (`{status:UP}`) | `{version,name}` | trocar por `Info{Version,Name}` |
| Timeout de readiness | 90 s (`cmd/start.go`) | start ~3 s | reduzir p/ 60 s (folga p/ download de JRE) |
| Parar | `proc.Kill()` por PID (`cmd/stop.go`) | `POST /shutdown` (graceful) **+ kill fallback** | adicionar shutdown gracioso |
| `IsPortFree` | `net.Listen` TCP | independe de TLS | **sem mudança** |
| PID gravado pelo CLI | sim | jar não grava | **sem mudança** |
| `--server.port=N` ao jar | passado em `start.go` | aceito (Spring padrão) | **sem mudança** |

## 4. Mudanças por arquivo

Escopo concentrado em **um pacote** (`internal/simserver`) + 3 ajustes em `cmd` + fixtures de teste.
**Nenhuma mudança arquitetural.** `simjar`, `internal/config`, `shared/{jre,release,process}` **não mudam**.

### 4.1 `internal/simserver/manager.go` — cliente HTTPS e probe `/api/info`

- Adicionar `crypto/tls` e um cliente compartilhado com `InsecureSkipVerify`:

```go
// O simulador serve HTTPS com certificado self-signed (keystore p12 embutido) e
// client-auth: want (não exige cert de cliente para GETs). Pulamos a verificação da
// cadeia porque o objetivo é apenas gerência de ciclo de vida local.
var httpClient = &http.Client{
	Timeout: httpTimeout,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	},
}
```

- Renomear `actuatorURL` → `baseURL` e trocar o esquema para `https`:

```go
func baseURL(port int, path string) string {
	return fmt.Sprintf("https://%s:%d%s", dialHost, port, path)
}
```

- Substituir `HealthStatus`/`Health` por `Info`/`Probe` (rota `/api/info`):

```go
// Info reflete o corpo de GET /api/info do simulador.
type Info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Probe consulta GET /api/info; 200 significa simulador no ar.
func Probe(port int) (*Info, error) {
	resp, err := httpClient.Get(baseURL(port, "/api/info"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status inesperado %d em /api/info", resp.StatusCode)
	}
	var info Info
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("resposta de /api/info inválida: %w", err)
	}
	return &info, nil
}
```

- `IsResponding(port)` passa a usar `httpClient` e `baseURL(.../api/info)` (mantém a semântica
  "qualquer round-trip HTTP = há servidor"). **Atenção:** com o cliente plain `http` atual contra
  porta TLS, hoje daria erro; com o cliente TLS, passa a funcionar.
- `IsPortFree`, `WriteProcessInfo`, `ReadProcessInfo`, `ClearProcessInfo`, `pidFile`: **inalterados**.
- Atualizar o comentário de cabeçalho do pacote (hoje cita "hubsaude-validador-api", "Actuator",
  "não há shutdown") para o contrato novo.

### 4.2 `internal/simserver/wait.go` — readiness via `/api/info`

- Sondar `GET /api/info` até **200** (em vez de `/actuator/health/readiness`), reusando `httpClient`
  (remover o `&http.Client{}` local, que não tem o transport TLS):

```go
func WaitUntilReady(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := baseURL(port, "/api/info")
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		resp, err := httpClient.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
	return fmt.Errorf("simulador não ficou pronto (/api/info) após %s", timeout)
}
```

- Atualizar o comentário (não há cold start de 20 s; o serviço sobe em ~3 s).

### 4.3 `cmd/root.go` — porta padrão 8443

```go
// 8443 = porta/`base-url` própria do simulador.jar (HTTPS); difere do assinador.jar (8080).
const defaultPort = 8443
```

> **Por que 8443 e não outra:** o `application.yml` fixa `server.base-url: https://localhost:8443`.
> Os documentos de discovery (`/.well-known/smart-configuration`, `/metadata`) emitem esse `baseUrl`
> independentemente de `--server.port`. Mantendo o default em 8443, as URLs de discovery ficam
> coerentes. Ainda evita colisão com o assinador (8080).

### 4.4 `cmd/start.go` — textos e timeout

- `WaitUntilReady(startPort, 60*time.Second)` (era 90 s).
- Ajustar as mensagens "pode levar ~1 min no primeiro start" → "alguns segundos" (o cold start
  pesado era do validador-api). O download do JRE no primeiro uso justifica manter alguma folga.
- A linha `exec.Command(javaPath, "-jar", jarPath, fmt.Sprintf("--server.port=%d", startPort))`
  **permanece** (Spring aceita). *Opcional:* quando `--port` divergir de 8443, passar também
  `--server.base-url=https://localhost:<porta>` para manter o discovery coerente (o
  `DiscoveryController` lê `server.base-url`). Recomendado só se houver caso de uso de porta custom.

### 4.5 `cmd/stop.go` — shutdown gracioso com fallback por PID (recomendado)

> ⚠️ **EM ABERTO (não implementado).** Na correção de 2026-06-16 foi adotada a **"alternativa
> mínima"** abaixo: o `stop` continua encerrando por `proc.Kill()` no PID. O `POST /shutdown`
> gracioso **recomendado** nesta seção **ainda não foi feito** — depende de uma decisão (ver
> US-03 na especificação, que cita o endpoint `/shutdown`). Para retomar: implementar os helpers
> `RequestShutdown`/`WaitUntilDown` em `simserver` e o fluxo de `stop.go` esboçados aqui, mais os
> testes do §5/§10.

- **Recomendado:** tentar `POST /shutdown` primeiro (contrato do próprio jar) e, em falha/instância
  não responsiva, cair para `proc.Kill()` no PID registrado. Esboço:

```go
info, err := simserver.ReadProcessInfo() // como hoje
...
if err := simserver.RequestShutdown(info.Port); err == nil {
	// aguarda o processo sumir (graceful); se não sumir no prazo, mata por PID
	if simserver.WaitUntilDown(info.Port, 10*time.Second) == nil {
		simserver.ClearProcessInfo()
		fmt.Printf("Simulador encerrado via /shutdown (porta %d, PID %d)\n", info.Port, info.PID)
		return nil
	}
}
proc, _ := os.FindProcess(info.PID) // fallback kill por PID (caminho atual)
_ = proc.Kill()
simserver.ClearProcessInfo()
```

  Novos helpers em `simserver`: `RequestShutdown(port)` (`POST /shutdown`, ok se 200) e
  `WaitUntilDown(port, timeout)` (poll `/api/info` até erro de conexão).
- **Alternativa mínima:** manter `proc.Kill()` por PID como está (continua funcionando). Nesse caso
  `stop.go` não muda. O ganho do `/shutdown` é encerramento limpo do contexto Spring (libera a porta
  via graceful shutdown do Tomcat) e robustez quando o PID estiver órfão/reciclado.

### 4.6 `cmd/status.go` — imprimir a partir de `/api/info`

- Trocar `simserver.Health(port)` por `simserver.Probe(port)` e imprimir `version`/`name`
  em vez de `health.Status`:

```go
info, err := simserver.Probe(port)
if err != nil {
	fmt.Printf("Simulador não está em execução na porta %d\n", port)
	... // reconciliação de registro órfão permanece
	return nil
}
fmt.Printf("Simulador em execução na porta %d (PID %d) — %s %s\n",
	port, pidInfo.PID, info.Name, info.Version)
```

## 5. Impacto nos testes

| Arquivo | Mudança |
|---------|---------|
| `internal/simserver/manager_test.go` | `serverPort` helper: `httptest.NewServer` → **`httptest.NewTLSServer`** (o cliente passa a falar TLS; `InsecureSkipVerify` aceita o cert do httptest). `dialHost = "127.0.0.1"` permanece. `TestHealth` → `TestProbe` (handler em `/api/info` devolvendo `{"version","name"}`; assertivas nos campos). `TestIsResponding`/`TestWaitUntilReady_*`: handlers no caminho `/api/info`; lógica idêntica. |
| `cmd/commands_test.go` | `TestStartFlags` lê `defaultPort` (passa automaticamente com 8443). `TestProcessInfo_RoundTrip`/`TestStatus_NaoEmExecucao`/`TestStop_*` não tocam o jar — continuam válidos; atualizar literais de porta (8081→8443) por clareza. Se `stop.go` ganhar `/shutdown`, cobrir o fallback por PID e (opcional) o caminho `/shutdown` com `httptest.NewTLSServer`. |
| (novo, opcional) | Teste de `RequestShutdown`/`WaitUntilDown` contra `httptest.NewTLSServer` se 4.5 for adotado. |

Rodar por módulo: `go -C projetos/simulador test ./...` e `go -C projetos/simulador vet ./...`.

## 6. Impacto em documentação e dados (fora do código Go)

| Artefato | Mudança necessária |
|----------|--------------------|
| `runner/CLAUDE.md` | Seção "Simulador CLI architecture": porta 8081→**8443 HTTPS+mTLS**, `/actuator/health*`→**`/api/info`**, acrescentar `POST /shutdown`, nota de TLS self-signed/`InsecureSkipVerify`. Atualizar a linha da tabela "Key Files" de `internal/simserver`. Em "What Is Not Yet Implemented", corrigir: o artefato é **`hubsaude-simulador`** (SMART/OAuth2 mTLS), não `hubsaude-validador-api`. |
| `docs/planos-implementacoes/plano-cli-simulador.md` | A seção "Contrato confirmado do jar externo (v0.1.10)" descreve o **jar errado**. Adicionar banner no topo apontando para este plano como correção, ou marcar a seção como obsoleta. |
| `docs/CONCLUSAO-SPRINT4.md` | Adendo registrando a correção do jar/contrato (se a Sprint 4 foi concluída sobre o contrato antigo). |
| `release.json` | **FEITO (2026-06-23):** `simulador` reapontado para o artefato real em **`kyriosdata/runner`** (tag `hubsaude-simulador-v0.1.11`, URL fixada por tag, `version 0.1.11`) e acrescido de `sha256` (do `checksums.txt` do release). A struct ganhou o campo `SHA256` e `simjar.download` agora confere o hash (US-03.4). Falta apenas **publicar em `danilo-sgalvao/runner@main`** (origem lida em runtime). |
| `README.md` (se existir) | Referências de porta/uso do `simulador` (8443). |
| Memória do agente | `project_simulador_jar_contract.md` já atualizada com o contrato correto. |

## 7. O que **não** é afetado (reuso preservado)

- `internal/simjar/manager.go` e testes — download + cache por versão independem do contrato HTTP.
- `internal/config/paths.go` — caminhos `simulador.{jar,pid,version}` inalterados.
- `shared/{config,release,jre,process}` — inalterados (o reuso da Sprint 4 continua válido).
- Fluxo de PID gravado pelo CLI; `IsPortFree`; `process.Detach`; bootstrap do JRE.
- CI/CD (`build.yml`/`release.yml`) — compila/empacota igual; só muda *dados* do `release.json`.

## 8. Riscos e pontos em aberto

- **Cert self-signed → `InsecureSkipVerify`.** Aceitável para gerência local de ciclo de vida (probe
  a `localhost`). Documentar a decisão (não é canal de dados sensível; é controle de processo local).
- **mTLS no `/auth/token`.** O `MtlsTokenEndpointFilter` impõe cert de cliente **apenas** no token
  endpoint; o CLI não chama `/auth/token`, então não precisa de keystore de cliente. Confirmado que
  GETs de probe passam sem cert.
- **`base-url` fixo em 8443.** Com `--port` diferente, os documentos de discovery emitem URLs com 8443
  (inconsistentes). Mitigação: default 8443; se precisar de porta custom, passar `--server.base-url`
  (4.4). Não afeta a gerência de ciclo de vida.
- **Contrato preso a esta build.** Verificado sobre `hubsaude-simulador-0.0.0-SNAPSHOT.jar`
  (build 2026-03-09). **Reconfirmar** porta/`/api/info`/`/shutdown` quando sair o release oficial e
  ao fixar `release.json` (owner/repo/nome reais).
- **`POST /shutdown` tem delay interno (~0,5 s).** Se adotar 4.5, `WaitUntilDown` deve dar folga
  (≥5–10 s) antes de cair para `proc.Kill()`.
- **Porta 8443 ocupada na máquina do usuário.** `IsPortFree` no `start` já trata (aborta com mensagem
  ou orienta `--port`).

## 9. Plano de validação (após implementar)

1. `go -C projetos/simulador vet ./...` + `go -C projetos/simulador test ./...` verdes.
2. Build local: `go -C projetos/simulador build -o simulador.exe .`
3. Ponta-a-ponta com o jar real (colocar `simulador.jar` ao lado do binário ou usar `--source`):
   - `.\simulador.exe start` → sobe em ~poucos segundos, grava `~/.hubsaude/simulador.pid`, reporta pronto.
   - `.\simulador.exe status` → "em execução ... HubSaúde Simulador 0.0.0-SNAPSHOT".
   - `.\simulador.exe stop` → encerra (via `/shutdown` se 4.5; senão kill por PID); `status` reporta parado.
   - `start` duas vezes → segunda detecta instância existente (reuso) ou aborta por porta ocupada.
4. Conferir que `status` numa porta sem servidor reporta "não está em execução" sem travar (timeout TLS curto).

## 10. Ordem de implementação sugerida

1. `internal/simserver/manager.go` (cliente TLS + `baseURL` + `Info`/`Probe`) e `wait.go` (`/api/info`).
2. Ajustar testes de `simserver` (`NewTLSServer`, `TestProbe`).
3. `cmd/root.go` (`defaultPort = 8443`), `cmd/start.go` (timeout/textos), `cmd/status.go` (`Probe`).
4. (Recomendado) `cmd/stop.go` + helpers `RequestShutdown`/`WaitUntilDown` + testes.
5. Rodar a suíte do módulo `simulador` (verde) e `vet`.
6. Atualizar docs/dados: `CLAUDE.md`, banner em `plano-cli-simulador.md`, `release.json` (quando os
   dados reais estiverem disponíveis), `CONCLUSAO-SPRINT4.md`.
7. Validação ponta-a-ponta da seção 9 com o jar real.

---

**Esforço estimado:** moderado e contido (~1 sessão focada). A maior parte é a troca do contrato
HTTP/health em `simserver` e a migração das fixtures de teste para `httptest.NewTLSServer`; o resto
são ajustes pontuais (uma constante de porta, textos, e — opcional — o shutdown gracioso).
