#!/bin/sh
# Installs the latest ndo release for Linux or macOS.
#
#   curl -fsSL https://raw.githubusercontent.com/NoufalBhai/ndo/main/install.sh | sh
#
# Override the install directory with NDO_INSTALL_DIR (default: /usr/local/bin,
# falling back to ~/.local/bin if that isn't writable).

set -eu

repo="NoufalBhai/ndo"

os=$(uname -s)
case "$os" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *)
    echo "ndo: unsupported OS: $os (releases exist for Linux and macOS only; Windows users should use Scoop)" >&2
    exit 1
    ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "ndo: unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

version=${NDO_VERSION:-}
if [ -z "$version" ]; then
  version=$(curl -fsSL "https://api.github.com/repos/$repo/releases/latest" \
    | grep -m1 '"tag_name"' | cut -d '"' -f4)
  if [ -z "$version" ]; then
    echo "ndo: could not resolve the latest release tag" >&2
    exit 1
  fi
fi

url="https://github.com/$repo/releases/download/$version/ndo_${version#v}_${os}_${arch}.tar.gz"

install_dir=${NDO_INSTALL_DIR:-}
if [ -z "$install_dir" ]; then
  if [ -w /usr/local/bin ]; then
    install_dir=/usr/local/bin
  else
    install_dir="$HOME/.local/bin"
    mkdir -p "$install_dir"
  fi
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "ndo: downloading $url" >&2
curl -fsSL "$url" -o "$tmp/ndo.tar.gz"
tar -xzf "$tmp/ndo.tar.gz" -C "$tmp"
install -m 755 "$tmp/ndo" "$install_dir/ndo"

echo "ndo: installed to $install_dir/ndo" >&2
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "ndo: warning: $install_dir is not on your PATH" >&2 ;;
esac
"$install_dir/ndo" --version
