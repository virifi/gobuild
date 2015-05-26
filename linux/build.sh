#!/bin/sh

set -e

if [ $# -ne 1 ]; then
	echo "Usage: $0 <go commit sha1>"
	exit 1
fi

GO_HASH=$1


if [ -d /usr/lib/go ]; then
	rm -rf /usr/lib/go
fi
cp -r /go /usr/lib/go
cd /usr/lib/go/src
git fetch
git checkout $GO_HASH
#git checkout go1.4.2
GOROOT_BOOTSTRAP=/go_amd64 GOOS=linux GOARCH=amd64 GOROOT_FINAL=/usr/lib/go ./all.bash
tar -C /usr/lib -cvzf /out/go.linux-amd64.$GO_HASH.tar.gz go

if [ -d /usr/lib/go ]; then
	rm -rf /usr/lib/go
fi
cp -r /go /usr/lib/go
cd /usr/lib/go/src
git fetch
git checkout $GO_HASH
#git checkout go1.4.2
GOROOT_BOOTSTRAP=/go_386 GOOS=linux GOARCH=386 GOROOT_FINAL=/usr/lib/go ./all.bash
mv /usr/lib/go/bin/linux_386/* /usr/lib/go/bin/
rm -rf /usr/lib/go/bin/linux_386
tar -C /usr/lib -cvzf /out/go.linux-386.$GO_HASH.tar.gz go
