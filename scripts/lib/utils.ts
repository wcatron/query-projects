interface ScriptConfig {
  type: 'csv' | 'json' | 'text';
  columns?: string[];
}

type ScriptReturn<T extends ScriptConfig['type']> = 
  T extends 'csv' ? string[] :
  T extends 'json' ? Record<string, unknown> :
  T extends 'text' ? string :
  never;

interface PackageJSON {
  dependencies?: Record<string, string>;
  devDependencies?: Record<string, string>;
}

class PackageManager {
  private static instance: PackageManager;
  private _packageJson: PackageJSON | null = null;
  private isLoaded = false;

  static getInstance(): PackageManager {
    if (!PackageManager.instance) {
      PackageManager.instance = new PackageManager();
    }
    return PackageManager.instance;
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
    } catch (err) {
      // If the package.json file is not found, use an empty object
      // This is to avoid errors when the package.json file is not found
      // and to allow the script to run without errors
      this._packageJson = {};
      this.isLoaded = true;
      console.warn('package.json not found, using empty object');
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
}

export const packageManager: PackageManager = PackageManager.getInstance();

export async function script<T extends ScriptConfig['type']>(
  config: ScriptConfig & { type: T },
  script: () => ScriptReturn<T> | Promise<ScriptReturn<T>>
): Promise<void> {
  // Validate config
  if (config.type === 'csv' && (!config.columns || config.columns.length === 0)) {
    throw new Error('CSV output type requires columns to be specified');
  }

  // Add --info flag handling
  if (Deno.args.includes('--info')) {
    const info = {
      version: '1.0.0',
      output: config.type,
      columns: config.columns || [],
    };
    console.log(JSON.stringify(info));
    Deno.exit(0);
  }

  try {
    const result = await script();
    switch (config.type) {
      case 'csv': {
        const csvResult = result as string[];
        console.log(csvResult.join(','));
        break;
      }
      case 'json': {
        const jsonResult = result as Record<string, unknown>;
        console.log(JSON.stringify(jsonResult, null, 2));
        break;
      }
      case 'text': {
        const textResult = result as string;
        console.log(textResult);
        break;
      }
    }
  } catch (err) {
    const error = err as Error;
    console.error(`[ERROR] Script execution failed: ${error.message}`);
    Deno.exit(1);
  }
}

export function value(filename: string, fieldAccessor: string): any {
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

function getNestedValue(obj: any, path: string): any {
  return path.split('.').reduce((current, key) => {
    return current && current[key] !== undefined ? current[key] : null;
  }, obj);
}
