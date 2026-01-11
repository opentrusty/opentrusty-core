# OpenTrusty Deployment Guide

This document provides a comprehensive guide for deploying the OpenTrusty platform in a production-like environment using systemd.

## 1. Architecture Overview

OpenTrusty is split into several independent planes:
- **Authentication Plane (`auth`)**: Handles OIDC/OAuth2 login.
- **Administration Plane (`admin`)**: Management API.
- **Control Panel (`console`)**: Static frontend UI.
- **CLI (`opentrusty`)**: Migrations and bootstrapping.

> [!TIP]
> **Unix Daemon Naming**: All backend services follow the Unix daemon convention with a `d` suffix (e.g., `opentrusty-admind`, `opentrusty-authd`).

## 2. Infrastructure Requirements

- **Operating System**: Linux (amd64).
- **Database**: PostgreSQL 15+.
- **Reverse Proxy**: Caddy, Nginx, or similar.

### 3.0 One-Click Installation (Recommended)

If you have internet access on the target machine, you can install the full OpenTrusty stack (admin, auth, and control panel) with a single command:

```bash
curl -sSL https://raw.githubusercontent.com/opentrusty/opentrusty-core/main/scripts/bootstrap.sh | sudo bash
```

This script will automatically detect your OS/Arch, fetch the latest release, and execute the individual component installers.

> [!NOTE]
> You can install specific components by setting `INSTALL_COMPONENTS` environment variable:
> `curl ... | sudo INSTALL_COMPONENTS="admin" bash`

### 3.1 Prerequisities (Manual Mode)

1. Create a dedicated system user:
   ```bash
   sudo useradd -r -s /bin/false opentrusty
   ```

2. Create storage directory:
   ```bash
   sudo mkdir -p /var/lib/opentrusty
   sudo chown opentrusty:opentrusty /var/lib/opentrusty
   ```

3. Create configuration directory:
   ```bash
   sudo mkdir -p /etc/opentrusty
   ```

### 3.2 Component Installation

Download the latest release tarballs for each component from GitHub.

For each component (auth, admin, cli):
1. Extract the tarball.
2. Run `sudo ./install.sh`.

### 3.3 Configuration

Configuration is managed via environment files in `/etc/opentrusty/`.

1. **`shared.env`**: Variables shared across `auth`, `admin`, and `cli`.
   - `OPENTRUSTY_DATABASE_URL`
   - `OPENTRUSTY_IDENTITY_SECRET`
   - `OPENTRUSTY_SESSION_SECRET`

2. **Component Envs**:
   - `auth.env`: Auth-specific settings.
   - `admin.env`: Admin-specific settings.

Refer to the `.env.example` file in each component package for a full list of variables.

### 3.4 Bootstrapping

Before starting the services, you must migrate the database and bootstrap the platform admin:

```bash
# Run migrations
export OPENTRUSTY_DATABASE_URL=...
opentrusty migrate

# Bootstrap platform admin
export OPENTRUSTY_IDENTITY_SECRET=...
opentrusty bootstrap
```

### 3.5 Starting Services

```bash
sudo systemctl enable --now opentrusty-authd
sudo systemctl enable --now opentrusty-admind
```

## 4. Reverse Proxy Cleanup

It is recommended to use Caddy or Nginx to handle TLS and aggregate the planes.

- **Admin API**: Proxy `/api/*` to `localhost:8081`.
- **Auth UI/OIDC**: Proxy `/auth/*` and `/.well-known/*` to `localhost:8080`.
- **Control Panel**: Serve static files from `dist/` and route SPA paths to `index.html`.

Refer to `opentrusty-control-panel/Caddyfile.example` for a standard configuration.
