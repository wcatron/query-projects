# Query Projects

[![Dependabot Updates](https://github.com/wcatron/query-projects/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/wcatron/query-projects/actions/workflows/dependabot/dependabot-updates)
[![CodeQL](https://github.com/wcatron/query-projects/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/wcatron/query-projects/actions/workflows/github-code-scanning/codeql)
[![Go Cyclo Complexity Check](https://github.com/wcatron/query-projects/actions/workflows/gocyclo.yml/badge.svg)](https://github.com/wcatron/query-projects/actions/workflows/gocyclo.yml)

A cli for running scripts across your organizations repos.

- Manages a local copy of all your repos for fast up-to-date access to the current state of your system. `query-projects add`, `query-projects pull`, and `query-projects sync`.
- Provides a Typescript and Lua (Currently referred to as "plans") scripting system to run deterministic analysis across your codebases with `query-projects run`.
- Explores ways to use AI to generate analysis across your codebases with `query-projects ask`.

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

It also updates (or creates) the projects.json file with the new project's information.

Confirm your project was added:

```
query-projects info
```

You should see the "Number of Projects" value incremented by 1.

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

### Syncing Project Metadata

The `sync` command allows you to synchronize project metadata from a specified code repository. Currently, it supports syncing from GitHub. The command requires a single argument specifying the repository type (e.g., "github"). It uses the `GITHUB_TOKEN` environment variable for authentication.

Example usage:
```
query-projects sync github
```
This command will fetch metadata for all projects listed in the `projects.json` file from GitHub and update the project metadata with topics and archive status.


### Pulling Updates

The `pull` command allows you to update all tracked repositories with the latest changes from their remote sources. This command is useful for keeping your local copies of repositories up to date.

Example usage:
```
# Pull updates for all projects
query-projects pull

# Pull updates for specific topics
query-projects pull --topics typescript,react

# Pull updates with required/excluded topics
query-projects pull --topics +typescript,-deprecated
```

The `pull` command supports the same topic filtering as the `run` command:
- Use `+` prefix for required topics
- Use `-` prefix for excluded topics
- No prefix for an inclusive search

The command will:
1. Update each repository using `git pull`
2. Show the status of each update
3. Report any errors that occur during the update process

A note on authentication with GitHub:
- Ideal: Use your system's git configuration
- Optional: Provide a `GITHUB_TOKEN` env with a personal access token

### Create Scripts

There are several ways to create scripts for use with `query-projects`:

#### 1. Using GPT (AI-Generated Scripts)

The `ask` command allows you to generate scripts using OpenAI's GPT models. This is a useful way to get results for a given question quickly without having to write the script yourself. 

```bash
# Generate and run a script to analyze the 
query-projects ask "How many services are still using Node 18?"

# Generate a script with specific requirements
query-projects ask "Find all React components that use the useState hook"
```

Requirements:
- `OPENAI_API_KEY`: Your OpenAI API key
- `OPENAI_API_BASE`: (Optional) Base URL for the OpenAI API. Defaults to `https://api.openai.com/v1`

#### 2. Manual Script Creation

You can create scripts manually in either TypeScript or Lua. Scripts should be placed in a `scripts` directory.

TypeScript Example (`scripts/find-ts-files.ts`):
```typescript
import { script } from "../scripts/lib/utils.ts";

await script({ type: "text" }, () => {
  // Your script logic here
  return "Script output"
});
```

Lua Example (`scripts/find-ts-files.lua`):
```lua
-- Your Lua script logic here
return "Script output"
```

#### 3. Script Output Types

Scripts can output data in different formats:

- **Text**: Simple string output
  ```typescript
  await script({ type: "text" }, (emit) => {
    emit("Hello");
    emit("World");
  });
  ```

- **CSV**: Tabular data with columns
  ```typescript
  await script({ 
    type: "csv",
    columns: ["name", "version"]
  }, (emit) => {
    emit(["typescript", "4.9.0"]);
  });
  ```

- **JSON**: Structured data
  ```typescript
  await script({ type: "json" }, (emit) => {
    emit({
      name: "project",
      version: "1.0.0"
    });
  });
  ```

#### 4. Script Utilities

The `utils.ts` library provides helpful functions for common tasks:

- `packageManager`: Access package.json information
  ```typescript
  const version = packageManager.dependency("typescript");
  ```

- `value`: Extract values from configuration files
  ```typescript
  const version = value("package.json", "dependencies.typescript");
  ```

#### 5. Best Practices

1. **Error Handling**: Always include proper error handling in your scripts only stdout is captured in the final analysis
   ```typescript
   try {
     // Script logic
   } catch (error) {
     console.error("Script failed:", error);
   }
   ```

2. **Output Formatting**: Choose the appropriate output type for your data
   - Use CSV for tabular data
   - Use JSON for structured data
   - Use text for simple outputs

3. **Performance**: Consider the impact of your script on large codebases
   - Use efficient file operations
   - Avoid unnecessary file reads
   - Consider using caching for repeated operations

4. **Documentation**: Include comments explaining what your script does and how to use it

### Run Scripts

The `run` command executes scripts across your tracked repositories. It supports both TypeScript (and Lua should be added in the future), with various options for filtering and output formatting.

#### Basic Usage

```bash
# Run a TypeScript script
query-projects run --script scripts/find-ts-files.ts
```

#### Output Options

The `--output` flag allows you to specify the format(s) for script results:

```bash
# Output in multiple formats
query-projects run --script scripts/find-ts-files.ts --output md,csv,json

# Default behavior: automatically choose format based on content
query-projects run --script scripts/find-ts-files.ts
```

Output format selection:
- **JSON**: Used when outputs are valid JSON objects
- **CSV**: Used for tabular data or when outputs are single-line
- **Markdown**: Used for text-based outputs or when outputs contain multiple lines

#### Project Filtering

Filter which projects to run the script against using topics:

```bash
# Run on projects with specific topics
query-projects run scripts/find-ts-files.ts --topics typescript,react

# Run on projects with required/excluded topics
query-projects run scripts/find-ts-files.ts --topics +typescript,-deprecated
```

Topic filtering syntax:
- `+topic`: Project must have this topic
- `-topic`: Project must not have this topic
- `topic`: Project may have this topic (inclusive search)

#### Response Counting

Use the `--count` flag to quickly analyze the distribution of script responses:

```bash
# Count unique responses
query-projects run scripts/find-ts-files.ts --count
```

This will output a table showing:
- Each unique response
- Number of occurrences
- Percentage of total responses

Note: The count feature works best with simple string responses (no line breaks).

#### Script Environment

Scripts have access to:
1. The current project's root directory as the *current working directory*
2. The `jsr:@query-projects/scripts` library for common utilities
3. Standard Deno APIs
4. Any arguments passed to `query-projects run` (i.e. `query-projects run typescript`)

Example script checking version of dependency:
```typescript

import { script, packageManager } from "jsr:@query-projects/scripts";

script({ type: 'text' }, () => {
  const packageStr = Deno.args[0];

  const projectPath = Deno.cwd();
  
  return packageManager.dependency(packageStr) || packageManager.devDependency(packageStr);
});

```

#### Error Handling

- Script errors are captured and reported per project
- The command continues running on other projects even if some fail
- Error messages are included in the final output

#### Performance Tips

1. **Caching**: Consider caching expensive operations
2. **Selective Processing**: Use topic filtering to run only on relevant projects
3. **Output Size**: Keep outputs concise to improve performance and ease analysis

#### Example Use Cases

1. **Dependency Analysis**:
```bash
query-projects run --script scripts/check-deps.ts --topics +typescript
```

2. **Code Quality Checks**:
```bash
query-projects run--script scripts/lint-check.ts --output md,csv
```

3. **Configuration Audit**:
```bash
query-projects run--script scripts/check-config.ts --topics +react
```

4. **Custom Metrics**:
```bash
query-projects run--script scripts/custom-metrics.ts --count
```

## Contributing

See [contributing](./CONTRIBUTING.md).

## Similar Tools

- https://github.com/nosarthur/gita (python)
