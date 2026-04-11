module github.com/libdns/timeweb/timeweb-tests

go 1.21

require (
	github.com/joho/godotenv v1.5.1
	github.com/libdns/libdns v1.2.0-alpha.1
	github.com/libdns/timeweb v0.0.0
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/libdns/timeweb => ../
