# Task Completion Checklist

## Before Committing Code

### Code Quality Checks
1. **Format Code**: Run `go fmt ./...` to ensure consistent formatting
2. **Static Analysis**: Run `go vet ./...` to catch potential issues
3. **Build Check**: Run `go build ./...` to ensure compilation succeeds
4. **Test Execution**: Run `go test ./...` (when tests exist)

### Documentation
1. **Package Documentation**: Ensure packages have proper documentation
2. **Method Documentation**: Add documentation for public methods and types
3. **README Updates**: Update README.md when adding new features (when it exists)

### Git Workflow
1. **Stage Changes**: `git add .` or specific files
2. **Commit**: `git commit -m "descriptive message"`
3. **Push**: `git push origin main`

## Development Guidelines

### When Adding New Features
1. **Interface First**: Define interfaces before implementations
2. **Error Handling**: Always handle errors appropriately
3. **Test Coverage**: Add tests for new functionality
4. **Documentation**: Document new public APIs

### When Implementing Disk Operations
1. **Error Handling**: File operations should handle errors gracefully
2. **Resource Management**: Ensure proper file handle cleanup
3. **Thread Safety**: Consider concurrent access patterns
4. **Performance**: Consider disk I/O optimization

### Code Review Checklist
- [ ] Code compiles without errors
- [ ] Code is properly formatted (go fmt)
- [ ] No static analysis warnings (go vet)
- [ ] Public APIs are documented
- [ ] Error handling is appropriate
- [ ] Tests pass (when they exist)
- [ ] Interface compliance is verified