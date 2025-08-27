# Tech Stack and Dependencies

## Programming Language
- **Go 1.25.0** (latest version requirement)

## Module Information
- Module path: `github.com/snowmerak/mls`
- No external dependencies currently (only standard library)

## Project Structure
```
mls/
├── go.mod              # Go module definition
├── LICENSE             # MIT License
├── .gitignore          # Standard Go gitignore
└── lib/
    └── tree/
        ├── tree.go     # Core interfaces (Element, Tree)
        └── disk/
            └── tree.go # Disk implementation (unimplemented)
```

## Development Environment
- Target OS: Linux (user's environment)
- Shell: fish
- No specific IDE configuration files present