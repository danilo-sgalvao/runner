# Conclusão da Sprint 3

Resumo do estado atual em relação ao [`plano-sprints.md`](./plano-sprints.md) (Sprint 3) e decisões de arquitetura registradas.

## 1. Status por User Story

| US | Cumpridos | Pendentes |
|----|-----------|-----------|
| US-02.4 — Endpoints HTTP `/sign` e `/validate` | 4/4 | — |
| US-02.5 — Integração PKCS#11 | 4/4 | — |
| US-01.5 — CLI inicia servidor em background | 4/4 | — |
| US-01.6 — CLI invoca via HTTP, fallback local | 4/4 | — |
| US-01.7 — Detecta e reutiliza instância ativa | 4/4 | — |
| US-01.8 — Comando `assinatura stop` | 4/4 | — |
| US-01.9 — Parâmetro `--timeout` por inatividade | 3/3 | — |

Sprint 3: **100% concluída.** Todos os critérios de aceitação marcados como `[x]` em `plano-sprints.md`.

---

## 2. O que foi entregue

### Lado Java (`projetos/assinador-java`)

| Arquivo | Descrição |
|---------|-----------|
| `presentation/http/HealthController.java` | `GET /health` → `{"status":"UP"}`, usado pelo CLI para detectar startup |
| `infrastructure/pkcs11/Pkcs11Config.java` | Lê `HUBSAUDE_PKCS11_LIBRARY`, `HUBSAUDE_PKCS11_NAME`, `HUBSAUDE_PKCS11_PIN` |
| `infrastructure/pkcs11/Pkcs11ServiceFactory.java` | Configura `SunPKCS11` e abre o `KeyStore` do dispositivo |
| `domain/service/Pkcs11SignatureService.java` | Assina com RSA via chave privada PKCS#11; valida tentando todos os certificados |
| `infrastructure/config/AppConfig.java` | Usa PKCS#11 quando configurado; cai para `FakeSignatureService` com aviso |
| `infrastructure/RequestTimestamp.java` | `AtomicLong` com o instante da última requisição HTTP |
| `infrastructure/InactivityFilter.java` | Chama `touch()` a cada requisição recebida |
| `infrastructure/InactivityShutdown.java` | Lê `HUBSAUDE_TIMEOUT_MINUTES` e encerra via `System.exit(0)` após inatividade |

### Lado Go (`projetos/assinatura`)

| Arquivo | Descrição |
|---------|-----------|
| `internal/process/detach_unix.go` | `Setsid: true` — processo sobrevive ao encerramento do CLI no Linux/macOS |
| `internal/process/detach_windows.go` | `CREATE_NEW_PROCESS_GROUP` — equivalente no Windows |
| `internal/server/manager.go` | Lê/limpa `~/.hubsaude/assinador.pid`; `IsResponding` faz health check HTTP |
| `internal/server/client.go` | `Sign()` e `Validate()` via `POST /sign` e `POST /validate` |
| `cmd/start.go` | `assinatura start [--port] [--timeout]` — inicia servidor em background |
| `cmd/stop.go` | `assinatura stop [--port]` — encerra processo pelo PID registrado |
| `cmd/sign.go` | Rota HTTP quando servidor ativo; fallback `java -jar`; flag `--local` |
| `cmd/validate.go` | Idem para validação |

---

## 3. Decisões de arquitetura

### D1. Health check como contrato de startup

O CLI Go não confia no arquivo de PID para confirmar que o servidor está pronto — ele sonda `/health` ativamente por até 30 segundos após o `Start()`. O arquivo de PID é responsabilidade do Java (`ServerStartupHandler`) e só é lido para recuperar PID e porta nas operações subsequentes (`stop`, `sign`, `validate`).

**Por quê:** o PID pode ser gravado antes do Tomcat estar pronto para aceitar conexões. Usar `/health` garante que o CLI só devolve controle ao usuário quando o servidor já aceita requisições.

### D2. Fallback automático sem flag obrigatória

`sign` e `validate` tentam HTTP primeiro e caem para `java -jar` silenciosamente se o servidor não estiver respondendo, sem exigir nenhuma configuração do usuário. A flag `--local` existe para forçar o modo local explicitamente.

**Por quê:** o objetivo da Sprint 3 é reduzir latência quando o servidor está ativo, não mudar o comportamento quando ele não está. O usuário que não rodou `start` não deve perceber diferença.

### D3. PKCS#11 é opt-in por variável de ambiente

A integração com PKCS#11 só é ativada quando `HUBSAUDE_PKCS11_LIBRARY` está definida. Ausente a variável, o sistema usa `FakeSignatureService` sem mensagem ao usuário (comportamento idêntico às Sprints anteriores).

**Por quê:** hardware criptográfico não está disponível em todas as máquinas de desenvolvimento. Forçar a detecção na inicialização quebraria o fluxo padrão.

### D4. Timeout por variável de ambiente, não por argumento Java

O CLI passa `HUBSAUDE_TIMEOUT_MINUTES` como variável de ambiente ao iniciar o subprocesso Java, em vez de usar um argumento de linha de comando. O Java lê a variável em `InactivityShutdown`.

**Por quê:** o argumento `serve` já é parseado por `AssinadorApplication` com uma estrutura simples (`for` + `--port`). Adicionar `--timeout` ali sem um framework de parsing de args criaria código frágil. Variável de ambiente é mais limpa e não exige mudança no parser existente.

### D5. Detach multiplataforma em arquivos separados por build tag

`internal/process/detach_unix.go` e `detach_windows.go` usam build tags (`//go:build`) em vez de `runtime.GOOS` em tempo de execução, porque `syscall.SysProcAttr` tem campos diferentes por plataforma e não compila em arquivo único.

---

## 4. Resultado dos testes

| Suite | Testes | Resultado |
|-------|--------|-----------|
| Java — domínio + aplicação (anteriores) | 22 | ✅ |
| Java — integração HTTP (anteriores) | 7 | ✅ |
| Java — PKCS#11 (novos) | 5 | ✅ |
| Java — health no smoke test (novo) | 1 | ✅ |
| Go cmd — anteriores | 17 | ✅ |
| Go cmd — start (novos) | 6 | ✅ |
| Go cmd — stop (novos) | 5 | ✅ |
| Go server — manager (novos) | 4 | ✅ |
| Go server — client (novos) | 3 | ✅ |
| **Total** | **70** | **✅** |

---

## 5. O que a Sprint 3 deixa pronto para a Sprint 4

- O ciclo de vida do `assinador.jar` está totalmente gerenciado pelo CLI (`start`, `stop`, detecção automática) — a mesma estrutura será replicada para o `simulador.jar` na Sprint 4.
- O pacote `internal/server` é genérico o suficiente para ser reutilizado (ou adaptado) no CLI `simulador`.
- O mecanismo de `HUBSAUDE_TIMEOUT_MINUTES` pode ser aplicado diretamente ao simulador sem alteração.
- A Sprint 4 começa com a infraestrutura de servidor HTTP e gerenciamento de processos já consolidada e testada.
