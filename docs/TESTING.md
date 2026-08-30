# Testes e análise estática

## Testes unitários

```bash
make test        # go test ./... -cover (rápido, sem Docker)
```

Os testes unitários (`internal/handlers`, `internal/usecases`, `internal/jwt`, `internal/config`, `internal/validators`) usam apenas a biblioteca padrão, com dublês escritos à mão para o banco e o gerador de token. Entre eles, cobrem os quatro caminhos de resposta da Lambda (200/400/401/500).

## Testes de integração

A camada de repositório (`internal/repositories/customers`) é testada contra um Postgres real, subido e destruído automaticamente pelo [Testcontainers](https://testcontainers.com/):

```bash
make test-integration   # exige Docker em execução
```

O container usa a mesma imagem do compose (`postgres:16`) e carrega o `scripts/seed.sql` como script de inicialização. O schema e as fixtures são exatamente os do ambiente local, sem duplicação de DDL. Um único container é criado por execução do pacote (via `TestMain`) e encerrado ao final.

Cenários cobertos em `GetCustomerByCpfCnpj`:

| Cenário | Esperado |
|---|---|
| CPF de cliente ativo | cliente com `Active: true` |
| CNPJ de cliente ativo | cliente encontrado (valida a coluna `varchar(14)` cheia) |
| Cliente inativo | retorna o cliente com `Active: false`, sem erro — a regra de negócio fica no usecase |
| CPF inexistente | `usecases.ErrCustomerNotFound` |
| String vazia | `usecases.ErrCustomerNotFound` |
| Contexto cancelado | erro de infraestrutura, distinto de `ErrCustomerNotFound` |

Esses testes ficam atrás da build tag `integration`, então `make test` continua rápido e sem Docker durante o desenvolvimento. Não é preciso subir o `make dev-up`: o container de teste é independente do Postgres de desenvolvimento.

## Cobertura

```bash
make test-race   # roda tudo (unitário + integração) com -race e gera coverage.txt
make cover-check # test-race + gate de cobertura (mínimo 70%)
```

`test-race` e `cover-check` usam `-tags=integration`, de modo que a cobertura medida localmente é a mesma reportada pelo CI e pelo SonarQube, sem divergência entre os ambientes. Em contrapartida, **exigem Docker**, e o `make pre-commit` depende do `cover-check`.

O cálculo exclui `/cmd/` e `/internal/models/`, por serem, respectivamente, wiring de inicialização e structs sem lógica. `internal/database` permanece sem cobertura: é o único ponto que ainda depende de conexão externa não testada.

## Análise estática (SonarQube)

O SonarQube roda localmente em container próprio (`compose.sonar.yaml`), separado do ambiente de desenvolvimento:

```bash
make sonar-up    # sobe o SonarQube em http://localhost:9000 (compose.sonar.yaml)
make sonar       # roda os testes com cobertura e envia a análise ao SonarQube
make sonar-down  # derruba o SonarQube
```

Na primeira execução, acesse `http://localhost:9000` (login inicial `admin`/`admin`), troque a senha e gere um token em *My Account > Security*. Grave-o como `SONAR_TOKEN` no `.env`, o `make sonar` falha com uma mensagem de orientação se a variável estiver ausente. Use `make sonar-logs` para acompanhar a subida do container, que leva cerca de um minuto.

`make sonar` depende de `cover-check`, então reaproveita o `coverage.txt` gerado pelo `go test -covermode=atomic` e envia a cobertura junto com o código. O scanner roda via Docker (`sonarsource/sonar-scanner-cli`), sem precisar instalar nada localmente, e lê `sonar-project.properties`, onde os arquivos `*_test.go` são classificados como testes e `volume/`/`scripts/` ficam fora da análise.

## Verificação completa

```bash
make pre-commit  # tidy, fmt-check, vet, lint, test-race, cover-check, sec
```

Cobre unitários e integração de uma vez, com o mesmo gate de cobertura do CI. Exige Docker em execução.