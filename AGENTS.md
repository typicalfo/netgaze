# Development Guidelines for Agents

## Build/Test Commands
```bash
# Build the project
make build

# Run tests
go test ./...

# Run single test with verbose output
go test -v ./internal/collector -run TestSpecificFunction

# Run tests with coverage
go test -cover ./...

# Lint and format
go fmt ./...
go vet ./...
```

## Code Style Guidelines

### Imports & Formatting
- Use standard Go formatting (`go fmt`)
- Group imports: stdlib, third-party, internal
- Use `gofmt` and `go vet` before committing

### Naming Conventions
- Package names: lowercase, single word when possible
- Functions: camelCase, export if public
- Variables: camelCase, descriptive names
- Constants: UPPER_SNAKE_CASE for exported constants

### Error Handling
- Always handle errors, never ignore them
- Use fmt.Errorf for wrapping errors with context
- Return errors from all functions that can fail
- Use structured error messages with collector names

### Types & Structs
- Use the exact Report struct from dev-docs/01-report-struct-schema.md
- Add JSON tags for all exported fields
- Use omitempty for optional fields
- Include validation methods where appropriate

### Critical Requirements
- **NO SPECIAL CHARACTERS**: No emoji, unicode symbols, or decorative characters in any code, templates, or output
- Use "Success"/"Failed"/"Warning" instead of symbols
- Ensure all output is copy-paste compatible
- Follow graceful degradation pattern from dev-docs/14-error-handling-graceful-degradation.md

### Testing
- Write unit tests for all collectors
- Test error conditions and timeouts
- Use table-driven tests for multiple scenarios
- Mock external dependencies in tests

### Performance
- Target sub-12 second total runtime
- Use context.WithTimeout for all network operations
- Run collectors in parallel using errgroup
- Optimize memory allocations

### Status Tracking
- Update dev-docs/CURRENT_STATUS.md before starting any task
- Update dev-docs/CURRENT_STATUS.md after completing any task
- Include task name, phase, and completion status
- Keep entries brief and timestamped

Follow the implementation plan in dev-docs/IMPLEMENTATION_PLAN.md and reference individual collector specifications in dev-docs/.