// Define the path to the README.md file
const readmePath = './README.md';

// Take an info argument and return information about the script
if (Deno.args.length > 0 && Deno.args[0] === '--info') {
    console.log(JSON.stringify({
        version: '1.0',
        output: 'text'
    }));
    Deno.exit();
}

try {
    // Check if the README.md file exists in the current working directory
    const readmeExists = await Deno.stat(readmePath).then(() => true).catch(() => false);

    // If README.md exists, print "Yes", otherwise print "No"
    if (readmeExists) {
        console.log('Yes');
    } else {
        console.log('No');
    }
} catch (error) {
    console.log('Error');
    console.error('Error checking README.md file:', error);
}