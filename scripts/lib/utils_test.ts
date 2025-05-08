import { assertEquals, assertThrows } from "https://deno.land/std@0.220.1/assert/mod.ts";
import { packageManager, script, value } from "./utils.ts";

// Mock console.log
let consoleOutput: string[] = [];
const originalConsoleLog = console.log;
const originalConsoleError = console.error;

function mockConsole() {
  consoleOutput = [];
  console.log = (...args: unknown[]) => {
    consoleOutput.push(args.map(String).join(" "));
  };
  console.error = (...args: unknown[]) => {
    consoleOutput.push(`ERROR: ${args.map(String).join(" ")}`);
  };
}

function restoreConsole() {
  console.log = originalConsoleLog;
  console.error = originalConsoleError;
}

// Mock Deno.readTextFileSync
const originalReadTextFileSync = Deno.readTextFileSync;
function mockReadTextFileSync(content: string) {
  Deno.readTextFileSync = () => content;
}

function restoreReadTextFileSync() {
  Deno.readTextFileSync = originalReadTextFileSync;
}

Deno.test("packageManager dependency tracking", () => {
  mockConsole();
  const packageJson = `{
    "dependencies": {
      "typescript": "4.9.0",
      "eslint": "8.0.0"
    },
    "devDependencies": {
      "prettier": "2.8.0",
      "jest": "29.0.0"
    }
  }`;
  mockReadTextFileSync(packageJson);

  // Test regular dependencies
  assertEquals(packageManager.dependency("typescript"), "4.9.0");
  assertEquals(packageManager.dependency("eslint"), "8.0.0");
  assertEquals(packageManager.dependency("nonexistent"), undefined);

  // Test dev dependencies
  assertEquals(packageManager.devDependency("prettier"), "2.8.0");
  assertEquals(packageManager.devDependency("jest"), "29.0.0");
  assertEquals(packageManager.devDependency("nonexistent"), undefined);

  // Test getting all dependencies
  assertEquals(packageManager.getDependencies(), {
    typescript: "4.9.0",
    eslint: "8.0.0"
  });

  // Test getting all dev dependencies
  assertEquals(packageManager.getDevDependencies(), {
    prettier: "2.8.0",
    jest: "29.0.0"
  });

  restoreReadTextFileSync();
  restoreConsole();
});

Deno.test("packageManager with missing package.json", () => {
  mockConsole();
  mockReadTextFileSync(""); // Simulate file not found

  // Test with empty dependencies
  assertEquals(packageManager.dependency("typescript"), undefined);
  assertEquals(packageManager.devDependency("prettier"), undefined);
  assertEquals(packageManager.getDependencies(), {});
  assertEquals(packageManager.getDevDependencies(), {});

  restoreReadTextFileSync();
  restoreConsole();
});

Deno.test("script with CSV config", () => {
  assertThrows(
    () => script({ type: "csv" }, () => {
      return [];
    }),
    Error,
    "CSV output type requires columns to be specified"
  );

  assertThrows(
    () => script({ type: "csv", columns: [] }, () => {
      return [];
    }),
    Error,
    "CSV output type requires columns to be specified"
  );

  // Test valid CSV config
  script({ type: "csv", columns: ["name", "version"] }, () => {
    return ["some-name", "some-version"];
  });
});

Deno.test("script with --info flag", async () => {
  mockConsole();
  
  // Create a new instance of the script with --info flag
  const script = new Function(`
    const { script } = await import("./utils.ts");
    script({ type: "text" }, () => {});
  `);

  // Mock Deno.args for this test
  const originalArgs = Deno.args;
  Object.defineProperty(Deno, "args", {
    value: ["--info"],
    configurable: true
  });

  await script();

  assertEquals(JSON.parse(consoleOutput[0]), {
    version: "1.0.0",
    output: "text",
    columns: [],
  });

  // Restore Deno.args
  Object.defineProperty(Deno, "args", {
    value: originalArgs,
    configurable: true
  });

  restoreConsole();
});

Deno.test("script error handling", () => {
  mockConsole();

  script({ type: "text" }, () => {
    throw new Error("Test error");
  });

  assertEquals(consoleOutput[0], "ERROR: [ERROR] Script execution failed: Test error");

  restoreConsole();
});

Deno.test("value function with JSON", () => {
  const jsonContent = `{
    "dependencies": {
      "typescript": {
        "version": "4.9.0",
        "type": "dev"
      }
    }
  }`;

  mockReadTextFileSync(jsonContent);

  // Test simple field access
  assertEquals(value("package.json", "dependencies"), {
    typescript: {
      version: "4.9.0",
      type: "dev"
    }
  });

  // Test nested field access
  assertEquals(value("package.json", "dependencies.typescript.version"), "4.9.0");
  assertEquals(value("package.json", "dependencies.typescript.type"), "dev");

  // Test non-existent field
  assertEquals(value("package.json", "dependencies.nonexistent"), null);
  assertEquals(value("package.json", "dependencies.typescript.nonexistent"), null);

  restoreReadTextFileSync();
});

Deno.test("value function with XML", () => {
  const xmlContent = `
    <settings>
      <debug>true</debug>
      <version>1.0.0</version>
    </settings>
  `;

  mockReadTextFileSync(xmlContent);

  // Test XML field access
  assertEquals(value("config.xml", "debug"), "true");
  assertEquals(value("config.xml", "version"), "1.0.0");

  // Test non-existent field
  assertEquals(value("config.xml", "nonexistent"), null);

  restoreReadTextFileSync();
});

Deno.test("value function with unsupported file type", () => {
  const content = "some content";
  mockReadTextFileSync(content);

  // Test unsupported file type
  assertEquals(value("file.txt", "field"), null);

  restoreReadTextFileSync();
}); 