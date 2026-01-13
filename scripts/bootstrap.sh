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

# 4. Global Commands (Uninstall)
if [ "$1" == "uninstall" ]; then
  log_info "=== OpenTrusty Global Uninstaller ==="
  echo "This will uninstall selected components from this host."
  
  shift # Remove 'uninstall' from args
  UNINSTALL_COMPONENTS="$@"
  if [ -z "$UNINSTALL_COMPONENTS" ]; then
    if [ -t 0 ]; then
      echo "Which components would you like to uninstall? (Separate by space, or leave empty for ALL)"
      echo "Options: cli, admin, auth, control-panel"
      read -p "Selection [cli admin auth control-panel]: " SELECTED
      UNINSTALL_COMPONENTS=${SELECTED:-"cli admin auth control-panel"}
    else
      UNINSTALL_COMPONENTS="cli admin auth control-panel"
    fi
  fi
  
  for comp in $UNINSTALL_COMPONENTS; do
    repo="opentrusty-$comp"
    comp_version=$(curl -s "https://api.github.com/repos/opentrusty/$repo/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$comp_version" ]; then comp_version="v0.1.0"; fi
    
    tarball=""
    if [ "$comp" == "control-panel" ]; then tarball="opentrusty-control-panel-$comp_version.tar.gz"
    elif [ "$comp" == "cli" ]; then tarball="opentrusty-cli-$comp_version-linux-$ARCH.tar.gz"
    else tarball="opentrusty-$comp-$comp_version-linux-$ARCH.tar.gz"; fi
    
    URL="${REPO_BASE}/${repo}/releases/download/${comp_version}/${tarball}"
    curl -sL -f -O "$URL"
    mkdir -p "$comp-uninstall"
    tar -xzf "$tarball" -C "$comp-uninstall" --strip-components=1
    (cd "$comp-uninstall" && bash ./uninstall.sh)
    rm -rf "$comp-uninstall" "$tarball"
  done
  
  log_success "Uninstallation complete."
  exit 0
fi

# 5. Component Selection Logic
# Priority: 1. CLI Arguments, 2. INSTALL_COMPONENTS Env, 3. Interactive Prompt
COMPONENTS=""

if [ $# -gt 0 ]; then
  COMPONENTS="$@"
elif [ -n "$INSTALL_COMPONENTS" ]; then
  COMPONENTS="$INSTALL_COMPONENTS"
else
  if [ -t 0 ]; then
    log_info "No components specified. Entering interactive selection..."
    echo "Which components would you like to install? (Separate by space, or leave empty for ALL)"
    echo "Options: cli, admin, auth, control-panel"
    read -p "Selection [cli admin auth control-panel]: " SELECTED
    COMPONENTS=${SELECTED:-"cli admin auth control-panel"}
  else
    log_info "Non-interactive mode, installing all components."
    COMPONENTS="cli admin auth control-panel"
  fi
fi

log_info "Installing components: $COMPONENTS"

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
  elif [ "$comp" == "cli" ]; then
    # CLI binary is named 'opentrusty'
    tarball="opentrusty-cli-$comp_version-linux-$ARCH.tar.gz"
  else
    tarball="opentrusty-$comp-$comp_version-linux-$ARCH.tar.gz"
  fi
  
  URL="${REPO_BASE}/${repo}/releases/download/${comp_version}/${tarball}"
  
  log_info "Downloading $tarball..."
  if ! curl -sL -f -O "$URL"; then
    log_error "Failed to download $comp ($URL). Skipping."
    return 1
  fi
  
  log_info "Extracting $tarball..."
  local extract_dir="$comp-extract"
  mkdir -p "$extract_dir"
  tar -xzf "$tarball" -C "$extract_dir" --strip-components=1
  
  log_info "Running installer for $comp..."
  (cd "$extract_dir" && bash ./install.sh)
  
  log_success "$comp installation completed."

  # CLI Post-install initialization logic
  if [[ "$comp" == "cli" && -t 0 ]]; then
    run_cli_bootstrapper
  fi
}

run_cli_bootstrapper() {
  echo ""
  log_info "=== OpenTrusty CLI Interactive Setup ==="
  log_info "Note: This will perform initialization (migration & bootstrap)."
  
  # Discrete DB Collect
  read -p "Enter Database Host [localhost]: " OT_DB_HOST
  OT_DB_HOST=${OT_DB_HOST:-"localhost"}
  read -p "Enter Database Port [5432]: " OT_DB_PORT
  OT_DB_PORT=${OT_DB_PORT:-"5432"}
  read -p "Enter Database User [postgres]: " OT_DB_USER
  OT_DB_USER=${OT_DB_USER:-"postgres"}
  read -s -p "Enter Database Password [password]: " OT_DB_PASS
  echo ""
  OT_DB_PASS=${OT_DB_PASS:-"password"}
  read -p "Enter Database Name [opentrusty]: " OT_DB_NAME
  OT_DB_NAME=${OT_DB_NAME:-"opentrusty"}
  
  export OPENTRUSTY_DB_HOST="$OT_DB_HOST"
  export OPENTRUSTY_DB_PORT="$OT_DB_PORT"
  export OPENTRUSTY_DB_USER="$OT_DB_USER"
  export OPENTRUSTY_DB_PASSWORD="$OT_DB_PASS"
  export OPENTRUSTY_DB_NAME="$OT_DB_NAME"
  export OPENTRUSTY_DB_SSLMODE="disable"

  log_info "Running database migrations..."
  if ! opentrusty migrate; then
    log_error "Migration failed. Please check your DB credentials."
    return 1
  fi
  log_success "Migrations completed."

  echo ""
  read -p "Do you want to bootstrap the platform admin now? (y/N): " RUN_BOOTSTRAP
  if [[ "$RUN_BOOTSTRAP" =~ ^[Yy]$ ]]; then
    read -p "Enter OPENTRUSTY_IDENTITY_SECRET (32-byte hex): " IDENT_SECRET
    read -p "Enter Platform Admin Email: " ADMIN_EMAIL
    read -s -p "Enter Platform Admin Password: " ADMIN_PASSWORD
    echo ""

    if [ -n "$IDENT_SECRET" ] && [ -n "$ADMIN_EMAIL" ] && [ -n "$ADMIN_PASSWORD" ]; then
      export OPENTRUSTY_IDENTITY_SECRET="$IDENT_SECRET"
      export OPENTRUSTY_BOOTSTRAP_ADMIN_EMAIL="$ADMIN_EMAIL"
      export OPENTRUSTY_BOOTSTRAP_ADMIN_PASSWORD="$ADMIN_PASSWORD"
      
      if opentrusty bootstrap; then
        log_success "Platform admin bootstrapped."
      else
        log_error "Bootstrap failed."
      fi
    else
      log_warn "Missing required fields, skipping bootstrap."
    fi
  fi

  echo ""
  read -p "Do you want to persist these settings to /etc/opentrusty/cli.env? (y/N): " PERSIST
  if [[ "$PERSIST" =~ ^[Yy]$ ]]; then
    cat > /etc/opentrusty/cli.env << EOF
# OpenTrusty CLI Configuration
OPENTRUSTY_DB_HOST=$OPENTRUSTY_DB_HOST
OPENTRUSTY_DB_PORT=$OPENTRUSTY_DB_PORT
OPENTRUSTY_DB_USER=$OPENTRUSTY_DB_USER
OPENTRUSTY_DB_PASSWORD=$OPENTRUSTY_DB_PASSWORD
OPENTRUSTY_DB_NAME=$OPENTRUSTY_DB_NAME
OPENTRUSTY_DB_SSLMODE=$OPENTRUSTY_DB_SSLMODE
OPENTRUSTY_IDENTITY_SECRET=$OPENTRUSTY_IDENTITY_SECRET
EOF
    chmod 600 /etc/opentrusty/cli.env
    log_success "Persisted CLI configuration to /etc/opentrusty/cli.env"
  fi
}

# 5. Execution Loop
# Ensure CLI is installed first if selected
if [[ "$COMPONENTS" == *"cli"* ]]; then
  install_component "cli"
  # Filter out cli from the rest to avoid double install
  COMPONENTS=$(echo "$COMPONENTS" | sed 's/\bcli\b//g')
fi

for comp in $COMPONENTS; do
  if [ -n "$comp" ]; then
    install_component "$comp"
  fi
done

# 6. Cleanup
log_info "Cleaning up..."
rm -rf "$TMP_DIR"

log_success "OpenTrusty Bootstrap Complete!"
log_info "Please follow the 'Next Steps' provided by each component's installer above."
