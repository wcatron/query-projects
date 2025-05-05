print("repoName: " .. repoName)

version = value("package.json", "version")
print("version: " .. version)

majorString = string.match(version, "%d+")
if majorString == nil then
    print("No major version found")
    major = 0
else
    major = tonumber(majorString)
end

if major > 0 then
    tsVersion = run("scripts/what-version-of-typescript-is-being-used.ts")
    print("tsVersion: " .. tsVersion)
    if tsVersion == "5.1.3" then
        run("scripts/get-compiler-options-from-tsconfig.ts")
    end
else
    print("Not over 1 " .. major)
end
