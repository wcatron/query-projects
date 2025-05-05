rm query-projects
go build -o query-projects ./
# Check if build was successful
if [ $? -ne 0 ]; then
    echo "Build failed"
    exit 1
fi