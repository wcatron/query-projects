import { walk } from "https://deno.land/std/fs/walk.ts";
import * as path from 'https://deno.land/std@0.102.0/path/mod.ts';

// # Question: "Return the path to every markdown file in the project"

// # Constants
// Define the current working directory
const currentDirectory = Deno.cwd();

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
    // Use Deno's walk function to recursively walk through the current directory
    for await (const entry of walk(currentDirectory, {
        includeDirs: false,
        includeFiles: true,
        exts: [".md"],
      })) {
        const rel = path.relative(currentDirectory, entry.path);
        console.log(rel);
    }
} catch (error) {
    // # Log error to stdout and error information to stderr
    console.log('Error');
    console.error('Error reading directory:', error);
}