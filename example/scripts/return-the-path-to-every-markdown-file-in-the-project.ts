import { walk } from "https://deno.land/std/fs/walk.ts";
import * as path from 'https://deno.land/std@0.102.0/path/mod.ts';
import { script } from "jsr:@query-projects/scripts";

// # Question: "Return the path to every markdown file in the project"

// # Constants
// Define the current working directory
const currentDirectory = Deno.cwd();

script({ 
  type: 'csv',
  columns: ['path']
}, async (emit) => {
    // Walk through the current directory
    for await (const entry of walk(currentDirectory, {
      includeDirs: false,
      includeFiles: true,
      exts: [".md"],
    })) {
      const rel = path.relative(currentDirectory, entry.path);
      emit([rel]);
    }
});