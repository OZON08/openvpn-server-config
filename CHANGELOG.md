# Changelog

All notable changes to this project will be documented in this file.

## [v0.5.0] - 2026-10-31

### Security
- Fixed shell injection through the EasyRSA vars file: `GetText` and `SaveToFile` now reject string values containing anything other than letters, digits, spaces and `. , _ @ & + ( ) : -`
- Added optional `MIPassword` to the Management Interface client, so the interface can be protected with a `pw-file`
- Moved the line break check for Management Interface input into the shared send path, which also covers `Execute` and `SendCommand`
- Fixed use of the command as a format string in `SendCommand`
- Server and client configs now reject line breaks in single-value fields, and client configs reject key material that could close its inline block
- These checks do not make configs safe for untrusted input: fields such as `ScriptSecurity`, `Route` and the custom options hold complete OpenVPN directives by design, so only trusted administrators may set them
- `SaveToFile` now tightens existing files to `0600` before changing their content. Files are still updated in place, so containers that bind mount a single file see the change

### Changed
- `go.mod` and `go.sum` added, module path is `github.com/OZON08/openvpn-server-config`, requires Go 1.26.10
- Dependabot now checks Go modules
- Line endings normalized to LF through `.gitattributes`
- Template handling moved into the shared package `internal/render`

### Fixed
- `client/config` template no longer references the missing field `Keysize`, which made rendering fail
- Client templates render the `tls-crypt` block only when `Ta` is set, and `client/config` now includes it
- `ReadResponse` skips asynchronous `>` notifications, which made single-line replies fail to parse
- `ParseStatus` returns an error for `ERROR:` replies and malformed rows instead of reporting no connected clients
- Outdated tests corrected and test coverage added for `client/client-config` and the Management Interface login

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
