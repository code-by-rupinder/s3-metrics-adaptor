# Contributing to s3-metrics-adaptor

First off, thank you for considering contributing to S3 Event Exporter! It's people like you that make this tool better for everyone.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the issue list as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

- Use a clear and descriptive title
- Describe the exact steps which reproduce the problem
- Provide specific examples to demonstrate the steps
- Describe the behavior you observed after following the steps
- Explain which behavior you expected to see instead and why
- Include logs if relevant

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please provide:

- Use a clear and descriptive title
- Provide a step-by-step description of the suggested enhancement
- Provide specific examples to demonstrate the steps
- Describe the current behavior and explain which behavior you expected to see instead
- Explain why this enhancement would be useful

### Pull Requests

- Fill in the required template
- Do not include issue numbers in the PR title
- Include screenshots and animated GIFs in your pull request whenever possible
- Follow the Go coding style
- Include adequate tests
- Document new code
- End all files with a newline

## Development Setup

1. Fork the repo
2. Clone your fork
3. Create a branch
4. Make your changes
5. Run tests
6. Push to your fork
7. Submit a Pull Request

### Setting up your development environment

```bash
# Clone your fork
git clone https://github.com/code-by-rupinder/s3-metrics-adaptor.git

# Add upstream remote
git remote add upstream https://github.com/code-by-rupinder/s3-metrics-adaptor.git

# Create a branch
git checkout -b feature/your-feature-name

# Install dependencies
go mod download
```

### Running Tests

```bash
go test ./...
```

## Questions?

Feel free to open an issue with your question.
