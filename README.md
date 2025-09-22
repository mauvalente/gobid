# GO BID

- Modulo 5

Uma APP para gerenciar Leilões, onde usuarios autenticados podem adicionar produtos para venda e fazer propostas para o item anunciado até a finalização da sessão do Leilão.

## Features
- Cadastro de usuários
- Cadastro de Produtos
- Login/Logout de usuários

## Techs

- Chi: Router
- PostgreSQL: como base de dados
- SQLC: para criação de CRUD facilitada
- Tern: para criação de migrations facilitada
- Pgx: driver para acesso a base postgres
- SCS: gerenciamento de sessões em Go
- Gorilla CSRF Token: segurança na manitulação de tokens
- Gorilla WebSocket: abrir uma sala de leilão

## Dev

### Air

Iniciar o hot-reload Air
```bash
air --build.cmd "go build -o ./bin/api ./cmd/api" --build.bin "./bin/api"
```

### SQLc

Criar models a partir dos SQLs
```bash
sqlc generate -f ./internal/store/pgstore/sqlc.yml
```

### Tern

Executar as migrations
```bash
go run ./cmd/terndotenv
```

Criar tabelas
```bash
cd internal/store/pgstore/migrations
tern new create_<table_name>_table
```


