# Code Style and Conventions

## Go Conventions
The project follows standard Go conventions:

### Naming
- Interfaces use PascalCase: `Element`, `Tree`
- Methods use PascalCase: `Name()`, `Value()`, `LeftChild()`
- Packages use lowercase: `tree`, `disk`
- No specific naming prefix/suffix patterns observed yet

### Interface Design
- Interfaces are kept minimal and focused
- Clear separation between `Element` (node operations) and `Tree` (tree operations)
- Use of interface{} is avoided in favor of concrete types where possible

### Package Structure
- Interfaces defined in parent package (`tree`)
- Implementations in subpackages (`tree/disk`)
- Interface compliance verified with compile-time checks: `var _ tree.Element = &Element{}`

### Error Handling
- Methods return `error` type for operations that can fail
- Find operations return `(Element, bool)` pattern for optional results

### Documentation
- Currently minimal documentation
- No package-level documentation present
- No method documentation present (should be added)

## Missing Standards
The following should be established:
- Comprehensive package and method documentation
- Code formatting standards (gofmt is implied)
- Testing patterns and conventions
- Error message formatting standards