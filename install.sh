#!/usr/bin/env sh
set -eu

REPO="Q42/sqill"
BINARY="sqill"

uname_os() {
	os="$(uname -s | tr '[:upper:]' '[:lower:]')"
	case "$os" in
		linux) echo linux ;;
		darwin) echo darwin ;;
		*) echo "error: unsupported OS: $os" >&2; exit 1 ;;
	esac
}

uname_arch() {
	arch="$(uname -m)"
	case "$arch" in
		x86_64 | amd64) echo amd64 ;;
		arm64 | aarch64) echo arm64 ;;
		*) echo "error: unsupported architecture: $arch" >&2; exit 1 ;;
	esac
}

OS="$(uname_os)"
ARCH="$(uname_arch)"
ASSET="${BINARY}_${OS}_${ARCH}.tar.gz"

if [ "${1:-}" = "--version" ] && [ "${2:-}" != "" ]; then
	TAG="$2"
else
	TAG="$(curl -fsS -o /dev/null -w '%{redirect_url}' "https://github.com/${REPO}/releases/latest" \
		| sed 's|.*/tag/||')"
	[ -n "$TAG" ] || { echo "error: could not determine latest release" >&2; exit 1; }
fi

URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading $ASSET ($TAG)..."
curl -fsSL -o "$TMP/$ASSET" "$URL" \
	|| { echo "error: download failed. Does release $TAG exist with asset $ASSET?" >&2; exit 1; }

tar -xzf "$TMP/$ASSET" -C "$TMP"
[ -f "$TMP/$BINARY" ] || { echo "error: binary not found in archive" >&2; exit 1; }
chmod +x "$TMP/$BINARY"

INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
	if command -v sudo >/dev/null 2>&1; then
		echo "Installing to $INSTALL_DIR (requires sudo)..."
		sudo install -m 0755 "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
	else
		INSTALL_DIR="$HOME/.local/bin"
		mkdir -p "$INSTALL_DIR"
		echo "Installing to $INSTALL_DIR (no write access to /usr/local/bin)..."
		install -m 0755 "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
	fi
else
	install -m 0755 "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
fi

case ":$PATH:" in
	*":$INSTALL_DIR:"*) ;;
	*)
		echo ""
		echo "Installed to $INSTALL_DIR but it is not on your PATH."
		echo "Add it with:"
		echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
		;;
esac

echo "$BINARY $TAG installed at $INSTALL_DIR/$BINARY"
