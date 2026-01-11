#!/bin/bash
set -e

# OpenTrusty One-Click Bootstrap Installer
# Purpose: Remote installer for quick setup via curl | sh

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 1. Pre-flight checks
if [ "$EUID" -ne 0 ]; then
  log_error "This script must be run as root (or via sudo)."
  exit 1
fi

# 2. Environment & OS Detection
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$OS" != "linux" ]; then
  log_error "OpenTrusty currently only supports Linux (for systemd)."
  exit 1
fi

case $ARCH in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# 3. Versioning
REPO_BASE="https://github.com/opentrusty"
GLOBAL_VERSION=${VERSION:-""}

# 4. Component Selection
# Use INSTALL_COMPONENTS="admin auth control-panel" to customize
COMPONENTS=${INSTALL_COMPONENTS:-"admin auth control-panel"}

TMP_DIR="/tmp/opentrusty-bootstrap"
mkdir -p "$TMP_DIR"
cd "$TMP_DIR"

install_component() {
  local comp=$1
  local repo="opentrusty-$comp"
  local comp_version="$GLOBAL_VERSION"
  
  log_info "--- Preparing $comp ---"

  # Detection version for this component if no global version is set
  if [ -z "$comp_version" ]; then
    log_info "Fetching latest version for $repo from GitHub..."
    comp_version=$(curl -s "https://api.github.com/repos/opentrusty/$repo/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$comp_version" ] || [[ "$comp_version" == *"api.github.com"* ]]; then
      log_warn "Failed to fetch version for $comp via API, falling back to v0.1.0"
      comp_version="v0.1.0"
    fi
  fi
  
  log_info "Target Version for $comp: ${comp_version}"

  local tarball=""
  if [ "$comp" == "control-panel" ]; then
    tarball="opentrusty-control-panel-$comp_version.tar.gz"
  else
    tarball="opentrusty-$comp-$comp_version-linux-$ARCH.tar.gz"
  fi
  
  URL="${REPO_BASE}/${repo}/releases/download/${comp_version}/${tarball}"
  
  log_info "Downloading $tarball..."
  if ! curl -L -O "$URL"; then
    log_error "Failed to download $comp. Skipping."
    return 1
  fi
  
  log_info "Extracting $tarball..."
  mkdir -p "$comp-extract"
  tar -xzf "$tarball" -C "$comp-extract" --strip-components=1
  
  log_info "Running installer for $comp..."
  (cd "$comp-extract" && sudo ./install.sh)
  
  log_success "$comp installation attempt completed."
}

# 5. Execution Loop
for comp in $COMPONENTS; do
  install_component "$comp"
done

# 6. Cleanup
log_info "Cleaning up..."
rm -rf "$TMP_DIR"

log_success "OpenTrusty Bootstrap Complete!"
log_info "Please follow the 'Next Steps' provided by each component's installer above."
