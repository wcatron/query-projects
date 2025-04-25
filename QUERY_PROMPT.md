The user is asking a question about the provide software development project in the current working directory.
Please generate a TypeScript script that can be run using Deno to answer this question. 
Assume the script will run in a Deno environment with access to the local filesystem. 
The script should focus on answering the question by performing relevant file operations.
Ensure the script is well-commented and follows best practices for file access in Deno.
The script should console log only the answer to the question or "Error".
If the script requires subjective analysis it should pull relevant files and prompt ChatGPT for the answer.
The script should assume an API Key is available in the environment variable OPENAI_API_KEY.
Use local git commit history instead of the GitHub api for any git history information

Example:

```ts

// # Question: "what version of typescript is being used?"

// # Constants
// Define the path to the package.json file
const packageJsonPath = './package.json';

// Take an info argument and return information about the script
if (Deno.args.length > 0 && Deno.args[0] === '--info') {
    console.log(JSON.stringify({
        version: '1.0',
        output: 'text'
    }));
    Deno.exit();
}

try {
    // # Load releavant data
    // Use Deno's readTextFile function to read the contents of the package.json file asynchronously
    const decoder = new TextDecoder("utf-8");
    const packageJsonContent = await Deno.readFile(packageJsonPath);

    // Parse the JSON content of the package.json file
    const packageJson = JSON.parse(decoder.decode(packageJsonContent));

    // Check for TypeScript in dependencies or devDependencies
    const typescriptVersion =
        (packageJson.dependencies && packageJson.dependencies.typescript) ||
        (packageJson.devDependencies && packageJson.devDependencies.typescript);

    // # Log results
    // If TypeScript is found, print the version, otherwise print an appropriate message
    if (typescriptVersion) {
        console.log(typescriptVersion);
    } else {
        console.log('N/A');
    }
} catch (error) {
  // # Log error to stdout and error information to stderr
  console.log('Error');
  console.error('Error reading package.json file:', error);
}
```

The script should take an info argument and return one of the following options:
1. `{"version":"1.0","output": "text"}`
2. `{"version":"1.0","output": "csv", "columns:":["a", "b"]}`
3. `{"version":"1.0","output": "json"}`

Here is the question they would like to answer:

{{QUESTION}}