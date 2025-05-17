import { assertEquals, assertRejects } from "https://deno.land/std@0.220.1/assert/mod.ts";
import { packageManager, script, value } from "./utils.ts";

// Mock console.log


// Mock Deno.readTextFileSync
const originalReadTextFileSync = Deno.readTextFileSync;
function mockReadTextFileSync(content: string) {
  Deno.readTextFileSync = () => content;
}

function restoreReadTextFileSync() {
  Deno.readTextFileSync = originalReadTextFileSync;
}

Deno.test("packageManager dependency tracking", () => {
  packageManager.resetInstance();

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
});

Deno.test("packageManager with missing package.json", () => {
  packageManager.resetInstance();

  mockReadTextFileSync(""); // Simulate file not found

  // Test with empty dependencies
  assertEquals(packageManager.dependency("typescript"), undefined);
  assertEquals(packageManager.devDependency("prettier"), undefined);
  assertEquals(packageManager.getDependencies(), {});
  assertEquals(packageManager.getDevDependencies(), {});

  restoreReadTextFileSync();
});

Deno.test("packageManager peer dependencies", () => {
  packageManager.resetInstance();
  const packageJson = `{
    "peerDependencies": {
      "react": ">=16.8.0",
      "react-dom": ">=16.8.0"
    }
  }`;
  mockReadTextFileSync(packageJson);
  
  assertEquals(packageManager.getPeerDependencies(), {
    "react": ">=16.8.0",
    "react-dom": ">=16.8.0"
  });

  restoreReadTextFileSync();
});

Deno.test("script with CSV config", async () => {
  await assertRejects(
    () => script({ type: "csv" }, () => {
      return [];
    }),
    Error,
    "CSV output type requires columns to be specified"
  );

  await assertRejects(
    () => script({ type: "csv", columns: [] }, () => {
      return [];
    }),
    Error,
    "CSV output type requires columns to be specified"
  );

  // Test valid CSV config
  const result = await script({ type: "csv", columns: ["name", "version"] }, () => {
    return ["some-name", "some-version"];
  }, { captureConsole: true });

  assertEquals(result.stdout, "some-name,some-version\n");
});

Deno.test("script emitter with different types", async () => {
  // Test text emitter
  const resultText = await script({ type: "text" }, (emit) => {
    emit("test string");
  }, { captureConsole: true });
  assertEquals(resultText.stdout, "test string\n");

  // Test json emitter
  const resultJSON = await script({ type: "json" }, (emit) => {
    emit({ key: "value" });
  }, { captureConsole: true });
  assertEquals(resultJSON.stdout, '{\n  "key": "value"\n}\n');

  // Test csv emitter with multiple rows
  const resultCSV = await script({ type: "csv", columns: ["col1", "col2"] }, (emit) => {
    emit(["row1", "row2"]);
    emit(["row3", "row4"]);
  }, { captureConsole: true });
  assertEquals(resultCSV.stdout, "row1,row2\nrow3,row4\n");
});

Deno.test("script error cases", async () => {
  // Test invalid emitter type for text
  const resultText = await script({ type: 'text' }, (emit) => {
    // deno-lint-ignore no-explicit-any
    emit({'a':'b'} as any); // Invalid type
  }, {
    captureConsole: true,
    captureExit: true
  })
  assertEquals(resultText.exitCode, 1)
  assertEquals(resultText.stderr,  '[ERROR] Script execution failed: Row with type "text" is not a string or number\n')

  // Test invalid emitter type for csv
  const resultCSV = await script({ type: "csv", columns: ["col1"] }, (emit) => {
    emit({} as any); // Invalid type
  }, {
    captureConsole: true,
    captureExit: true
  })
  assertEquals(resultCSV.exitCode, 1)
  assertEquals(resultCSV.stderr, '[ERROR] Script execution failed: Row with type "csv" is not an array\n')

  // Test invalid emitter type for csv
  const resultJSON = await script({ type: "json" }, (emit) => {
    emit("not an object" as any); // Invalid type
  }, {
    captureConsole: true,
    captureExit: true
  })
  assertEquals(resultJSON.exitCode, 1)
  assertEquals(resultJSON.stderr, '[ERROR] Script execution failed: Row with type "json" is not an object\n')
});

Deno.test("script error handling", async () => {
  const { stderr, exitCode } = await script({ type: "text" }, () => {
    throw new Error("Test error");
  }, { captureConsole: true, captureExit: true });

  assertEquals(stderr, "[ERROR] Script execution failed: Test error\n");
  assertEquals(exitCode, 1);
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