# Scripts

Simple self contained scripts that can be executed against each project in your meta project.

## Deno

```ts
// query-project-api-version: 1.0

// Should return column headers if output type is csv and "--columns" argument is passed

if (Deno.args.length > 0 && Deno.args[0] === '--columns') {
    console.log([]'column_1', 'column_2'].join(","));
    Deno.exit();
}

const currentWorkingDirectory = Deno.cwd();

// Should run process for cwd

console.log(['datum_1', 'datum_2'].join(","))

```