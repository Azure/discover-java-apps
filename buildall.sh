rm -rf ./output
GOOS=linux 
GOARCH=amd64
echo "Building discovery with GOOS=$GOOS GOARCH=$GOARCH"
go build -mod=mod -o ./output/discovery-l cli/main.go

GOOS=windows 
CGO_ENABLED=1 
CC=x86_64-w64-mingw32-gcc 
CXX=x86_64-w64-mingw32-g++ 
GOARCH=amd64
echo "Building discovery with GOOS=$GOOS GOARCH=$GOARCH"
go build -mod=mod -o ./output/discovery.exe cli/main.go
chmod +x ./output/discovery.exe

