// Take an info argument and return information about the script
if (Deno.args.length > 0 && Deno.args[0] === '--info') {
    console.log(JSON.stringify({
        version: '1.0',
        output: 'text'
    }));
    Deno.exit();
}

// Define the path to the package.json file
const packageJsonPath = './package.json';

try {
    // Use Deno's readTextFile function to read the contents of the package.json file asynchronously
    const decoder = new TextDecoder("utf-8");
    const packageJsonContent = await Deno.readFile(packageJsonPath);

    // Parse the JSON content of the package.json file
    const packageJson = JSON.parse(decoder.decode(packageJsonContent));

    // Check for TypeScript in dependencies or devDependencies
    const typescriptVersion =
        (packageJson.dependencies && packageJson.dependencies.typescript) ||
        (packageJson.devDependencies && packageJson.devDependencies.typescript);

    // If TypeScript is found, print the version, otherwise print an appropriate message
    if (typescriptVersion) {
        console.log(typescriptVersion);
    } else {
        console.log('N/A');
    }
} catch (error) {
  console.log('Error');
  console.error('Error reading package.json file:', error);
}
