import { assertEquals } from "https://deno.land/std@0.224.0/testing/asserts.ts";

function extractTypeScriptCode(response: string): string | null {
    const codeBlockRegex = /```(?:typescript|ts)\s+([\s\S]+?)\s*```/;
    const match = response.match(codeBlockRegex);
    return match ? match[1].trim() : null;
}

Deno.test("extractTypeScriptCode - valid TypeScript block", () => {
    const response = "Here is some code:\n```typescript\nconst x = 10;\n```";
    const expected = "const x = 10;";
    const result = extractTypeScriptCode(response);
    assertEquals(result, expected);
});

Deno.test("extractTypeScriptCode - valid ts block", () => {
    const response = "Here is some code:\n```ts\nlet y = 20;\n```";
    const expected = "let y = 20;";
    const result = extractTypeScriptCode(response);
    assertEquals(result, expected);
});

Deno.test("extractTypeScriptCode - no TypeScript block", () => {
    const response = "Here is some text without code blocks.";
    const result = extractTypeScriptCode(response);
    assertEquals(result, null);
});

Deno.test("extractTypeScriptCode - empty TypeScript block", () => {
    const response = "Here is an empty code block:\n```typescript\n\n```";
    const expected = "";
    const result = extractTypeScriptCode(response);
    assertEquals(result, expected);
});
