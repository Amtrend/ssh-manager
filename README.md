# SSH Connection Manager

A modern, lightweight, and secure web-based SSH connection manager. Manage your servers directly from your browser with a built-in terminal and a full-featured file manager.

## Key Features
* **Web-Terminal:** Fully functional console (xterm.js) right in your browser.
* **Integrated SFTP Manager:** * Fast directory navigation with breadcrumbs.
    * Single file downloads and **Multi-file ZIP downloads** on the fly.
    * Drag-and-drop file uploads.
    * Recursive folder compression for downloads.
* **Flexible Authentication:** Connect to hosts using either **Private Keys** or **Passwords**.
* **Persistent Sessions:** Connections remain active for a set duration even if you close the tab. SFTP and Terminal share the same secure tunnel.
* **High Security:** Both SSH Private Keys and Host Passwords are encrypted using **AES-256 GCM** before being stored in the database.
* **Zero Config:** Automatically initializes tables and creates an admin account on the first run.
* **Smart Cleanup:** Automatically closes abandoned SSH sessions based on a configurable timeout.
* **PWA Support:** Install the application on your Desktop (Windows/Linux/macOS) or Mobile (Android/iOS) as a standalone app with its own icon and splash screen.
* **Real-time and Smart Badge Push Notifications:** Receive instant alerts about:
    * Automatic cleanup of abandoned sessions.
    * The app icon shows the exact number of unread session alerts. The counter automatically resets when you open the application.

---

## Deployment Methods

### 1. Docker Compose (Recommended)

#### Option A: SQLite (Quick Start)
```yaml
services:
  ssh-manager:
    image: alexmaltz/ssh-manager:latest
    container_name: ssh-manager
    ports:
      - "8080:8080"
    volumes:
      - ./data:/root/data
    environment:
      - APP_ENV=prod
      - DB_TYPE=sqlite
      - SESSION_SECRET=[CREATE_A_SECRET]
      - ENCRYPTION_KEY=[GENERATE_32_BYTE_HEX]
      - INITIAL_ADMIN_USER=admin
      - INITIAL_ADMIN_PASSWORD=admin
    restart: unless-stopped

```

#### Option B: PostgreSQL

```yaml
services:
  app:
    image: alexmaltz/ssh-manager:latest
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=prod
      - DB_TYPE=postgres
      - DB_HOST=db
      - DB_PORT=5432
      - DB_NAME=ssh_db
      - DB_USER=postgres
      - DB_PASSWORD=postgres_pass
      - SESSION_SECRET=[CREATE_A_SECRET]
      - ENCRYPTION_KEY=[GENERATE_32_BYTE_HEX]
    depends_on:
      - db

  db:
    image: postgres:16
    environment:
      POSTGRES_DB: ssh_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres_pass
    volumes:
      - ./pg_data:/var/lib/postgresql/data

```

### 2. Docker Run (Standalone Container)

```bash
docker run -d \
  -p 8080:8080 \
  -e APP_ENV=prod \
  -e DB_TYPE=sqlite \
  -e ENCRYPTION_KEY=[YOUR_KEY] \
  -e SESSION_SECRET=[YOUR_SECRET] \
  -v $(pwd)/data:/root/data \
  --name ssh-manager \
  alexmaltz/ssh-manager:latest

```

---

## Progressive Web App (PWA) & Notifications

The service is fully compatible with **PWA standards**. You can "install" it via your browser (Chrome/Edge/Safari) to use it as a native-like application.

### Push Notifications Setup
To enable Push Notifications, you need to provide **VAPID keys**. You can generate them using the following command (or any online tool):

```bash
# Example using npx (web-push)
npx web-push generate-vapid-keys
```

Once generated, add `VAPID_PUBLIC_KEY` and `VAPID_PRIVATE_KEY` to your environment variables. 

> **Tip:** You can also configure `PUSH_TTL` (Time To Live). This defines how long (in seconds) the push server (Google/Apple) will keep the message if the user's device is offline. Default is `3600` (1 hour).

---

## Environment Variables

| Variable | Description | Default |
| --- | --- | --- |
| `APP_ENV` | Application environment (`prod` or `debug`). If prod, cookies are sent only over HTTPS | `debug` |
| `PORT` | Web interface port | `8080` |
| `DB_TYPE` | Database type (`postgres` or `sqlite`) | `sqlite` |
| `DB_NAME` | DB name (or filename for sqlite) | `ssh_manager` |
| `ENCRYPTION_KEY` | **(Required)** 32-byte Hex key for AES-256 | - |
| `SESSION_SECRET` | **(Required)** Secret key for signing session cookies | - |
| `SESSION_TIMEOUT` | Max life for abandoned sessions (e.g., 10m, 1h) | `10m` |
| `CLEANUP_INTERVAL` | Cleanup frequency for dead sessions (e.g., 2m) | `2m` |
| `INITIAL_ADMIN_USER` | Admin username on first startup | `admin` |
| `INITIAL_ADMIN_PASSWORD` | Admin password on first startup | `admin` |
| `VAPID_PUBLIC_KEY` | Public key for Push Notifications (Base64) | - |
| `VAPID_PRIVATE_KEY` | Private key for Push Notifications (Base64) | - |
| `VAPID_EMAIL` | Contact email for Push Notifications (e.g., mailto:admin@example.com) | - |
| `PUSH_TTL` | Time-to-live for notifications in seconds | `3600` |

## Security & Production Notes

### Production vs Debug Mode
The application uses the `APP_ENV` variable to handle session security:

* **Debug Mode (`debug`):** Suitable for local testing. Cookies are sent over plain HTTP.
* **Production Mode (`prod`):** **Recommended for public servers.** Enables the `Secure` flag on cookies. 
    * **Note:** In `prod` mode, the app **requires HTTPS** (except on `localhost`). Accessing via IP or non-secure domain will result in login failure.

### Key Generation

To run the application, you need to generate two random keys using your terminal:

* **ENCRYPTION_KEY (for AES-256):**
```bash
openssl rand -hex 32

```


* **SESSION_SECRET:**
```bash
openssl rand -base64 32

```


> **Warning:** Losing your `ENCRYPTION_KEY` will make it impossible to decrypt existing SSH keys stored in the database.

---

## Configuration & Usage

### Host Management
When adding a new host, you can specify:
* **Auth Type:** Choose between Password or Private Key.
* **Default Path:** Set a starting directory for the SFTP manager (e.g., `/var/www/html` or `/home/user/logs`).
* **Encryption:** The system automatically encrypts your credentials using your `ENCRYPTION_KEY`.

### SFTP Capabilities
The built-in file manager allows you to:
1. **Navigate:** Click through directories with instant breadcrumb updates.
2. **Download ZIP:** Select multiple files or folders; the server will stream them to you as a single ZIP archive without creating temporary files on the remote host.
3. **Upload:** Upload files via the web interface directly to the current remote directory.

---

## Resources

* **Docker Hub:** [https://hub.docker.com/r/alexmaltz/ssh-manager](https://hub.docker.com/r/alexmaltz/ssh-manager)
* **Standalone Binaries:** Find pre-compiled versions for Windows and Linux in the [Releases](https://github.com/Amtrend/ssh-manager/releases) section.

---

## Security

* **Credential Encryption:** All sensitive data (SSH keys and passwords) is encrypted at rest. Without your `ENCRYPTION_KEY`, the data is useless.
* **CSRF Protection:** Secure tokens are required for all file operations (Upload/Download).
* **Multiplexing:** SFTP operations run over the same encrypted SSH tunnel as your terminal, reducing the attack surface.
* **Access Protection:** All user passwords are hashed using `bcrypt`.
* **Push Privacy:** Notifications are sent using VAPID (Voluntary Application Server Identification), ensuring that only your server can send messages to your browser.
* **Offline Ready:** While the core functionality requires a network, the PWA manifest ensures the app shell loads instantly.
