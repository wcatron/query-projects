#!/bin/bash

# Build the binary
bash ./build.sh

cd example
../query-projects info
../query-projects sync
../query-projects pull
../query-projects run --script scripts/do-they-have-a-readme.ts --output json,csv
../query-projects run --script scripts/does-the-project-have-a-linter.ts --output csv
../query-projects run --script scripts/how-activily-maintained-is-the-project.ts --output md
../query-projects run --script scripts/return-the-path-to-every-markdown-file-in-the-project.ts --output csv
../query-projects run --script scripts/what-version-of-package-is-being-used.ts typescript
../query-projects run --script scripts/which-test-framework-is-being-used.ts
../query-projects run --script scripts/get-compiler-options-from-tsconfig.ts
../query-projects plan plans/test.lua

# Run all scripts in one project
cd projects/ai
../../../query-projects run --all

cd ../../
../query-projects run --all