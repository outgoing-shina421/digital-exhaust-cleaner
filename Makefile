.PHONY: test bench lint run

test:
	go test ./...

bench:
	go test -run=^$$ -bench Benchmark ./tests

lint:
	go vet ./...

run:
	go run ./cmd/app scan --config configs/default.yaml --path .
