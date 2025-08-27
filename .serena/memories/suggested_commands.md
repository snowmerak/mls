# Suggested Commands for MLS Development

## Basic Go Commands

### Building
```fish
# Build all packages
go build ./...

# Build specific package
go build ./lib/tree
go build ./lib/tree/disk
```

### Testing
```fish
# Run all tests (none exist yet)
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Code Quality
```fish
# Format code
go fmt ./...

# Run static analysis
go vet ./...

# Run golint (if installed)
golint ./...
```

### Dependencies
```fish
# Download dependencies
go mod download

# Clean up dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Development Workflow
```fish
# Check if code compiles
go build ./...

# Format and vet before committing
go fmt ./... && go vet ./...
```

## Git Commands (Linux)
```fish
# Basic git operations
git status
git add .
git commit -m "message"
git push origin main

# View project structure
tree # or ls -la if tree not available
find . -name "*.go" -type f
```

## System Commands (Linux/fish)
```fish
# Navigate project
cd /home/rileyhalo/Projects/snowmerak/mls

# Search in code
grep -r "pattern" lib/
find . -name "*.go" -exec grep -l "pattern" {} \;

# File operations
ls -la
cat filename
head -n 20 filename
```