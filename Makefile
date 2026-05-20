.PHONY: test bench lint run build ui-install ui-build ui-dev

# Frontend targets
ui-install:
	cd web && npm install

ui-build: ui-install
	cd web && npm run build

ui-dev:
	cd web && npm run dev

# Go targets
build: ui-build
	go build -o bin/app ./cmd/app

test:
	go test ./...

bench:
	go test -run=^$$ -bench Benchmark ./tests

lint:
	go vet ./...

run: ui-build
	go run ./cmd/app scan --config configs/default.yaml --path .
