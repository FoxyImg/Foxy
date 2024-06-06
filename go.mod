module foxy

go 1.22

require (
	github.com/aws/aws-sdk-go v1.53.10
	github.com/cyphar/filepath-securejoin v0.2.5
	github.com/davidbyttow/govips/v2 v2.14.0
	github.com/gosimple/slug v1.14.0
	github.com/jackc/pgx/v5 v5.5.5
	github.com/jinzhu/copier v0.4.0
	github.com/joho/godotenv v1.5.1
	github.com/redis/go-redis/v9 v9.5.1
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/image v0.10.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/davidbyttow/govips/v2 => github.com/interfacelab/govips-foxy/v2 v2.14.2
