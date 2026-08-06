# go-expert-clima-cep-observability

Sistema distribuído em Go composto por dois microsserviços que cooperam para consultar o clima de uma cidade baseada no CEP, com observabilidade implementada via OpenTelemetry e Zipkin.

## Arquitetura

```
Client (POST /) → Serviço A (Input) → Serviço B (Orquestração) → ViaCEP / WeatherAPI
                          ↓                        ↓
                    OTEL Collector ←───────────────┘
                          ↓
                       Zipkin
```

- **Serviço A (Input):** Recebe requisições do usuário, valida o CEP e encaminha para o Serviço B.
- **Serviço B (Orquestração):** Recebe o CEP, consulta a cidade via ViaCEP, consulta a temperatura via WeatherAPI e retorna os dados formatados.
- **OTEL Collector + Zipkin:** Infraestrutura de coleta e visualização de tracing distribuído.

## Pré-requisitos

- [Docker](https://docs.docker.com/get-docker/) e [Docker Compose](https://docs.docker.com/compose/install/)
- Chave de API do [WeatherAPI](https://www.weatherapi.com/) (gratuita)

## Como executar

### 1. Defina a chave da WeatherAPI

```bash
export WEATHER_API_KEY=sua-chave-aqui
```

### 2. Suba os containers

```bash
docker compose up -d
```

### 3. Verifique se tudo está rodando

```bash
docker compose ps
```

Devem aparecer 4 containers: `zipkin`, `otel-collector`, `service-b` e `service-a`.

## Como usar

### Requisição POST para o Serviço A

```bash
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{"cep": "29902555"}'
```

#### Resposta de sucesso (200 OK)

```json
{
  "city": "São Paulo",
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.65
}
```

#### Erros

| Status | Mensagem | Quando ocorre |
|--------|----------|---------------|
| `422` | `invalid zipcode` | CEP com formato inválido (diferente de 8 dígitos) |
| `404` | `can not find zipcode` | CEP com formato correto, mas não encontrado |

## Observabilidade

### Acessando o Zipkin

Abra http://localhost:9411 no navegador para acessar a interface do Zipkin.

Clique em **"Run Query"** para visualizar os traces. Cada requisição gera o seguinte fluxo de spans:

1. **service-a** — span automático da requisição HTTP recebida
2. **service-a** — span automático do client HTTP enviado ao Serviço B (com propagação de trace context)
3. **service-b** — span automático da requisição HTTP recebida
4. **viacep-get-location** — span manual que mede o tempo de resposta da API ViaCEP
5. **weatherapi-get-temperature** — span manual que mede o tempo de resposta da WeatherAPI

### Spans manuais

Além do tracing automático das requisições HTTP, o Serviço B cria spans manuais para medir o tempo das chamadas às APIs externas:

- **viacep-get-location:** Mede o tempo da consulta de CEP na API ViaCEP. Atributos: `cep`, `city`.
- **weatherapi-get-temperature:** Mede o tempo da consulta de temperatura na WeatherAPI. Atributos: `city`, `temp_c`, `temp_f`, `temp_k`.

## Estrutura do projeto

```
.
├── docker-compose.yaml
├── otel-collector-config.yaml
├── service-a/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── main_test.go
├── service-b/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── internal/
│       ├── model/
│       │   └── model.go
│       └── service/
│           ├── service.go
│           ├── cep_test.go
│           ├── conversion_test.go
│           └── service_test.go
└── README.md
```

## Como executar os testes

### Serviço A

```bash
cd service-a && go test ./... -v
```

### Serviço B

```bash
cd service-b && go test ./... -v
```
