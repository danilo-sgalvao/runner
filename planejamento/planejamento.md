# Planejamento do Sistema Runner + entendimento atual

## 1. Visão Consolidada do Sistema

O Sistema Runner é uma camada de abstração que:

- Oculta a complexidade de execução de aplicações Java
- Oferece uma interface CLI amigável
- Gerencia:
  - runtime Java (JDK)
  - execução de aplicações (`assinador.jar`, `simulador.jar`)
  - modos de execução (local vs server)
- Centraliza a simulação de assinatura digital com validação rigorosa

O sistema é um orquestrador local de aplicações Java com foco em usabilidade, automação e padronização de execução

---

## 2. Decomposição do Sistema

### 2.1 CLI Assinatura (Go)
- Interface com o usuário
- Orquestração do assinador
- Gerenciamento de execução

### 2.2 CLI Simulador (Go)
- Gerenciamento do ciclo de vida do simulador.jar
- Download e versionamento

### 2.3 Assinador (Java)
- Validação de parâmetros
- Simulação de assinatura e validação
- Exposição via CLI e HTTP

### 2.4 Infraestrutura Local
- Gerenciamento de:
  - JDK
  - arquivos
  - processos
  - portas
  - metadados

### 2.5 Pipeline CI/CD
- Build multiplataforma
- Assinatura de artefatos
- Distribuição

---

## 3. Arquitetura Proposta

### 3.1 Estilo
- CLI Orchestrator + Service Backend

### 3.2 Camadas

#### CLI (Go)
- Parser de comandos
- Gerenciador de processos
- Cliente HTTP
- Gerenciador de ambiente

#### Serviço (Java)
- Controller HTTP
- Validação
- Serviço de assinatura

#### Infraestrutura
- Sistema de arquivos
- Gerenciamento de processos
- Configuração local

### 3.3 Decisões arquiteturais
- Separação Go (orquestração) x Java (domínio)
- Dois modos: local e server
- Simulação desacoplada via interface

---

## 4. Fluxos Principais

### 4.1 Assinatura
1. CLI recebe comando
2. Valida entrada básica
3. Decide modo (local/server)
4. Invoca assinador
5. Assinador valida parâmetros
6. Retorna resultado
7. CLI formata saída

### 4.2 Inicialização do ambiente
1. Verificar JDK
2. Baixar se necessário
3. Preparar diretórios
4. Verificar jars
5. Atualizar estado local

### 4.3 Execução em modo server
1. Verificar instância existente
2. Se não existir:
   - escolher porta
   - iniciar processo
3. Registrar PID e porta
4. Reutilizar instância

---

## 5. Decisões Técnicas Pendentes

### CLI
- Estrutura de comandos
- Flags
- Formato de saída

### Comunicação
- Payload HTTP (JSON)
- Contrato entre sistemas

### Persistência
- Arquivo JSON ou banco leve
- Estrutura de diretórios

### Logs
- Níveis
- Localização

---

## 6. Estratégia de Testes

### Unitários
- Validação (Java)
- Parsing CLI

### Integração
- CLI <-> Java
- HTTP

### End-to-end
- Fluxo completo

### Casos de erro
- Parâmetros inválidos
- Porta ocupada
- JDK ausente

---

## 7. Próximos Passos

### 1. Clarificação do domínio
- Definir parâmetros
- Definir contratos
- Refinar requisitos

### 2. Design da CLI
- Estrutura de comandos
- UX e mensagens

### 3. Design da API
- Endpoints
- Request/response
- Erros

### 4. Estruturação dos projetos
- Organização Go
- Organização Java

### 5. Implementação base
- CLI funcional
- API mínima

### 6. Implementação incremental
- Validação
- Execução local e HTTP
- Gerenciamento de processos

### 7. Infra local
- Diretórios
- JDK
- Estado

### 8. Simulador
- Download
- Execução

### 9. Testes
- Unitários
- Integração

### 10. Distribuição
- Build
- Releases
- Cosign

### 11. Documentação
- Guia de uso
- Integração
- Exemplos

---
O projeto vai além da implementação de código, envolvendo:

- Integração de sistemas
- Experiência do usuário via CLI
- Orquestração de runtime
- Engenharia de distribuição

