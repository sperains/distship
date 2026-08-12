#!/bin/sh

set -eu

repository="sperains/distship"
requested_version=""
install_dir="${DISTSHIP_INSTALL_DIR:-${HOME}/.local/bin}"

usage() {
    cat <<'EOF'
Install DistShip from an official GitHub Release.

Usage:
  install.sh [--version <vX.Y.Z>] [--install-dir <absolute-path>]

Options:
  --version      Install a specific version. Defaults to the latest stable release.
  --install-dir  Install into this directory. Defaults to $HOME/.local/bin.
  -h, --help     Show this help.
EOF
}

fail() {
    printf 'error: %s\n' "$1" >&2
    exit 1
}

require_command() {
    command -v "$1" >/dev/null 2>&1 || fail "required command is not available: $1"
}

while [ "$#" -gt 0 ]; do
    case "$1" in
        --version)
            [ "$#" -ge 2 ] || fail "--version requires a value"
            requested_version="$2"
            shift 2
            ;;
        --install-dir)
            [ "$#" -ge 2 ] || fail "--install-dir requires a value"
            install_dir="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            fail "unknown option: $1"
            ;;
    esac
done

case "$install_dir" in
    /*) ;;
    *) fail "install directory must be an absolute path: $install_dir" ;;
esac

require_command curl
require_command install
require_command tar
require_command uname

case "$(uname -s)" in
    Darwin)
        platform="Darwin"
        require_command shasum
        checksum_command="shasum"
        ;;
    Linux)
        platform="Linux"
        require_command sha256sum
        checksum_command="sha256sum"
        ;;
    *)
        fail "unsupported operating system: $(uname -s)"
        ;;
esac

case "$(uname -m)" in
    x86_64|amd64)
        architecture="x86_64"
        ;;
    arm64|aarch64)
        architecture="arm64"
        ;;
    *)
        fail "unsupported architecture: $(uname -m)"
        ;;
esac

if [ -n "$requested_version" ]; then
    case "$requested_version" in
        v*) version_tag="$requested_version" ;;
        *) version_tag="v$requested_version" ;;
    esac
else
    latest_url="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/${repository}/releases/latest")"
    version_tag="${latest_url##*/}"
fi

version="${version_tag#v}"
printf '%s\n' "$version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$' || fail "invalid release version: $version_tag"

archive="distship_${version}_${platform}_${architecture}.tar.gz"
release_base="https://github.com/${repository}/releases/download/${version_tag}"
temporary_dir="$(mktemp -d "${TMPDIR:-/tmp}/distship-install.XXXXXX")"
staged_binary=""

cleanup() {
    if [ -n "$staged_binary" ]; then
        rm -f -- "$staged_binary"
    fi
    rm -rf -- "$temporary_dir"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM

printf 'Downloading DistShip %s for %s/%s...\n' "$version" "$platform" "$architecture"
curl -fL --retry 3 --silent --show-error --output "$temporary_dir/$archive" "$release_base/$archive"
curl -fL --retry 3 --silent --show-error --output "$temporary_dir/checksums.txt" "$release_base/checksums.txt"

awk -v archive="$archive" '$2 == archive { print }' "$temporary_dir/checksums.txt" > "$temporary_dir/archive.checksum"
[ -s "$temporary_dir/archive.checksum" ] || fail "release checksum is missing for $archive"
[ "$(wc -l < "$temporary_dir/archive.checksum" | tr -d ' ')" = "1" ] || fail "release checksum is ambiguous for $archive"

(
    cd "$temporary_dir"
    if [ "$checksum_command" = "shasum" ]; then
        shasum -a 256 -c archive.checksum
    else
        sha256sum -c archive.checksum
    fi
    tar -xzf "$archive"
)

[ -f "$temporary_dir/distship" ] || fail "release archive does not contain the distship binary"
chmod 0755 "$temporary_dir/distship"
actual_version="$($temporary_dir/distship version | sed -n '1p')"
[ "$actual_version" = "distship $version" ] || fail "binary version mismatch: expected $version, got $actual_version"

mkdir -p "$install_dir"
[ -d "$install_dir" ] || fail "install directory is not available: $install_dir"
[ -w "$install_dir" ] || fail "install directory is not writable: $install_dir"

staged_binary="$install_dir/.distship.new.$$"
install -m 0755 "$temporary_dir/distship" "$staged_binary"
mv -f "$staged_binary" "$install_dir/distship"
staged_binary=""

printf 'Installed %s\n' "$install_dir/distship"
"$install_dir/distship" version

case ":${PATH}:" in
    *":${install_dir}:"*) ;;
    *) printf 'Add %s to PATH before running distship by name.\n' "$install_dir" ;;
esac
