// # Question: "Does the project have a linter?"

// # Constants
// Define the path to the package.json file
const packageJsonPath = './package.json';

try {
    // # Load relevant data
    // Use Deno's readTextFile function to read the contents of the package.json file asynchronously
    const decoder = new TextDecoder("utf-8");
    const packageJsonContent = await Deno.readFile(packageJsonPath);

    // Parse the JSON content of the package.json file
    const packageJson = JSON.parse(decoder.decode(packageJsonContent));

    // Check if the project has ESLint installed as a devDependency
    const hasESLint = packageJson.devDependencies && packageJson.devDependencies.eslint;

    // # Log results
    // If ESLint is found, print 'Yes', otherwise print 'No'
    console.log(hasESLint ? 'Yes' : 'No');
} catch (error) {
    // # Log error to stdout and error information to stderr
    console.log('Error');
    console.error('Error reading package.json file:', error);
}