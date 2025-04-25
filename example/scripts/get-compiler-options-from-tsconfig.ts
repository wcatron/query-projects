// # Requirements:
// - Fix error: Module not found "file:///Users/westoncatron/github/query-projects/example/Users/westoncatron/github/query-projects/example/scripts/get-compiler-options-from-tsconfig.ts". 
// - Expect the cwd to be set

// # Question: "get compiler options from tsconfig"

// # Constants
// Define the path to the tsconfig.json file
const tsconfigPath = './tsconfig.json';

// Take an info argument and return information about the script
if (Deno.args.length > 0 && Deno.args[0] === '--info') {
    console.log(JSON.stringify({
        version: '1.0',
        output: 'json'
    }));
    Deno.exit();
}

try {
    // # Load releavant data
    // Use Deno's readTextFile function to read the contents of the tsconfig.json file asynchronously
    const decoder = new TextDecoder("utf-8");
    const tsconfigContent = await Deno.readFile(tsconfigPath);

    // Parse the JSON content of the tsconfig.json file
    const tsconfig = JSON.parse(decoder.decode(tsconfigContent));

    // Check if "compilerOptions" exist in tsconfig
    const compilerOptions = tsconfig.compilerOptions;

    // # Log results
    // If compilerOptions are found, print them, otherwise print an appropriate message
    if (compilerOptions) {
        console.log(JSON.stringify(compilerOptions, null, 2));
    } else {
        console.log('Compiler options not found in tsconfig.json');
    }
} catch (error) {
  // # Log error to stdout and error information to stderr
  console.log('Error');
  console.error('Error reading tsconfig.json file:', error);
}