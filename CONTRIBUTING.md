# Contributing to Digital Exhaust Cleaner

Thank you for your interest in contributing to Digital Exhaust Cleaner. This project relies on community involvement to maintain its high standards for privacy and performance.

## Core Principles

1.  **Privacy First**: No features may be added that require network access during the analysis pipeline. All processing must remain local.
2.  **Safety First**: Cleanup actions must be reversible (quarantine/restore) by default to prevent accidental data loss.
3.  **Performance**: Code should be optimized for high-concurrency local filesystem operations.

## Getting Started

### Project Structure
-   `cmd/app`: Entry point for the Command Line Interface and Web Server.
-   `internal/`: Core business logic (scanning, deduplication, intelligence engines).
-   `configs/`: Configuration templates.

### Development Workflow
1.  Fork the repository and create a feature branch.
2.  Install dependencies using `go mod download`.
3.  Implement changes following the established architecture.
4.  Execute tests: `go test ./...`.
5.  Execute linting: `go vet ./...`.
6.  Ensure all public-facing functions are properly documented.

## Commit Message Standards

We adhere to the [Conventional Commits](https://www.conventionalcommits.org/) specification:
-   `feat(scanner): support for hidden file detection`
-   `fix(storage): resolve race condition in SQLite persistence`
-   `docs(readme): update SEO terminology`
-   `test(intelligence): unit tests for image hashing`

## Quality Standards

-   **Test Coverage**: New features must include comprehensive unit tests.
-   **Documentation**: Public APIs must have clear, descriptive comments.
-   **Idiomatic Go**: Code should follow standard Go conventions (verified via `gofmt` and `go vet`).

---

*Thank you for helping build a more private digital ecosystem.*
