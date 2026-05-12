#!/usr/bin/env bash
#
# mlb-mcp installer
# Usage: curl -fsSL https://github.com/retr0h/mlb-mcp/raw/main/install.sh | bash
#
# Env overrides:
#   MLB_MCP_VERSION       install a specific version (e.g. 1.0.0) instead of latest
#   MLB_MCP_INSTALL_DIR   force install destination, skipping the default rules

set -euo pipefail
APP=mlb-mcp

# MLB blue accent (#002D72)
MUTED='\033[0;2m'
RED='\033[0;31m'
ACCENT='\033[38;2;0;45;114m'
NC='\033[0m'

err() {
    printf "${RED}mlb-mcp: %s${NC}\n" "$1" >&2
    exit 1
}

print_message() {
    local level=$1
    local message=$2
    local color=""
    case $level in
        info)    color="${NC}" ;;
        warning) color="${ACCENT}" ;;
        error)   color="${RED}" ;;
    esac
    printf "${color}${message}${NC}\n"
}

have() {
    command -v "$1" >/dev/null 2>&1
}

detect_os() {
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$os" in
        linux|darwin) ;;
        *) err "unsupported OS: $os" ;;
    esac
}

detect_arch() {
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) err "unsupported architecture: $arch" ;;
    esac
}

resolve_version() {
    if [ -n "${MLB_MCP_VERSION:-}" ]; then
        version="$MLB_MCP_VERSION"
    else
        version="$(curl -fsSL "https://api.github.com/repos/retr0h/$APP/releases/latest" \
            | grep '"tag_name"' | sed 's/.*"v\(.*\)".*/\1/')" \
            || err "cannot determine latest version"
    fi
}

resolve_install_dir() {
    needs_symlink=0
    if [ -n "${MLB_MCP_INSTALL_DIR:-}" ]; then
        install_dir="$MLB_MCP_INSTALL_DIR"
    elif [ -d "$HOME/.local/bin" ] && path_contains "$HOME/.local/bin"; then
        install_dir="$HOME/.local/bin"
    elif [ -d "$HOME/bin" ] && path_contains "$HOME/bin"; then
        install_dir="$HOME/bin"
    elif [ -d "$HOME/go/bin" ] && path_contains "$HOME/go/bin"; then
        install_dir="$HOME/go/bin"
    else
        install_dir="$HOME/.local/bin"
        needs_symlink=1
    fi
}

path_contains() {
    case ":$PATH:" in
        *":$1:"*) return 0 ;;
    esac
    return 1
}

setup_tmp() {
    tmp="$(mktemp -d)"
    trap 'rm -rf "$tmp"' EXIT
}

download() {
    local url="https://github.com/retr0h/$APP/releases/download/v${version}/${APP}_${os}_${arch}"
    printf "${MUTED}downloading mlb-mcp v%s (%s/%s)...${NC}\n" "$version" "$os" "$arch"
    curl -fsSL "$url" -o "$tmp/$APP" || err "download failed: $url"
    chmod +x "$tmp/$APP"
}

verify_checksum() {
    local checksums_url="https://github.com/retr0h/$APP/releases/download/v${version}/checksums.txt"
    if have sha256sum; then
        curl -fsSL "$checksums_url" -o "$tmp/checksums.txt" 2>/dev/null || return 0
        expected="$(grep "${APP}_${os}_${arch}" "$tmp/checksums.txt" | awk '{print $1}')"
        actual="$(sha256sum "$tmp/$APP" | awk '{print $1}')"
        [ "$expected" = "$actual" ] || err "checksum mismatch"
    fi
}

strip_quarantine() {
    [ "$os" = "darwin" ] || return 0
    xattr -d com.apple.quarantine "$tmp/$APP" 2>/dev/null || true
}

install_binary() {
    mkdir -p "$install_dir" || err "cannot create $install_dir"
    install -m 755 "$tmp/$APP" "$install_dir/$APP" \
        || err "cannot write to $install_dir/$APP"
}

maybe_symlink() {
    [ "$needs_symlink" = "1" ] || return 0
    if [ -w /usr/local/bin ]; then
        ln -sf "$install_dir/$APP" /usr/local/bin/$APP 2>/dev/null || true
    fi
}

print_summary() {
    printf "\n"
    printf "${MUTED}█▀▄▀█ █░░ █▄▄   █▀▄▀█ █▀▀ █▀█${NC}   ${MUTED}installed to${NC} ${ACCENT}%s/$APP${NC}\n" "$install_dir"
    printf "${ACCENT}█░▀░█ █▄▄ █▄█   █░▀░█ █▄▄ █▀▀${NC}   ${MUTED}version${NC} ${NC}%s${NC}\n" "$version"
    printf "\n"
    if ! path_contains "$install_dir"; then
        print_message warning "Add this to your shell rc:"
        printf "  ${NC}export PATH=\"%s:\$PATH\"${NC}\n\n" "$install_dir"
    fi
    printf "${MUTED}Start the MCP server:${NC}\n"
    printf "  mlb-mcp mcp start       ${MUTED}# stdio transport${NC}\n"
    printf "\n"
    printf "${MUTED}Docs:${NC} https://github.com/retr0h/mlb-mcp\n"
    printf "\n"
}

main() {
    detect_os
    detect_arch
    resolve_version
    resolve_install_dir
    setup_tmp
    download
    verify_checksum
    strip_quarantine
    install_binary
    maybe_symlink
    print_summary
}

main "$@"
