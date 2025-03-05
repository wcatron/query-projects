// # Question: "what package manager is being used?"

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

    // Check for the presence of a lockfile to determine the package manager
    // In this example, we assume that if a yarn.lock file is present, Yarn is used; otherwise, we assume npm is used.
    const yarnLockExists = await Deno.stat('./yarn.lock').then(() => true).catch(() => false);

    // # Log results
    // Print the package manager based on the lockfile existence
    if (yarnLockExists) {
        console.log('Yarn');
    } else {
        console.log('npm');
    }
} catch (error) {
    // # Log error to stdout and error information to stderr
    console.log('Error');
    console.error('Error reading package.json file:', error);
}