# Contributing to Go Agent Kit

Thank you for your interest in contributing to Go Agent Kit! We welcome contributions from the community and are excited to see what you'll build.

## Quick Start

1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/go-agent-kit.git`
3. Create a feature branch: `git checkout -b feature/amazing-feature`
4. Make your changes
5. Run tests: `make test`
6. Run linting: `make lint`
7. Commit your changes: `git commit -m 'Add amazing feature'`
8. Push to the branch: `git push origin feature/amazing-feature`
9. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Make (for running Makefile commands)

### Local Development

```bash
# Clone the repository
git clone https://github.com/ratlabs-io/go-agent-kit.git
cd go-agent-kit

# Run tests
make test

# Run linting
make lint

# Generate coverage report
make coverage

# Build the library
make build
```

## Testing

We maintain high test coverage. Please ensure your contributions include tests:

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run tests for a specific package
go test ./pkg/workflow/...
```

### Testing Guidelines

- Write unit tests for all new functionality
- Include edge cases and error conditions
- Use table-driven tests where appropriate
- Mock external dependencies
- Maintain or improve test coverage

## Code Style

### Go Style Guidelines

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` and `goimports` for formatting
- Run `golangci-lint` before submitting
- Write clear, self-documenting code
- Include godoc comments for public functions and types

### Naming Conventions

- Use clear, descriptive names
- Follow Go naming conventions (CamelCase for exported, camelCase for unexported)
- Avoid abbreviations unless they're well-known (HTTP, JSON, etc.)

### Code Organization

- Keep the core library dependency-free
- Put examples in `examples/` directory
- External integrations go in `examples/integrations/`
- Tools implementations go in `examples/tools/`

## Contribution Areas

We welcome contributions in these areas:

### LLM Integrations

Add support for new LLM providers in `examples/integrations/`:

- Anthropic Claude
- Google Gemini
- Cohere
- Local models (Ollama, etc.)
- Enterprise providers

### Tools

Create useful reference tools in `examples/tools/`:

- HTTP/API tools
- File system tools
- Database tools
- System utilities
- Integration tools

### Workflow Patterns

Add new execution patterns in `pkg/workflow/`:

- Rate limiting workflows
- Retry mechanisms
- Circuit breakers
- Pipeline patterns

### Documentation

- Improve code documentation
- Add more examples
- Write guides and tutorials
- Improve README clarity

### Testing & Quality

- Expand test coverage
- Add benchmarks
- Performance optimizations
- Bug fixes

## Pull Request Guidelines

### Before Submitting

- [ ] Code follows Go style guidelines
- [ ] Tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Documentation is updated
- [ ] Examples are provided for new features
- [ ] CHANGELOG.md is updated (if applicable)

### PR Template

Please use this template for your pull requests:

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests added/updated
- [ ] Manual testing performed
- [ ] Examples work as expected

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Tests pass
- [ ] Documentation updated
```

## Bug Reports

When reporting bugs, please include:

1. **Go version**: `go version`
2. **Operating System**: OS and version
3. **Reproduction steps**: Minimal code example
4. **Expected behavior**: What should happen
5. **Actual behavior**: What actually happens
6. **Error messages**: Full error output

## Feature Requests

For new features, please:

1. Check existing issues first
2. Describe the use case clearly
3. Explain why it would be valuable
4. Consider implementation approaches
5. Be open to discussion and feedback

## Development Workflow

### Branching Strategy

- `main`: Stable releases
- `feature/description`: New features
- `bugfix/description`: Bug fixes
- `docs/description`: Documentation updates

### Commit Messages

Use clear, descriptive commit messages:

```
feat: add support for Claude API integration
fix: resolve race condition in parallel workflow
docs: update tool development guide
test: add coverage for error handling
```

### Release Process

1. Version bumps follow semantic versioning
2. Update CHANGELOG.md
3. Tag releases with `git tag v1.x.x`
4. GitHub Actions handles publishing

## Architecture Guidelines

### Core Principles

1. **Zero Dependencies**: Keep core library dependency-free
2. **Composability**: All components should be composable
3. **Generic Interfaces**: Support any LLM provider
4. **Clean Architecture**: Separate concerns clearly
5. **Production Ready**: Handle errors gracefully

### Package Structure

```
pkg/                    # Core library (zero dependencies)
├── workflow/           # Workflow orchestration
├── agent/             # Agent implementations  
├── tools/             # Tool system
└── llm/               # LLM abstraction

examples/              # Examples and integrations
├── workflows/         # Complete examples
├── tools/            # Reference tools
└── integrations/     # LLM provider integrations
```

## Community

- **Issues**: Bug reports and feature requests
- **Discussions**: General questions and ideas
- **Discord**: Real-time community chat (coming soon)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

All contributors will be recognized in our README and release notes. Thank you for making Go Agent Kit better!

---

**Questions?** Feel free to open an issue or start a discussion. We're here to help!