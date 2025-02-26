// # Question: "which test framework is being used?"

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

    // Check for test framework in scripts or devDependencies
    const testFramework = packageJson.devDependencies && packageJson.devDependencies.jest
            ? 'Jest'
            : packageJson.devDependencies && packageJson.devDependencies.mocha
            ? 'Mocha'
            : packageJson.devDependencies && packageJson.devDependencies['@testing-library']
            ? 'Testing Library' : null;

    // # Log results
    // If test framework is found, print the name, otherwise print "N/A"
    if (testFramework) {
        console.log(testFramework);
    } else {
        console.log('N/A');
    }
} catch (error) {
    // # Log error to stdout and error information to stderr
    console.log('Error');
    console.error('Error reading package.json file:', error);
}