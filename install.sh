#!/usr/bin/env bash

set -e

setup_shell_and_path() {
    if [[ ":$PATH:" == *":${BIN_DIR}:"* ]]; then
        return
    fi
    case $SHELL in
        */zsh) PROFILE=$HOME/.zshrc ;;
        */bash) PROFILE=$HOME/.bashrc ;;
        */fish) PROFILE=$HOME/.config/fish/config.fish ;;
        */ash) PROFILE=$HOME/.profile ;;
        *)
            echo "Add ${BIN_DIR} to your PATH to run $INSTALL_NAME."
            return
    esac
    echo >> "$PROFILE" && echo "export PATH=\"\$PATH:$BIN_DIR\"" >> "$PROFILE"
    echo "Added ${BIN_DIR} to PATH in $PROFILE (open a new terminal to use it)."
}

detect_platform_and_arch() {
    PLATFORM="$(uname | tr '[:upper:]' '[:lower:]')"
    if [[ "$PLATFORM" == mingw*_nt* ]]; then
        PLATFORM="windows"
    fi

    ARCHITECTURE="$(uname -m)"
    if [ "${ARCHITECTURE}" = "x86_64" ]; then
        # Redirect stderr to /dev/null to avoid printing errors if non Rosetta.
        if [ "$(sysctl -n sysctl.proc_translated 2>/dev/null)" = "1" ]; then
            ARCHITECTURE="arm64" # Rosetta.
        else
            ARCHITECTURE="amd64" # Intel.
        fi
    elif [ "${ARCHITECTURE}" = "arm64" ] || [ "${ARCHITECTURE}" = "aarch64" ]; then
        ARCHITECTURE="arm64" # Arm.
    else
        ARCHITECTURE="amd64" # Amd.
    fi

    if [[ "$PLATFORM" == "windows" ]]; then
        ARCHIVE_EXT=".zip"
        EXTENSION=".exe"
    else
        ARCHIVE_EXT=".tar.gz"
        EXTENSION=""
    fi
}

get_latest_version() {
    # Errors go to stderr: stdout is captured as the version.
    local tag
    tag=$(curl -fsSL "https://api.github.com/repos/sudarshan032/claude-booths/releases/latest" 2>/dev/null |
        grep -m1 '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
    if [ -z "$tag" ]; then
        echo "Could not find a published release at https://github.com/sudarshan032/claude-booths/releases" >&2
        exit 1
    fi
    echo "$tag"
}

verify_checksum() {
    local tmp_dir=$1
    local archive_name=$2
    if ! curl -fsSL "${RELEASE_URL}/checksums.txt" -o "${tmp_dir}/checksums.txt"; then
        echo "Error: could not download checksums.txt for v${VERSION}" >&2
        rm -rf "$tmp_dir"
        exit 1
    fi
    local expected actual
    expected=$(grep " ${archive_name}\$" "${tmp_dir}/checksums.txt" | cut -d' ' -f1)
    if command -v sha256sum &> /dev/null; then
        actual=$(sha256sum "${tmp_dir}/${archive_name}" | cut -d' ' -f1)
    else
        actual=$(shasum -a 256 "${tmp_dir}/${archive_name}" | cut -d' ' -f1)
    fi
    if [ -z "$expected" ] || [ "$expected" != "$actual" ]; then
        echo "Error: checksum mismatch for ${archive_name}; not installing." >&2
        rm -rf "$tmp_dir"
        exit 1
    fi
    echo "Checksum OK"
}

download_release() {
    local version=$1
    local binary_url=$2
    local archive_name=$3
    local tmp_dir=$4

    echo "Downloading binary from $binary_url"
    DOWNLOAD_OUTPUT=$(curl -sS -L -f -w '%{http_code}' "$binary_url" -o "${tmp_dir}/${archive_name}" 2>&1)
    HTTP_CODE=$?
    
    if [ $HTTP_CODE -ne 0 ]; then
        echo "Error: Failed to download release asset"
        echo "This could be because:"
        echo "1. The release ${version} doesn't have assets uploaded yet"
        echo "2. The asset for ${PLATFORM}_${ARCHITECTURE} wasn't built"
        echo "3. The asset name format has changed"
        echo ""
        echo "Expected asset name: ${archive_name}"
        echo "URL attempted: ${binary_url}"
        rm -rf "$tmp_dir"
        exit 1
    fi
}

extract_and_install() {
    local tmp_dir=$1
    local archive_name=$2
    local bin_dir=$3
    local extension=$4

    if [[ "$PLATFORM" == "windows" ]]; then
        if ! unzip -t "${tmp_dir}/${archive_name}" > /dev/null 2>&1; then
            echo "Error: Downloaded file is not a valid zip archive"
            rm -rf "$tmp_dir"
            exit 1
        fi
        ensure unzip "${tmp_dir}/${archive_name}" -d "$tmp_dir"
    else
        if ! tar tzf "${tmp_dir}/${archive_name}" > /dev/null 2>&1; then
            echo "Error: Downloaded file is not a valid tar.gz archive"
            rm -rf "$tmp_dir"
            exit 1
        fi
        ensure tar xzf "${tmp_dir}/${archive_name}" -C "$tmp_dir"
    fi

    if [ ! -d "$bin_dir" ]; then
        mkdir -p "$bin_dir"
    fi

    # Remove existing binary if upgrading
    if [ "$UPGRADE_MODE" = true ] && [ -f "$bin_dir/$INSTALL_NAME${extension}" ]; then
        echo "Removing previous installation from $bin_dir/$INSTALL_NAME${extension}"
        rm -f "$bin_dir/$INSTALL_NAME${extension}"
    fi

    # Install binary with desired name
    mv "${tmp_dir}/claude-booths${extension}" "$bin_dir/$INSTALL_NAME${extension}"
    rm -rf "$tmp_dir"

    if [ ! -f "$bin_dir/$INSTALL_NAME${extension}" ]; then
        echo "Installation failed, could not find $bin_dir/$INSTALL_NAME${extension}"
        exit 1
    fi

    chmod +x "$bin_dir/$INSTALL_NAME${extension}"
    
    echo ""
    if [ "$UPGRADE_MODE" = true ]; then
        echo "Successfully upgraded '$INSTALL_NAME' to:"
    else
        echo "Installed as '$INSTALL_NAME':"
    fi
    echo "$("$bin_dir/$INSTALL_NAME${extension}" version)"
}

check_command_exists() {
    if command -v "$INSTALL_NAME" &> /dev/null; then
        EXISTING_PATH=$(which "$INSTALL_NAME")
        echo "Found existing installation of '$INSTALL_NAME' at $EXISTING_PATH"
        echo "Will upgrade to the latest version"
        UPGRADE_MODE=true
    else
        UPGRADE_MODE=false
    fi
}

check_and_install_dependencies() {
    echo "Checking for required dependencies..."
    
    # Check for tmux
    if ! command -v tmux &> /dev/null; then
        echo "tmux is not installed. Installing tmux..."
        
        if [[ "$PLATFORM" == "darwin" ]]; then
            # macOS
            if command -v brew &> /dev/null; then
                ensure brew install tmux
            else
                echo "Homebrew is not installed. Please install Homebrew first to install tmux."
                echo "Visit https://brew.sh for installation instructions."
                exit 1
            fi
        elif [[ "$PLATFORM" == "linux" ]]; then
            # Linux
            if command -v apt-get &> /dev/null; then
                ensure sudo apt-get update
                ensure sudo apt-get install -y tmux
            elif command -v dnf &> /dev/null; then
                ensure sudo dnf install -y tmux
            elif command -v yum &> /dev/null; then
                ensure sudo yum install -y tmux
            elif command -v pacman &> /dev/null; then
                ensure sudo pacman -S --noconfirm tmux
            else
                echo "Could not determine package manager. Please install tmux manually."
                exit 1
            fi
        elif [[ "$PLATFORM" == "windows" ]]; then
            echo "For Windows, please install tmux via WSL or another method."
            exit 1
        fi
        
        echo "tmux installed successfully."
    else
        echo "tmux is already installed."
    fi
    
    # GitHub CLI is optional: only the 'p' (push branch) key uses it.
    if ! command -v gh &> /dev/null; then
        echo "Optional: install the GitHub CLI (gh) to push booth branches with 'p': https://cli.github.com"
    fi

    echo "All dependencies are installed."
}

main() {
    # Parse command line arguments
    INSTALL_NAME="claude-booths"
    UPGRADE_MODE=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --name)
                INSTALL_NAME="$2"
                shift 2
                ;;
            *)
                echo "Unknown option: $1"
                echo "Usage: install.sh [--name <n>]"
                exit 1
                ;;
        esac
    done

    check_command_exists
    detect_platform_and_arch
    
    check_and_install_dependencies

    BIN_DIR=${BIN_DIR:-$HOME/.local/bin}
    VERSION=${VERSION:-"latest"}
    if [[ "$VERSION" == "latest" ]]; then
        VERSION=$(get_latest_version) || exit 1
    fi

    RELEASE_URL="https://github.com/sudarshan032/claude-booths/releases/download/v${VERSION}"
    ARCHIVE_NAME="claude-booths_${VERSION}_${PLATFORM}_${ARCHITECTURE}${ARCHIVE_EXT}"
    BINARY_URL="${RELEASE_URL}/${ARCHIVE_NAME}"
    TMP_DIR=$(mktemp -d)
    
    download_release "$VERSION" "$BINARY_URL" "$ARCHIVE_NAME" "$TMP_DIR"
    verify_checksum "$TMP_DIR" "$ARCHIVE_NAME"
    extract_and_install "$TMP_DIR" "$ARCHIVE_NAME" "$BIN_DIR" "$EXTENSION"
    setup_shell_and_path
}

# Run a command that should never fail. If the command fails execution
# will immediately terminate with an error showing the failing
# command.
ensure() {
    if ! "$@"; then err "command failed: $*"; fi
}

err() {
    echo "$1" >&2
    exit 1
}

main "$@" || exit 1
