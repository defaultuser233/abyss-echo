#!/bin/sh
set -eu

cd "$(dirname "$0")/.."

VERSION="${VERSION:-1.0.0}"
DIST="${DIST:-dist}"
LDFLAGS="-s -w -X main.version=${VERSION}"

mkdir -p "$DIST"
find "$DIST" -maxdepth 1 -type f -name 'abyss-echo_*' -exec rm -f {} +

build() {
	os="$1"
	arch="$2"
	arm="$3"
	ext=""
	if [ "$os" = "windows" ]; then
		ext=".exe"
	fi
	name="abyss-echo_${VERSION}_${os}_${arch}${ext}"
	printf '  %-16s %s/%s\n' "$name" "$os" "$arch"
	if [ -n "$arm" ]; then
		CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" GOARM="$arm" \
			go build -trimpath -ldflags "$LDFLAGS" -o "$DIST/$name" .
	else
		CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
			go build -trimpath -ldflags "$LDFLAGS" -o "$DIST/$name" .
	fi
}

echo "Building 深渊回响 ${VERSION}"
build linux amd64 ""
build linux arm64 ""
build linux arm 7
build darwin amd64 ""
build darwin arm64 ""
build windows amd64 ""
build windows arm64 ""

(
	cd "$DIST"
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum abyss-echo_* > SHA256SUMS
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 abyss-echo_* > SHA256SUMS
	fi
)

echo
echo "Artifacts:"
ls -lh "$DIST" | sed '1d'
