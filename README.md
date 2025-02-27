# Query Projects

A simple cli for running scripts across many repositories.

## Getting Started (User Guide)

### Installation
Prerequisites:
Go 1.18+ installed on your system.
Install the CLI directly from GitHub:

```
go install github.com/wcatron/query-projects@latest
```

This command fetches, builds, and installs the query-projects executable into your Go bin directory. Make sure your PATH is configured to include $GOPATH/bin (or the appropriate Go bin folder on your system).
Verify Installation:

```
query-projects --help
```

You should see the CLI usage instructions.

### Adding Projects
Once the CLI is installed, you can start tracking repositories by adding them to projects.json:

Create or Navigate to a Working Directory (optional). Wherever you run commands from, projects.json will be generated or updated there.
Add a Project:

```
query-projects add <repo-url>
```

This command clones the repository (if not already present) into a projects/ folder.

It also updates (or creates) the projects.json file with the new project’s information.

Confirm your project was added:

```
query-projects info
```

You should see a JSON entry for your new repository.

### Next Steps:

Run Scripts: Use the `run` command to execute TypeScript scripts across all tracked projects.

Query GPT: Use the `query` command to generate new scripts via OpenAI (assuming you have `OPENAI_API_KEY` set in your environment).

Pull Updates: Use the pull command to update all tracked repositories with the latest changes.


## Contributing

## Similar Tools

- https://github.com/nosarthur/gita (python)
