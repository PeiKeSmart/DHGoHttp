# Changelog

All notable changes to this project will be documented in this file.

The format is inspired by Keep a Changelog, and this project adheres (lightly) to Semantic Versioning.

## [0.2.0] - 2025-09-13

### Added (Initial Release)

- `-port` flag (takes precedence over PORT env)
- Auto-increment free port scanning (default max 50 attempts) with `-max-port-scan` flag
- `-dir` flag for custom root directory
- Version banner output including selected root and port

### Changed

- README (CN/EN) updated to reflect implemented flags & behavior

### Notes (Initial Release)

- Auto-increment only scans sequentially upward; no wrap-around.

## [0.1.0] - 2025-09-13

### Added

- Initial minimal HTTP static file server.
- Windows elevation (UAC) attempt & firewall rule auto-creation / idempotent check.
- `-no-firewall` flag to skip elevation & firewall logic.
- Rule existence detection before creation.
- English & Chinese README documentation.
- MIT License.
- Planned flags placeholders: `-port`, `-dir` (not yet implemented).

### Notes

- Security is intentionally minimal; do not expose directly to untrusted networks without extra protection.

## [Unreleased]

- Graceful shutdown removing firewall rule
- Access logging / metrics
- Optional token-based access control
- Directory allow/deny filtering
- Bind address flag
- Token auth flag
- Readonly flag
- Multi-platform release workflow

## [0.3.0] - 2025-09-13

### Added (Security & Ops)

- `-bind` flag to specify listen address (default all interfaces)
- `-token` shared token auth (header X-Token or query ?token=)
- `-readonly` flag blocking directory listing
- Access log middleware (IP / method / path / status / bytes / duration)
- Graceful shutdown on Ctrl+C / SIGTERM (with firewall rule removal if created by process)

### Changed (Internals)

- Firewall rule function returns rule name & creation status
- Version bumped to 0.3.0-dev (pre-release state)
 
### Fixed

- Preserve working directory after Windows UAC elevation: if user didn't supply `-dir`, the original absolute startup directory is injected so elevated process serves the intended root instead of system32.
