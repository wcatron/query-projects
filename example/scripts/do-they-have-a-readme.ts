import { setupScript } from "jsr:@query-projects/scripts";

setupScript({ type: 'text' }, async () => {
  // Define the path to the README.md file
  const readmePath = './README.md';

  // Check if the README.md file exists in the current working directory
  const readmeExists = await Deno.stat(readmePath).then(() => true).catch(() => false);

  return readmeExists ? 'Yes' : 'No';
});