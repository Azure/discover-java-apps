rm -rf ./bin
GOOS=linux 
GOARCH=amd64
echo "Building discovery with GOOS=$GOOS GOARCH=$GOARCH"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod -o ./bin/discovery-l cli/main.go

CGO_ENABLED=0
GOOS=windows
CC=x86_64-w64-mingw32-gcc 
CXX=x86_64-w64-mingw32-g++ 
GOARCH=amd64
echo "Building discovery with GOOS=$GOOS GOARCH=$GOARCH"
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -mod=mod -o ./bin/discovery.exe cli/main.go

