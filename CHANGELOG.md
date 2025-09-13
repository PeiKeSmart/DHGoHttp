# Changelog

All notable changes to this project will be documented in this file.

The format is inspired by Keep a Changelog, and this project adheres (lightly) to Semantic Versioning.

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
- Add `-port` flag
- Add `-dir` flag
- Auto-increment free port selection
- Graceful shutdown removing firewall rule
- Access logging / metrics
- Optional token-based access control
