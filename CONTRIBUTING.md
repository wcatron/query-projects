# Contributing to the Project

Thank you for considering contributing to our project! Here are some guidelines to help you get started.

## Introduction

Query Projects is a simple CLI tool designed to run scripts across multiple repositories. It helps manage repositories and execute TypeScript scripts efficiently, making it easier to automate tasks and analyze projects.

## How to Contribute

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Write your code and tests.
4. Ensure all tests pass.
5. Submit a pull request.

## Running Locally

1. Build the application: `go build`
2. Change cwd to the example directory: `cd example`
3. Run the compiled application: `../query-project info` (`gow build` for watching)

TODO: Document a typical watch based appoach to local development.

## Style Guide

- Follow the existing code style.
- Use `gofmt` to format your code.
- Write clear and concise commit messages.

## Testing

- Run `go test ./tests/... ./commands/...` to execute all tests.
- Add new tests for your code.

## Visualizing

This project uses AI extensively for code generation. One means of protecting against poor quality is to evalute the code structure regularly. Run `go-callvis` to regenerate the  

```
go install github.com/ofabry/go-callvis@latest
go-callvis -nostd -format svg ./...
```

## Static Code Analysis (Experimental)

Check for high complexity code.

```
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
gocyclo -over 10 **/*.go
```

## Communication

- Use GitHub issues for bug reports and feature requests.
- Join our [Slack channel/Discord server] for discussions.