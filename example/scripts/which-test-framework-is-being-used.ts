// # Question: "which test framework is being used?"

// # Constants
// Define the path to the package.json file
const packageJsonPath = './package.json';

// Take an info argument and return information about the script
if (Deno.args.length > 0 && Deno.args[0] === '--info') {
    console.log(JSON.stringify({
        version: '1.0',
        output: 'csv'
    }));
    Deno.exit();
}

try {
    // # Load relevant data
    // Use Deno's readTextFile function to read the contents of the package.json file asynchronously
    const decoder = new TextDecoder("utf-8");
    const packageJsonContent = await Deno.readFile(packageJsonPath);

    // Parse the JSON content of the package.json file
    const packageJson = JSON.parse(decoder.decode(packageJsonContent));

    // Check for test framework and its version in devDependencies
    let testFramework = null;
    let version = null;

    if (packageJson.devDependencies) {
        if (packageJson.devDependencies.jest) {
            testFramework = 'Jest';
            version = packageJson.devDependencies.jest;
        } else if (packageJson.devDependencies.mocha) {
            testFramework = 'Mocha';
            version = packageJson.devDependencies.mocha;
        } else if (packageJson.devDependencies['@testing-library']) {
            testFramework = 'Testing Library';
            version = packageJson.devDependencies['@testing-library'];
        }
    }

    // # Log results
    // If test framework is found, print the name and version, otherwise print "N/A"
    if (testFramework && version) {
        console.log(`${testFramework},${version}`);
    } else {
        console.log('N/A,N/A');
    }
} catch (error) {
    // # Log error to stdout and error information to stderr
    console.log('Error');
    console.error('Error reading package.json file:', error);
}
