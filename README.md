# Rate Limiter

Rate Limiter feito para a Pós Graduação em Go da Full Cycle.

## Instalação

```bash
make prepare
``` 
E após isso
```
make run
```

## Testes

Para testar:
```
go test ./...
```
Ou então para testar por IP:
```
docker compose run --rm go-cli-test -url http://go-app:8080/hello-world -method GET -time 1 -req 10
```
Ou por API KEY fazer a request na url http://localhost:8080/api-key via Postman:
```
{
    "time_window": 1,
    "max_req": 10,
    "blocked_duration": 60
  }
```
Pegar a resposta da requisição e colocar na chave -key, no teste a seguir:
```
docker compose run \                                                                               
  --rm \
  go-cli-test \
  -url http://go-app:8080/hello-world-key \
  -method GET \
  -time 1 \
  -req 10 \
  -key 1a76f442d009951566881afcb24cca5ed5e39b9df187cdad298ad8ce62901b26
```
## Utilização por API KEY

Para utilizar é só fazer as requisições via Postman:

Para criar uma API KEY, fazer uma requisição POST em http://localhost:8080/generate-api-key:
```
{
  "time_window": 1,
  "max_req": 100,
  "blocked_duration": 60
}
```
Ou via arquivo .http no vscode
```
POST http://localhost:8080/generate-api-key
Content-Type: application/json

{
  "time_window": 1,
  "max_req": 100,
  "blocked_duration": 60
}
###
```

E depois uma requisição GET na url http://localhost:8080/req-by-key, ir na aba Authorization, colocar o Type API key e adicionar a Key: api_key e o Value a chave retornada da requisição POST.

Ou então via arquivo .http no vscode substituindo o valor da API_KEY pelo valor obtido na requisição POST:
```
GET http://localhost:8080/req-by-key
Content-Type: application/json
API_KEY: 03527ee760fda29b013c01937eb19b2508a31a8547413ea74b51474bef87d576
###
```

# Utilização por IP

Fazer uma requisição GET via Postman:
```
http://localhost:8080/req-by-ip
```
Ou via .http no vscode:
```
GET http://localhost:8080/req-by-ip
Content-Type: application/json
###
```
