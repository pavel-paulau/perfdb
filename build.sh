#!/bin/bash -ex

rm -fr build

for os in "windows" "linux" "darwin"; do
    ext=""
    if [ ${os} = "windows" ]; then
        ext=".exe"
    fi

    for arch in "386" "amd64"; do
        CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -a -tags netgo --ldflags '-s' -o build/perfdb.${os}-${arch}${ext}
    done
done
