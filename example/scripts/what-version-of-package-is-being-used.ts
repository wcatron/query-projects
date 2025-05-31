
import { script, packageManager } from "jsr:@query-projects/scripts";

script({ type: 'text' }, () => {
  const packageStr = Deno.args[0];
  
  return packageManager.dependency(packageStr) || packageManager.devDependency(packageStr);
});
