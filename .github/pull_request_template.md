## Description

Brief description of the changes introduced by this PR.

## Type of Change

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📚 Documentation update
- [ ] 🧪 Test improvements
- [ ] 🔧 Refactoring (code changes that neither fix bugs nor add features)
- [ ] ⚡ Performance improvements
- [ ] 🔒 Security improvements

## Changes Made

### Core Changes
- List the main changes made to the codebase
- Include new files, modified files, deleted files
- Mention any architectural changes

### API Changes (if applicable)
- List any changes to public APIs
- Include breaking changes and migration notes
- Document new public methods/types

## Testing

- [ ] Tests added for new functionality
- [ ] All existing tests pass (`make test`)
- [ ] Manual testing performed
- [ ] Examples updated and tested

### Test Coverage
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated  
- [ ] Example code tested
- [ ] Documentation examples verified

## Checklist

### Code Quality
- [ ] Code follows Go style guidelines
- [ ] Code has been self-reviewed
- [ ] Code is properly documented (godoc comments)
- [ ] Linting passes (`make lint`)
- [ ] No security vulnerabilities introduced

### Documentation
- [ ] README updated (if needed)
- [ ] Examples updated (if needed)
- [ ] CHANGELOG.md updated (if needed)
- [ ] Code comments are clear and helpful

### Dependencies
- [ ] No new dependencies added to core library
- [ ] External dependencies properly justified (examples/integrations only)
- [ ] go.mod and go.sum updated appropriately

## Breaking Changes

If this includes breaking changes, please describe:

1. **What breaks**: Describe what existing functionality changes
2. **Migration path**: How should users update their code?
3. **Deprecation**: Are there any deprecation warnings?

```go
// Example of old vs new usage
// Old way:
// agent.WithOldMethod(param)

// New way:
// agent.WithNewMethod(param)
```

## Performance Impact

- [ ] No performance impact
- [ ] Performance improved
- [ ] Performance impact acceptable for the feature
- [ ] Performance impact documented

**Benchmarks** (if applicable):
```
// Include benchmark results
```

## Examples

If this change affects how users interact with the library, provide examples:

### Basic Usage
```go
// Show how to use the new feature
```

### Advanced Usage  
```go
// Show more complex scenarios
```

## Screenshots/Logs

If applicable, add screenshots or log output to help explain the changes.

## Additional Context

Add any other context about the pull request here:
- Related issues or discussions
- Alternative approaches considered
- Future work planned
- Any concerns or questions

## Review Checklist for Maintainers

- [ ] PR title is clear and descriptive
- [ ] Code changes align with project goals
- [ ] Architecture and design are sound
- [ ] Security considerations addressed
- [ ] Performance implications considered
- [ ] Documentation is adequate
- [ ] Tests provide good coverage
- [ ] Breaking changes are justified and documented