$release_dir="bin"

If (!(test-path $release_dir)){
    md $release_dir
}

$GitCommit=git log -1 --format="%h"
$Version=git describe --tags --abbrev=0
$BuildAt=(Get-Date).ToString()
$BuildBy=go version
$BUILD_FLAGS="-X 'main.GitCommit=${GitCommit}' -X 'main.Version=${Version}' -X 'main.BuildAt=${BuildAt}' -X 'main.BuildBy=${BuildBy}'"

# echo $BUILD_FLAGS

# go env -w GOOS="linux"
# go build -o $release_dir/fcc -ldflags "$BUILD_FLAGS" .

go env -w GOOS="windows"
go build -o $release_dir/fcc.exe -ldflags "$BUILD_FLAGS" .
