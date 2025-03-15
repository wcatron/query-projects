// Take an info argument and return information about the script
if (Deno.args.length > 0 && Deno.args[0] === '--info') {
  console.log(JSON.stringify({
      version: '1.0',
      output: 'text'
  }));
  Deno.exit();
}

try {
  const projectPath = Deno.cwd(); // Get the current working directory path

  // Use local git to get the last commit date
  const gitCommand = new Deno.Command("git", {
    args: ["log", "-1", "--format=%cd"],
    cwd: projectPath,
    stdout: "piped",
  });

  // Execute the git command to get the last commit date
  const { stdout, code } = await gitCommand.output();

  if (code === 0) {
    const lastCommitDateText = new TextDecoder().decode(stdout).trim();
    const lastCommitDate = new Date(lastCommitDateText);
    const currentDate = new Date();

    const daysSinceLastCommit = Math.floor((currentDate.getTime() - lastCommitDate.getTime()) / (1000 * 60 * 60 * 24));
    console.log(`Last commit was ${daysSinceLastCommit} days ago.`);
  } else {
    console.log('Failed to retrieve last commit information from local Git.');
  }
} catch (error) {
console.log('Error')
  console.error('Error fetching Git history information:', error);
}
