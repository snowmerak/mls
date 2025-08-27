# Codebase Structure and Architecture

## Directory Structure
```
mls/
├── go.mod                    # Go module definition
├── LICENSE                   # MIT License
├── .gitignore               # Standard Go gitignore
├── .serena/                 # Serena AI assistant configuration
│   ├── project.yml          # Project configuration
│   └── memories/            # Memory files
└── lib/                     # Library code
    └── tree/                # Tree data structure package
        ├── tree.go          # Core interfaces
        └── disk/            # Disk-based implementation
            └── tree.go      # Disk implementation (stub)
```

## Core Architecture

### Interfaces (`lib/tree/tree.go`)
- **Element Interface**: Represents tree nodes with:
  - Data operations: `Name()`, `Value()`
  - Tree navigation: `LeftChild()`, `RightChild()`
  - Metadata: `LeftCount()`, `RightCount()`
  - Mutation: `SetLeftChild()`, `SetRightChild()`, `SetLeftCount()`, `SetRightCount()`

- **Tree Interface**: Represents tree operations:
  - Access: `Head()`, `Find(name string)`
  - Mutation: `Insert(name string, value []byte)`, `Delete(name string)`

### Implementation Strategy
- **Abstraction**: Clean separation between interface and implementation
- **Modularity**: Each implementation in its own package
- **Extensibility**: Easy to add new storage backends (memory, network, etc.)

## Design Patterns
- **Interface Segregation**: Clean, focused interfaces
- **Dependency Inversion**: Depend on abstractions, not concretions
- **Factory Pattern**: (Planned) for creating different tree implementations

## Current Limitations
- No concrete implementations are functional yet
- Missing comprehensive test suite
- No performance benchmarks
- No example usage or documentation