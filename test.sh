#!/bin/bash

# Build the binary
bash ./build.sh

cd example
../query-projects info
../query-projects sync
../query-projects pull
../query-projects run scripts/do-they-have-a-readme.ts --output json
../query-projects run scripts/does-the-project-have-a-linter.ts --output csv
../query-projects run scripts/how-activily-maintained-is-the-project.ts --output md
../query-projects run scripts/return-the-path-to-every-markdown-file-in-the-project.ts --output csv
../query-projects run scripts/what-version-of-typescript-is-being-used.ts
../query-projects run scripts/which-test-framework-is-being-used.ts
../query-projects run scripts/get-compiler-options-from-tsconfig.ts