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

### Topic Filtering

You can filter projects by topics when using the `run` and `pull` commands. The filtering logic supports:

- **Required Topics**: Prefix with `+` to include only projects with this topic.
- **Excluded Topics**: Prefix with `-` to exclude projects with this topic.
- **Optional Topics**: No prefix, includes projects with at least one of these topics.

Example usage:
```
query-projects run --topics a,b,+c,-d
```
This command runs scripts on projects that have topic `a` or `b`, must have `c`, and must not have `d`.

### Output Formats

The `run` command now supports specifying output formats using the `--output` flag. You can choose from `md`, `csv`, or `json`. By default, the tool will determine the best output format based on the script results:
- If the majority of outputs are valid JSON, it will export as JSON.
- If outputs are single-line, it will export as both Markdown and CSV.
- Users can override the default by specifying the desired format(s).

Example usage:
```
query-projects run --output md,csv
```
This command will execute the specified script and output the results in Markdown and CSV formats.

The `run` command includes a `--count` flag that allows you to count the number of unique responses from the scripts executed across all projects. This can be useful for analyzing the diversity of outputs from your scripts.

Example usage:
```
query-projects run --count
```
This command will execute the specified script and print a table showing each unique response and the count of occurrences. Note that the `count` feature requires simple strings with no line breaks, and all whitespace will be removed from the responses.

### Environment Variables

- `OPENAI_API_KEY`: Your OpenAI API key. This is required for querying GPT.
- `OPENAI_API_BASE`: The base URL for the OpenAI API. Defaults to `https://api.openai.com/v1` if not set.

### Next Steps:

- **Ask GPT**: Use the `ask` command to generate new scripts via OpenAI (assuming you have `OPENAI_API_KEY` set in your environment).
- **Run Scripts**: Use the `run` command to execute TypeScript scripts across all tracked projects.
- **Pull Updates**: Use the `pull` command to update all tracked repositories with the latest changes.


## Contributing

## Similar Tools

- https://github.com/nosarthur/gita (python)
