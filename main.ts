#!/usr/bin/env deno run --allow-read --allow-write --allow-net --allow-run

import { existsSync } from "https://deno.land/std/fs/mod.ts";

import { ensureDirSync } from "https://deno.land/std/fs/mod.ts";
import { dirname } from "https://deno.land/std@0.224.0/path/dirname.ts";
import { join } from "https://deno.land/std@0.224.0/path/join.ts";

// Placeholder functions for complex operations
function openPullRequest(message: string) {
    // TODO: Implement GitHub PR logic using GitHub API or `git`
    console.log(`Opening a pull request with message: "${message}"`);
}

/**
 * Clone a repository or pull the latest changes if the repository already exists.
 * @param repoUrl The URL of the Git repository.
 * @param projectPath The path where the repository should be cloned.
 * @returns The path of the project.
 */
async function cloneRepository(
    repoUrl: string,
    projectPath: string,
): Promise<string> {
    if (existsSync(projectPath)) {
        console.log(`Directory already exists at ${projectPath}`);

        // Check if the directory is a Git repository
        const gitDirExists = existsSync(`${projectPath}/.git`);
        if (gitDirExists) {
            console.log(
                `Repository already cloned. Pulling latest changes from ${repoUrl}...`,
            );

            // Run git pull to get the latest changes
            const pullCommand = new Deno.Command("git", {
                args: ["-C", projectPath, "pull"],
                stdout: "piped",
                stderr: "piped",
            });

            const pullProcess = pullCommand.spawn();
            const { success, stdout, stderr } = await pullProcess.output();

            const pullOutput = new TextDecoder().decode(stdout);
            const pullError = new TextDecoder().decode(stderr);

            if (success) {
                console.log(`Pulled latest changes:\n${pullOutput}`);
            } else {
                console.error(`Error pulling repository:\n${pullError}`);
            }

            return projectPath;
        } else {
            console.error(`Directory exists but is not a Git repository.`);
            Deno.exit(1);
        }
    } else {
        console.log(`Cloning repository from ${repoUrl} to ${projectPath}...`);

        const cloneCommand = new Deno.Command("git", {
            args: ["clone", repoUrl, projectPath],
            stdout: "piped",
            stderr: "piped",
        });

        const cloneProcess = cloneCommand.spawn();
        const { stderr, stdout, success } = await cloneProcess.output();
        const cloneOutput = new TextDecoder().decode(stdout);
        const cloneError = new TextDecoder().decode(stderr);

        if (success) {
            console.log(`Repository successfully cloned:\n${cloneOutput}`);
        } else {
            console.error(`Error cloning repository:\n${cloneError}`);
            Deno.exit(1);
        }

        return projectPath;
    }
}

import { Table } from "https://deno.land/x/cliffy@v1.0.0-rc.4/table/mod.ts";

/**
 * Run the given script for the project at the specified path, display status, and save output to files.
 * @param script The path of the script to run.
 * @param projectPath The project directory where the script will be executed.
 * @returns The status and output for table export.
 */
async function runScriptForProject(
    script: string,
    projectPath: string,
): Promise<
    {
        status: string;
        projectPath: string;
        outputFilePath: string;
        stdoutText: string;
        errorFilePath: string;
    }
> {
    // Define paths to store stdout and stderr outputs
    const filename = script.split("/").pop() || "unknown";
    const resultsFilenameForScript = filename.replace(/\//g, "_").replace(
        ".ts",
        "",
    );
    // Ensure the results directory exists
    const resultsDir = `./results/${resultsFilenameForScript}`;
    ensureDirSync(resultsDir);

    const outputFilePath = join(
        resultsDir,
        `${projectPath.replace(/\//g, "_")}_out.log`,
    );
    const errorFilePath = join(
        resultsDir,
        `${projectPath.replace(/\//g, "_")}_err.log`,
    );

    // Log the status of the project being executed
    console.log(`Running ${script} for ${projectPath}...`);

    try {
        // Run the script using Deno.Command and pipe stdout and stderr to files and live display
        const runCommand = new Deno.Command("deno", {
            args: [
                "run",
                "--allow-read",
                "--allow-run",
                join(Deno.cwd(), script),
            ],
            cwd: projectPath,
            stdout: "piped",
            stderr: "piped",
        });

        // Execute the command and collect the outputs
        const { code, stdout, stderr } = await runCommand.output();

        const stdoutText = new TextDecoder().decode(stdout);
        const stderrText = new TextDecoder().decode(stderr);

        // Write stdout and stderr to respective files
        await Deno.writeTextFile(outputFilePath, stdoutText);
        await Deno.writeTextFile(errorFilePath, stderrText);

        // Display the stdout and stderr live in the terminal
        if (stdoutText) console.log(`[${projectPath}] stdout:\n${stdoutText}`);
        if (stderrText) {
            console.error(`[${projectPath}] stderr:\n${stderrText}`);
        }

        if (code === 0) {
            console.log(`Successfully ran ${script} for ${projectPath}`);
            return {
                status: "Success",
                projectPath,
                outputFilePath,
                stdoutText,
                errorFilePath,
            };
        } else {
            console.log(
                `Script ${script} failed for ${projectPath} with exit code ${code}`,
            );
            return {
                status: "Failed",
                projectPath,
                outputFilePath,
                stdoutText,
                errorFilePath,
            };
        }
    } catch (error: any) {
        console.log(
            `Error running ${script} for ${projectPath}: ${error.message}`,
        );
        await Deno.writeTextFile(errorFilePath, error.message);
        return {
            status: "Error",
            projectPath,
            outputFilePath,
            stdoutText: "",
            errorFilePath,
        };
    }
}

/**
 * Runs a script across multiple projects in parallel with live status, result piping, and export table summary.
 * @param script The script to run.
 * @param projects Array of project paths where the script will be executed.
 */
async function runScriptsForAllProjects(
    script: string,
    projects: string[],
): Promise<void> {
    const table = new Table()
        .header([
            "Project Path",
            "Status",
            "Output",
            "Output File",
            "Error File",
        ])
        .border(true);

    const data: any[] = [];
    const tasks = projects.map(async (projectPath) => {
        const result = await runScriptForProject(script, projectPath);
        return [
            result.projectPath,
            result.status,
            result.stdoutText.trim(),
            result.outputFilePath,
            result.errorFilePath,
        ];
    });

    // Run all tasks in parallel
    const results = await Promise.all(tasks);

    table.body(results);

    // Render the final table with status of all projects
    table.render();

    // Save the table to a file
    const resultsDir = `./results`;
    // Pull just the filename from the script path and replace / with _ and remove
    // the .ts extension
    const filename = script.split("/").pop() || "unknown";
    const resultsFilenameForScript = filename.replace(/\//g, "_").replace(
        ".ts",
        "",
    );
    ensureDirSync(resultsDir);
    const tableFilePath = `${resultsDir}/${resultsFilenameForScript}.md`;
    // Convert the results to markdown format
    const headers = [
        "Project Path",
        "Status",
        "Output",
        "Output File",
        "Error File",
    ];
    const markdownTable = [
        `| ${headers.join(" | ")} |`,
        `| ${headers.map(() => "---").join(" | ")} |`,
        ...results.map((row) => `| ${row.join(" | ")} |`),
    ].join("\n");

    // Write the table to a markdown file
    await Deno.writeTextFile(tableFilePath, markdownTable);
}

/**
 * Extracts TypeScript code from a string containing code wrapped in markdown-style code blocks.
 * @param response The string containing the TypeScript code block.
 * @returns The extracted TypeScript code.
 */
function extractTypeScriptCode(response: string): string | null {
    // Regex pattern to match TypeScript code block in markdown (```typescript ... ```)
    const codeBlockRegex = /```typescript\n([\s\S]*?)\n```/;

    // Search for the TypeScript code block in the response
    const match = response.match(codeBlockRegex);

    // If a match is found, return the extracted code, otherwise return null
    return match ? match[1].trim() : null;
}

// Replace with your OpenAI API key
const OPENAI_API_KEY = Deno.env.get("OPENAI_API_KEY");
if (!OPENAI_API_KEY) {
    console.error("Please set the OPENAI_API_KEY environment variable.");
    Deno.exit(1);
}

// Function to call OpenAI API and generate a script based on the question
async function generateScriptForQuestion(question: string, scriptName: string) {
    // Get path relevant to this script, not the cwd
    const __dirname = dirname(new URL(import.meta.url).pathname);
    const pathToPrompt = join(__dirname, "QUERY_PROMPT.md");
    const prompt = Deno.readTextFileSync(pathToPrompt).replace(
        "{{QUESTION}}",
        question,
    );

    console.log("Generating script based on the question..." + prompt);
    // Call OpenAI to generate the script
    const openAIResponse = await fetch(
        "https://api.openai.com/v1/chat/completions",
        {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${OPENAI_API_KEY}`,
            },
            body: JSON.stringify({
                model: "gpt-3.5-turbo",
                messages: [
                    {
                        role: "system",
                        content:
                            "You are a helpful assistant who writes scripts for Deno projects.",
                    },
                    { role: "user", content: prompt },
                ],
                max_tokens: 1500, // You can adjust this based on the expected script length
                temperature: 0.7, // Adjust creativity level if needed
            }),
        },
    );

    if (!openAIResponse.ok) {
        console.error("Failed to generate script from OpenAI.");
        console.error(await openAIResponse.text());
        Deno.exit(1);
    }

    const responseData = await openAIResponse.json();
    console.log("Generated script:", responseData);
    const generatedScript = extractTypeScriptCode(
        responseData.choices[0].message.content,
    );

    if (generatedScript === null) {
        console.error("Failed to extract TypeScript code from the response.");
        Deno.exit(1);
    }

    // Save the generated script to the 'scripts' directory
    const scriptPath = `./scripts/${scriptName}`;
    Deno.writeTextFileSync(scriptPath, generatedScript, { create: true });
    console.log(`Generated script saved to: ${scriptPath}`);
}

// Load projects.json
function loadProjects() {
    if (existsSync("./projects.json")) {
        return JSON.parse(Deno.readTextFileSync("./projects.json"));
    } else {
        return { projects: [] };
    }
}

// Save projects.json
function saveProjects(projects: any) {
    Deno.writeTextFileSync(
        "./projects.json",
        JSON.stringify(projects, null, 2),
    );
}

// Command: add <repo-url>
async function addRepository(repoUrl: string) {
    const projects = loadProjects();
    const projectName = repoUrl.split("/").pop()?.replace(".git", "") ||
        "unknown-project";
    const projectPath = `./projects/${projectName}`;

    // Clone the repository or add an existing directory
    const clonedPath = cloneRepository(repoUrl, projectPath);

    // Add project to projects.json
    projects.projects.push({ name: projectName, path: projectPath, repoUrl });
    saveProjects(projects);
    console.log(`Added ${projectName} to projects.json.`);
}

// Command: query <question>
async function queryQuestion(question: string) {
    const scriptName = question.toLowerCase().replace(/\s+/g, "-") + ".ts";
    generateScriptForQuestion(question, scriptName);
    console.log(`Generated script: ./scripts/${scriptName}`);
}

// Command: run <script>
async function runScript(scriptName?: string) {
    const projects = loadProjects();
    const scriptPaths: string[] = [];

    if (scriptName) {
        scriptPaths.push(`./scripts/${scriptName}`);
    } else {
        for await (const dirEntry of Deno.readDir("./scripts")) {
            if (dirEntry.isFile && dirEntry.name.endsWith(".ts")) {
                scriptPaths.push(`./scripts/${dirEntry.name}`);
            }
        }
    }

    for (const scriptPath of scriptPaths) {
        await runScriptsForAllProjects(
            scriptPath,
            projects.projects.map((project: any) => project.path),
        );
    }
}

// Command: pr <message>
async function createPullRequest(message: string) {
    const projects = loadProjects();
    for (const project of projects.projects) {
        // Open a pull request for each project
        openPullRequest(message);
    }
}

// Main CLI function
async function main() {
    const [command, ...args] = Deno.args;

    switch (command) {
        case "add": {
            const repoUrl = args[0];
            if (!repoUrl) {
                console.error("Please provide a repository URL.");
                Deno.exit(1);
            }
            await addRepository(repoUrl);
            break;
        }
        case "query": {
            const question = args.join(" ");
            if (!question) {
                console.error("Please provide a question.");
                Deno.exit(1);
            }
            await queryQuestion(question);
            break;
        }
        case "run": {
            const scriptName = args[0];
            await runScript(scriptName);
            break;
        }
        case "pr": {
            const message = args.join(" ");
            if (!message) {
                console.error("Please provide a pull request message.");
                Deno.exit(1);
            }
            await createPullRequest(message);
            break;
        }
        default:
            console.error(
                "Unknown command. Available commands: add, query, run, pr",
            );
            Deno.exit(1);
    }
}

// Run the CLI
await main();
