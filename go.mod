module github.com/bitcoinschema/go-1sat-ord

go 1.24.1

require (
	github.com/bitcoin-sv/go-templates v0.0.0
	github.com/bsv-blockchain/go-sdk v1.1.22
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/bitcoin-sv/go-templates => ../go-templates

replace github.com/bsv-blockchain/go-sdk => github.com/b-open-io/go-sdk v1.1.22-0.20250329172752-ca68d5bf1bee
