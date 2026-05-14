# Operations

## Run a Scan

```powershell
go run ./cmd/app scan --path "C:\Users\You\Downloads" --config configs/default.yaml --report reports/scan.html
```

The report is a standalone HTML file that can be opened locally in a browser.

## Run the Interactive UI

```powershell
go run ./cmd/app serve --path "C:\Users\You\Downloads" --config configs/default.yaml --addr 127.0.0.1:8787
```

Open `http://127.0.0.1:8787` in your browser. The Delete button moves a file to quarantine; it does not hard-delete.

## Run Tests

```powershell
go test ./...
```

For a full local verification pass:

```powershell
make test
```

## Run Benchmarks

```powershell
go test -run=^$ -bench Benchmark ./internal/...
```

## Database

The default SQLite database is written to:

```txt
.digital-exhaust-cleaner/cleaner.db
```

The SQLite schema is embedded from `internal/storage/migrations/schema.sql`.
