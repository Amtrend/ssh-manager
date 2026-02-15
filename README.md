# SSH Connection Manager

A modern, lightweight, and secure web-based SSH connection manager. Manage your servers directly from your browser using a built-in terminal.

## Key Features
* **Web-Terminal:** Fully functional console (xterm.js) right in your browser.
* **Persistent Sessions:** Connections remain active for a set duration even if you close the tab or lose your internet connection.
* **High Security:** Private SSH keys are encrypted using **AES-256** before being stored in the database.
* **Database Versatility:** Supports **PostgreSQL** for large-scale deployments and **SQLite** for a quick start.
* **Zero Config:** Automatically initializes tables and creates an admin account on the first run.
* **Cleanup Service:** Automatically closes abandoned SSH sessions based on a configurable timeout.

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
  -e DB_TYPE=sqlite \
  -e ENCRYPTION_KEY=[YOUR_KEY] \
  -e SESSION_SECRET=[YOUR_SECRET] \
  -v $(pwd)/data:/root/data \
  --name ssh-manager \
  alexmaltz/ssh-manager:latest

```

---

## Environment Variables

| Variable | Description | Default |
| --- | --- | --- |
| `PORT` | Web interface port | `8080` |
| `DB_TYPE` | Database type (`postgres` or `sqlite`) | `sqlite` |
| `DB_NAME` | DB name (or filename for sqlite) | `ssh_manager` |
| `ENCRYPTION_KEY` | **(Required)** 32-byte Hex key for AES-256 | - |
| `SESSION_SECRET` | **(Required)** Secret key for signing session cookies | - |
| `SESSION_TIMEOUT` | Max life for abandoned sessions (e.g., 10m, 1h) | `10m` |
| `CLEANUP_INTERVAL` | Cleanup frequency for dead sessions (e.g., 2m) | `2m` |
| `INITIAL_ADMIN_USER` | Admin username on first startup | `admin` |
| `INITIAL_ADMIN_PASSWORD` | Admin password on first startup | `admin` |

### How to Generate Keys?

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

## Resources

* **Docker Hub:** [https://hub.docker.com/r/alexmaltz/ssh-manager](https://hub.docker.com/r/alexmaltz/ssh-manager)
* **Standalone Binaries:** Find pre-compiled versions for Windows and Linux in the [Releases](https://github.com/Amtrend/ssh-manager/releases) section.

---

## Security

* **Encryption:** Private SSH keys are encrypted before hitting the DB. Without your `ENCRYPTION_KEY`, the data is useless.
* **Access Protection:** All user passwords are hashed using `bcrypt`.
* **Resource Protection:** The system automatically terminates abandoned SSH tunnels after `SESSION_TIMEOUT` to save resources.
