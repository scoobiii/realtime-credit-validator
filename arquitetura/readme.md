📦 O que entregamos (vs. arquitetura correta do PDF)

Com base no código, testes e validações realizadas, entregamos os núcleos críticos da arquitetura correta, mas ainda faltam alguns componentes e integrações externas.

✅ Entregue e validado

Componente Como entregamos Evidência
Wallet Service (Ledger, Saldo, Rendimento) ✅ src/wallet com ledger ACID, UPSERT atômico, saldo, idempotência Testes k6 10 VUs: 100% sucesso, p95=11,7ms
Transaction Service (Orquestração / Saga) ✅ Embutido no handler com retry (backoff) e outbox pattern. Não é uma saga completa, mas garante atomicidade local + evento. Código AddCredit com tentativas e compensação implícita
Event Bus (Kafka / NATS) ✅ Kafka com outbox pattern, producer assíncrono, DLQ src/wallet/outbox.go, events.go
Event Store (Event Sourcing) ✅ Parcialmente: temos tabela credits imutável, mas não events com versionamento. A outbox guarda eventos, mas não é o event store completo. migrations/001_initial.sql
Banco ACID (Saldos, Contas) ✅ PostgreSQL com transações, constraints (CHECK balance >= 0), índice único Tabelas credits, user_balance
API Gateway (Autenticação, Rate Limit, WAF) ✅ Gateway Go com JWT, rate limit (Redis), health check, métricas, CORS, logs src/gateway/
Consent Service (Consents e Auditoria) ⚠️ Parcial: O token JWT carrega escopos, mas não há serviço separado de consentimento explícito para cada mutação financeira. O PDF exige ConsentService dedicado.
Pix SPI (BACEN) ❌ Simulado via webhook fake. Integração real depende de credenciais do BACEN. tests/load/anatel-latency-test.js chama /v1/credit diretamente, não via SPI
Banco Parceiro / Conta de Pagamento ❌ Não implementado (fora do escopo atual)
Anti-Fraude (ML em tempo real) ❌ Não implementado
Anatel / Reguladores (Notificações) ❌ Não implementado (apenas validação de conformidade via k6)
Notification Service ❌ Não implementado
Billing Service ❌ Separamos policy, mas não há serviço de faturamento

---

🔍 O que falta para atingir a arquitetura correta (segundo o PDF)

1. Consent Service (obrigatório para LGPD/Anatel)

· Serviço separado que gerencia consentimentos explícitos do usuário para cada ação (ex: debitar saldo, compartilhar dados).
· Auditoria de consentimento.
· Fácil de adicionar: uma nova tabela consents e uma chamada gRPC antes da transação.

2. Event Store completo (event sourcing)

· Tabela events com aggregate_id, version, payload, event_type.
· Replay de estado a partir dos eventos.
· Já temos outbox, mas não o versionamento obrigatório para projeções.

3. Transaction Service como saga distribuída

· Hoje o fluxo é local (ledger + outbox). Para uma saga real, precisamos de:
  · Orquestrador (ex: Temporal) que coordene AddCredit, ActivateEntitlement, UpdateBalance.
  · Compensação explícita (rollback de crédito se serviço falhar).

4. Integrações externas reais

· Pix SPI: Webhook assinado do BACEN (depende de cadastro).
· Anti-Fraude: Chamada a um serviço (ex: GRPC) antes de confirmar o crédito.
· Notificação: Envio de push/SMS/email após confirmação.

5. Billing Service (planos, faturas, cobrança)

· Separar lógica de planos e assinaturas do ledger.
· Consumir eventos de crédito para gerar faturas.

6. Validação formal (Lean 4)

· O PDF menciona teoremas em Lean 4. Nós fizemos testes de carga e propriedades informais, mas não uma prova formal.
· Para homologação máxima, poderíamos escrever especificações em TLA+ ou Lean 4.

7. Política de rede telecom (PCRF/PCF)

· Embora o PDF não detalhe, a arquitetura real de telecom exige integração com o core de rede (protocolos Diameter, rest API). Isso não foi implementado.

---

📊 Tabela resumo (entregue vs. pendente)

Camada (PDF) Entregue Pendente
Wallet Service ✅ -
Transaction Service (saga) ⚠️ retry local ❌ saga distribuída
Consent Service ❌ ✅ serviço + tabela
Event Bus ✅ -
Event Store ⚠️ outbox ✅ versionamento, replay
Banco ACID ✅ -
API Gateway ✅ -
Pix SPI (real) ❌ simulado ✅ integração BACEN
Banco Parceiro ❌ ✅ (se necessário)
Anti-Fraude ❌ ✅
Notificações ❌ ✅
Billing Service ❌ ✅
Validação formal ❌ ✅ Lean/TLA+
PCRF/PCF ❌ ✅ (para operadora)

---

🧭 Próximos passos recomendados

1. Adicionar Consent Service – mais simples e exige pouco esforço (1 tabela + validação no gateway).
2. Evoluir outbox para event store – adicionar coluna version e lógica de replay.
3. Implementar saga com Temporal – garantir atomicidade entre ledger e ativação de serviço.
4. Simular integrações externas (SPI, anti‑fraude, notificação) com mocks, mas com interfaces que permitam trocar pela real depois.
5. Preparar documentação de validação formal – escrever invariantes em TLA+ (ou Lean 4) e verificar com teste de modelo.
