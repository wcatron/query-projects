// # Question: "Does the project have a linter?"

import { script, packageManager } from "jsr:@query-projects/scripts";

script({ type: 'text' }, () => {
  try {
    // Check if ESLint is installed as a devDependency
    const hasESLint = packageManager.devDependency('eslint') !== undefined;
    return hasESLint ? 'Yes' : 'No';
  } catch (error) {
    console.error('Error checking for linter:', error);
    return 'Error';
  }
});