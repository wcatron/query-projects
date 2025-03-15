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

```