interface ScriptConfig {
  type: 'csv' | 'json' | 'text';
  columns?: string[];
}

type ScriptReturn<T extends ScriptConfig['type']> = 
  T extends 'csv' ? string[] | undefined | null | void:
  T extends 'json' ? Record<string, unknown> | undefined | null | void:
  T extends 'text' ? string | undefined | null | void:
  never;


type ScriptEmitter<T extends ScriptConfig['type']> = (row: ScriptReturn<T>) => void;

interface PackageJSON {
  dependencies?: Record<string, string>;
  devDependencies?: Record<string, string>;
  peerDependencies?: Record<string, string>;
}

export class PackageManager {
  private static instance: PackageManager | undefined = undefined;
  private _packageJson: PackageJSON | null = null;
  private isLoaded = false;

  static getInstance(): PackageManager {
    if (!PackageManager.instance) {
      PackageManager.instance = new PackageManager();
    }
    return PackageManager.instance;
  }
  resetInstance() {
    this._packageJson = null;
    this.isLoaded = false;
  }

  private get packageJson(): PackageJSON | null {
    if (this.isLoaded) {
      return this._packageJson;
    }
    try {
      const content = Deno.readTextFileSync('package.json');
      this._packageJson = JSON.parse(content);
      this.isLoaded = true;
      return this._packageJson;
    } catch (_) {
      // If the package.json file is not found, use an empty object
      // This is to avoid errors when the package.json file is not found
      // and to allow the script to run without errors
      this._packageJson = {};
      this.isLoaded = true;
      console.debug('package.json could not be read, using empty object');
      return this._packageJson;
    }
  }

  dependency(name: string): string | undefined {
    return this.packageJson?.dependencies?.[name];
  }

  devDependency(name: string): string | undefined {
    return this.packageJson?.devDependencies?.[name]
  }

  getDependencies(): Record<string, string> {
    return this.packageJson?.dependencies || {};
  }

  getDevDependencies(): Record<string, string> {
    return this.packageJson?.devDependencies || {};
  }

  // Peer dependencies
  getPeerDependencies(): Record<string, string> {
    return this.packageJson?.peerDependencies || {};
  }
}
export const packageManager: PackageManager = PackageManager.getInstance();

function emitter<T extends ScriptConfig['type']>(
  type: T,
  console: Console
): (row: ScriptReturn<T>) => void {
  return (row) => {
    if (type === 'text') {
      if (typeof row === 'string' || typeof row === 'number') {
        console.log(row);
        return;
      } else {
        throw new Error('Row with type "text" is not a string or number');
      }
    } else if (type === 'csv') {
      if (Array.isArray(row)) {
        console.log(row.join(','));
        return;
      } else {
        throw new Error('Row with type "csv" is not an array');
      }
    } else if (type === 'json') {
      if (typeof row === 'object' && row !== null) {
        console.log(JSON.stringify(row, null, 2));
        return;
      } else {
        throw new Error('Row with type "json" is not an object');
      }
    }
  };
}


export async function script<T extends ScriptConfig['type']>(
  config: ScriptConfig & { type: T },
  script: (emit: ScriptEmitter<T>) => ScriptReturn<T> | Promise<ScriptReturn<T>>,
  options?: {
    // If true, the console output will be captured and returned
    // This is useful for testing and should not be used in production
    captureConsole?: boolean;
    captureExit?: boolean;
  }
): Promise<{
  stdout?: string;
  stderr?: string;
  exitCode?: number;
}> {
  // Validate config
  if (config.type === 'csv' && (!config.columns || config.columns.length === 0)) {
    throw new Error('CSV output type requires columns to be specified');
  }

  let stdout = '';
  let stderr = '';
  let exitCode = 0;
  const captureConsole = options?.captureConsole || false;
  const console = captureConsole ? {
    log: (...args: any[]) => {
      stdout += args.join(' ') + '\n';
    },
    error: (...args: any[]) => {
      stderr += args.join(' ') + '\n';
    },
  } as Console : globalThis.console;
  
  const exit = (code: number) => {
    if (options?.captureExit) {
      exitCode = code;
    } else {
      Deno.exit(exitCode);
    }
  }

  // Add --info flag handling
  if (Deno.args.includes('--info')) {
    console.log(JSON.stringify({
      version: '1.0.0',
      output: config.type,
      columns: config.columns || [],
    }));
    Deno.exit(0);
  }

  try {
    const emit = emitter(config.type, console);
    const result = await script(emit);
    if (result) {
      emit(result);
    }
  } catch (err) {
    const error = err as Error;
    console.error(`[ERROR] Script execution failed: ${error.message}`);
    exit(1);
  }

  return { stdout, stderr, exitCode };
}

export function value(filename: string, fieldAccessor: string): string | number | null {
  try {
    const content = Deno.readTextFileSync(filename);
    const ext = filename.split('.').pop()?.toLowerCase();

    if (ext === 'json') {
      const data = JSON.parse(content);
      return getNestedValue(data, fieldAccessor);
    } else if (ext === 'xml') {
      // For XML, we'll use a simple regex-based approach
      // This is a basic implementation - you might want to use a proper XML parser
      const matches = content.match(new RegExp(`<${fieldAccessor}[^>]*>(.*?)</${fieldAccessor}>`, 's'));
      return matches ? matches[1].trim() : null;
    } else {
      throw new Error(`Unsupported file type: ${ext}`);
    }
  } catch (err) {
    const error = err as Error;
    console.error(`[ERROR] Failed to read value from ${filename}: ${error.message}`);
    return null;
  }
}

// deno-lint-ignore no-explicit-any
function getNestedValue(obj: any, path: string): any {
  return path.split('.').reduce((current, key) => {
    return current && current[key] !== undefined ? current[key] : null;
  }, obj);
}
