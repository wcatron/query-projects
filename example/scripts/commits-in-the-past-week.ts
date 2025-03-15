// # Question: "commits in the past week"

// # Constants
// Define the path to the current working directory
const currentWorkingDirectory = Deno.cwd();

// Take an info argument and return information about the script
if (Deno.args.length > 0 && Deno.args[0] === '--info') {
    console.log(JSON.stringify({
        version: '1.0',
        output: 'text'
    }));
    Deno.exit();
}

try {
    // # Load relevant data
    // Run the git log command to get the commit history in the past week
    const gitLogProcess = Deno.run({
        cmd: ['git', 'log', '--since="1 week ago"', '--oneline'],
        cwd: currentWorkingDirectory,
        stdout: "piped",
        stderr: "piped",
    });

    // Read and decode the output of the git log process
    const rawOutput = await gitLogProcess.output();
    const decoder = new TextDecoder();
    const gitLogOutput = decoder.decode(rawOutput);

    // Close the git log process
    gitLogProcess.close();

    // Split the output by newline characters to count the number of commits
    const commitLines = gitLogOutput.split('\n');
    const numCommits = commitLines.length - 1; // Exclude the last empty line

    // # Log results
    // Print the number of commits in the past week
    console.log(numCommits);
} catch (error) {
    // # Log error to stdout and error information to stderr
    console.log('Error');
    console.error('Error retrieving git commit history:', error);
}