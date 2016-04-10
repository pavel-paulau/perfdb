#!/bin/bash -ex

rm -fr build

for os in "windows" "linux" "darwin"; do
    ext=""
    if [ ${os} = "windows" ]; then
        ext=".exe"
    fi

    mkdir -p build/${os}
    for arch in "386" "amd64"; do
        mkdir build/${os}/${arch}
        GOOS=${os} GOARCH=${arch} go build -a -tags netgo --ldflags '-extldflags "-lm -lstdc++ -static"' -o build/${os}/${arch}/perfdb${ext}
    done
done
