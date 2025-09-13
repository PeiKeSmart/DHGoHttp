# DHGoHttp (English)

[简体中文 README](./README.md)

A minimal cross-platform directory HTTP file server. Drop it into a folder and instantly download files over HTTP. On Windows it can auto-elevate (UAC) and add a firewall inbound allow rule to make the port reachable from other machines. Now supports custom root, explicit port flag with auto-increment fallback, and adjustable scan range.

## Features

- Serve the current working directory as static files
- Zero configuration: `go run .` or run the compiled binary
- Windows specific:
  - Detects non-admin and attempts UAC elevation
  - Adds (idempotently) a Windows Defender Firewall inbound rule: `DHGoHttp-<port>`
  - Skips creation if rule already exists
- Override listen port via `-port` flag (preferred) or `PORT` environment variable (default start: 8080)
- Auto-increment to the next free port if the chosen one is occupied (up to configurable attempts)
- Change max scan attempts with `-max-port-scan` (default 50)
- Custom root directory via `-dir` (defaults to current working directory)
- `-no-firewall` flag to skip elevation + firewall logic (ideal for local only)

## Quick Start

### Clone or Download

```bash
# Clone
git clone <your-repo-url> dhgohttp
cd dhgohttp

# Or just copy main.go into any folder
```

### Run (Dev)

```bash
go run .
```

### Build & Run (recommended for Windows elevation test)

```bash
go build -o dhgohttp.exe .
./dhgohttp.exe
```

On first non-admin run (Windows) a UAC prompt appears; if accepted, an elevated process creates the firewall rule and starts the server; the original process exits.

## Accessing Files

Example directory structure:

```text
E:/Project/DHGoHttp/
  ├── main.go
  ├── README.md
  └── example-download.sh
```

Access:

```text
http://<host-ip>:8080/example-download.sh
```

Local:

```text
http://localhost:8080/README.md
```

Download via curl:

```bash
curl -O http://localhost:8080/example-download.sh
```

## Command-line Flags

| Flag | Description | Example |
|------|-------------|---------|
| `-no-firewall` | Skip elevation & firewall rule creation | `./dhgohttp.exe -no-firewall` |
| `-elevated` | Internal marker to prevent recursive elevation | (internal use) |
| `-port` | Starting port (auto-increments if busy) | `./dhgohttp.exe -port 9000` |
| `-dir` | Serve this directory instead of CWD | `./dhgohttp.exe -dir C:/Files` |
| `-max-port-scan` | Max increment attempts to find free port | `./dhgohttp.exe -port 9000 -max-port-scan 30` |
| `-bind` | Bind address (default all interfaces) | `./dhgohttp.exe -bind 127.0.0.1` |
| `-token` | Require shared token (X-Token header or ?token=) | `./dhgohttp.exe -token SECRET` |
| `-readonly` | Disable directory listing | `./dhgohttp.exe -readonly` |

## Environment Variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `PORT` | Starting listen port (used only if `-port` not given) | `PORT=9000 go run .` |

> Note: Only specify the number. Code adds the leading `:` automatically.

## Windows Firewall & Elevation

1. Non-admin launch triggers `ShellExecuteW` + `runas` for UAC.
2. If user cancels: continues normally (no firewall rule added) → only localhost / already allowed scopes.
3. Rule name format: `DHGoHttp-<port>`, e.g. `DHGoHttp-8080`.
4. Check rule existence:

```powershell
netsh advfirewall firewall show rule name="DHGoHttp-8080"
```

1. Add manually (admin PowerShell):

```powershell
netsh advfirewall firewall add rule name="DHGoHttp-8080" dir=in action=allow protocol=TCP localport=8080
```

1. Delete manually:

```powershell
netsh advfirewall firewall delete rule name="DHGoHttp-8080"
```

## Typical Use Cases

- Temporarily share scripts, installers, logs
- Quick intra-network file transfer
- Expose build artifacts inside CI/CD or containers (not recommended as-is for production)

## FAQ

### 1. Other machines cannot access the server

- No elevation → rule not created
- Port blocked by security software
- Default port 8080 already used
- Network / firewall policy restrictions

### 2. Custom port & auto-increment?

Option 1 (flag):

```bash
./dhgohttp.exe -port 9000
```

If 9000 is taken it will try 9001, 9002, ... until success or attempts exceed `-max-port-scan`.

Option 2 (environment variable):

```bash
PORT=9000 go run .
```

PowerShell:

```powershell
$env:PORT=9000; go run .
```

### 3. Avoid UAC every time?

- Run terminal as Administrator first
- Or pre-create firewall rule & use `-no-firewall`

### 4. Why "rule already exists"?

We check before creating—safe & idempotent.

### 5. Custom root / auto-increment / readonly / token?

Implemented: `-dir` (custom root), auto-increment port, `-readonly` (blocks directory listing), `-token` (simple shared secret). Logs show attempted port sequence when fallback happens.

### 6. Graceful shutdown?

On Ctrl+C (SIGINT) or SIGTERM:

1. Stop accepting new connections
2. Allow up to 5s for in-flight requests
3. If a firewall rule was created by this process it is removed
4. Exit with a final log line

## Code Layout

- `main.go` core logic (serve + elevation + firewall)
- `example-download.sh` sample file
- `README.md` Chinese documentation
- `README.en.md` English documentation

## Security Notes

- No auth—any reachable client can download all files.
- Don’t expose directly to the public internet without adding reverse proxy / auth / filtering.
- Elevation only on Windows; Linux/macOS require manual firewall configuration (e.g. ufw / security group).

## Future Enhancements (Ideas)

- Directory allow/deny filtering
- Extended metrics (export Prometheus)
- Optional per-file checksum endpoint

## License

MIT License. See [LICENSE](./LICENSE).

---
Feel free to open issues or request new features. 😊
