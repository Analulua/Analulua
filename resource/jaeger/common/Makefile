TESTS=go test $$(go list ./... | grep -v /vendor/) -cover

build:
	go build

unittest:
	go test -short $$(go list ./... | grep -v /vendor/)

test:
	go test ./... -coverpkg=./... -coverprofile=coverage.out -json > test.out
	gocover-cobertura < coverage.out > coverage.xml

lint:
	golangci-lint run --out-format checkstyle > lint.xml

.PHONY: unittest test
