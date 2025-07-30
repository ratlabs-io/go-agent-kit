# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.3.x   | :white_check_mark: |
| < 0.3.0 | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in Go Agent Kit, please report it by creating a GitHub issue or emailing the maintainers.

We'll do our best to respond promptly, though please keep in mind this is a small open source project.

## Security Considerations

When using Go Agent Kit, keep these security practices in mind:

### API Key Management
- Never commit API keys to your repository
- Use environment variables for sensitive credentials
- Rotate API keys regularly

```go
// ✅ Good
apiKey := os.Getenv("OPENAI_API_KEY")

// ❌ Bad  
apiKey := "sk-proj-1234567890abcdef"
```

### Tool Security
- Validate tool parameters carefully
- Be cautious with tools that access files or make network requests
- Consider the permissions tools need before using them

### Input Validation
- Sanitize user inputs before sending to LLM providers
- Implement reasonable length limits
- Handle errors gracefully

That's it! Go Agent Kit is a lightweight library - most security considerations are up to how you use it in your applications.