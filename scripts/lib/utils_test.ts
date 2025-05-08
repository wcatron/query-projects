import { assertEquals, assertThrows } from "https://deno.land/std@0.220.1/assert/mod.ts";
import { nodeJS, setupScript, value } from "./utils.ts";

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

Deno.test("NodeJS dependency tracking", () => {
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
  assertEquals(nodeJS.dependency("typescript"), "4.9.0");
  assertEquals(nodeJS.dependency("eslint"), "8.0.0");
  assertEquals(nodeJS.dependency("nonexistent"), null);

  // Test dev dependencies
  assertEquals(nodeJS.devDependency("prettier"), "2.8.0");
  assertEquals(nodeJS.devDependency("jest"), "29.0.0");
  assertEquals(nodeJS.devDependency("nonexistent"), null);

  // Test getting all dependencies
  assertEquals(nodeJS.getDependencies(), {
    typescript: "4.9.0",
    eslint: "8.0.0"
  });

  // Test getting all dev dependencies
  assertEquals(nodeJS.getDevDependencies(), {
    prettier: "2.8.0",
    jest: "29.0.0"
  });

  restoreReadTextFileSync();
  restoreConsole();
});

Deno.test("NodeJS with missing package.json", () => {
  mockConsole();
  mockReadTextFileSync(""); // Simulate file not found

  // Test with empty dependencies
  assertEquals(nodeJS.dependency("typescript"), null);
  assertEquals(nodeJS.devDependency("prettier"), null);
  assertEquals(nodeJS.getDependencies(), {});
  assertEquals(nodeJS.getDevDependencies(), {});

  restoreReadTextFileSync();
  restoreConsole();
});

Deno.test("setupScript with CSV config", () => {
  assertThrows(
    () => setupScript({ type: "csv" }, () => {}),
    Error,
    "CSV output type requires columns to be specified"
  );

  assertThrows(
    () => setupScript({ type: "csv", columns: [] }, () => {}),
    Error,
    "CSV output type requires columns to be specified"
  );

  // Test valid CSV config
  setupScript({ type: "csv", columns: ["name", "version"] }, () => {});
});

Deno.test("setupScript with --info flag", async () => {
  mockConsole();
  
  // Create a new instance of the script with --info flag
  const script = new Function(`
    const { setupScript } = await import("./utils.ts");
    setupScript({ type: "text" }, () => {});
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

Deno.test("setupScript error handling", () => {
  mockConsole();

  setupScript({ type: "text" }, () => {
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