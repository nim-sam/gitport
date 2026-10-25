# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-10-24

### Added
- Git over SSH server functionality.
- SSH-accessible TUI dashboard for user management, commit browsing, and server logs.
- Server overview tab with server status information.
- Commit activity graph visualization.
- Role and tab-based help restrictions with context-specific key bindings.
- mDNS-based LAN discovery for `.local` hostnames.
- Automatic gitport remote configuration on initialization.
- Initialization wizard (`gitport init`) and server startup (`gitport start`).
- Built-in database for user management.
- Non-invasive bare repository serving.
- Self-contained SSH server with built-in TUI dashboard.
- MIT License.

### Changed
- Improved tab navigation and focus management.
- Enhanced styling with truecolor color profile support.
- Updated dependencies.
- Improved role-based access control.

### Fixed
- Tab switching not locking in create or edit states.
- Commit date handling (converted to `time.Time`).

### Removed
- Legacy global keys handling in favor of context-specific help.
- Long help view.
