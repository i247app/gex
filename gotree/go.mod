// module github.com/tnmonex/gotree // Replace with your actual module path

module gotree

go 1.20

require (
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/joho/godotenv v1.5.1 // For reading .env files
	github.com/go-sql-driver/mysql v1.7.1
)

//	Indirect dependencies (these will be automatically added and managed by go mod)
require (
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
)
