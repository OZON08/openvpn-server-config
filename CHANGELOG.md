# Changelog

All notable changes to this project will be documented in this file.

## [v0.4.0] - 2026-04-12

### Security
- Fixed command injection vulnerability in `KillSession` and `Signal` via newline characters in Management Interface inputs
- Changed Management Interface default binding from `0.0.0.0` to `127.0.0.1` to prevent public exposure
- Added bounds checks in `ParseStatus` and `ParseStats` to prevent panics on malformed MI responses
- Changed config file write permissions from `0644` to `0600` (owner-only) for all generated config files
- Fixed ignored error return from `template.Execute` in all config packages — prevents silently writing corrupt configs
- Added connection timeout (10s) and read deadline to Management Interface client
- Added maximum line limit (10,000) to `ReadResponse` to prevent unbounded memory growth
- Upgraded default cipher from `AES-256-CBC` to `AES-256-GCM` in client config
- Aligned default HMAC from `SHA256` to `SHA512` in client configs to match server default

## [v0.3.5]

### Added
- 2FA authentication mode support (cert + OTP) with `FuncMode` flag
- `AuthUserPass` and `TFAIssuer` fields in client config for 2FA flows
- Three custom config option fields (`CustomOptOne/Two/Three`) for server and client
- `TLSControlChannel`, `TLSMinVersion`, `TLSRemoteCert` fields for TLS hardening
- EasyRSA vars template and config package
- Management Interface client (`server/mi`) with `GetPid`, `GetVersion`, `GetStatus`, `GetLoadStats`, `KillSession`, `Signal`, `RestartServer`
- `data-ciphers` directive replacing deprecated `ncp-ciphers`
- Server restart via `SIGUSR1` signal

### Changed
- Migrated from `html/template` to `text/template` for OpenVPN config generation
- Updated default server cipher to `AES-256-GCM` and auth to `SHA512`
- Deprecated `keysize` and `comp-lzo` options commented out in client template
- Removed deprecated OpenVPN parameters

### Fixed
- Port field handling in client config
