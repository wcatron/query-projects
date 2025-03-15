# Scripts

Simple self contained scripts that can be executed against each project in your meta project.

## Info Spec v1.0



## Deno

```ts
// query-project-api-version: 1.0

// Take an info argument and return information about the script
if (Deno.args.length > 0 && Deno.args[0] === '--info') {
    console.log(JSON.stringify({
        version: '1.0',
        output: 'text'
    }));
    Deno.exit();
}

const currentWorkingDirectory = Deno.cwd();

// Should run process for cwd

console.log(['datum_1', 'datum_2'].join(","))

```## Info Spec v1.0

| Key     | Description                                                                 |
|---------|-----------------------------------------------------------------------------|
| version | The version of the script. Should be '1.0' for compatibility.               |
| output  | The type of output the script generates. Can be 'text', 'csv', or 'json'.   |
| columns | Required if `output` is 'csv'. An array specifying the column headers.      |
