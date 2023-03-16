rm -rf output
set GOOS=windows
set GOARCH=amd64
echo "Building discovery with GOOS=$GOOS GOARCH=$GOARCH"
go build -mod=mod -o output/discovery.exe cli/main.go
