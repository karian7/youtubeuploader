# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go CLI tool for uploading videos to YouTube via the YouTube Data API v3. Supports local files, HTTP URLs, and stdin as input sources with OAuth2 authentication, rate limiting, and playlist management.

## Build & Development Commands

```bash
# Build
go build ./cmd/youtubeuploader

# Build with version info
go build -ldflags="-X main.appVersion=1.0.0" ./cmd/youtubeuploader

# Test (with race detection, as used in CI)
go test -v -race ./...

# Lint
golangci-lint run
```

Release builds use GoReleaser (`.goreleaser.yaml`) targeting Linux/Windows/macOS/FreeBSD/OpenBSD.

## Architecture

The root package (`youtubeuploader`) is a library. The CLI entry point is in `cmd/youtubeuploader/main.go`.

### Upload Flow

`main()` parses CLI flags into `Config` → `Open()` opens video source → `limiter.NewLimitTransport()` wraps HTTP transport with rate limiting → `Run()` orchestrates the full upload:
1. Authenticate via OAuth2 (`BuildOAuthHTTPClient` in `oauth.go`)
2. Load metadata from JSON or CLI flags (`LoadVideoMeta` in `files.go`)
3. Upload video via YouTube API
4. Set thumbnail, caption, and add to playlists

### Key Files

- `files.go` — `Config`, `VideoMeta`, `Date` types; `Open()` for multi-source file I/O; `LoadVideoMeta()` for metadata loading
- `run.go` — `Run()` main upload orchestration
- `oauth.go` — OAuth2 three-legged flow with browser-based auth and token caching
- `http.go` — `Playlistx`, `VideoMeta` types; playlist search/create/add logic
- `signal_unix.go` / `signal_windows.go` — Platform-specific signal handling (USR1 for progress on Unix)
- `internal/limiter/` — `LimitTransport` HTTP transport wrapper with time-window rate limiting
- `internal/progress/` — Upload progress reporting with configurable intervals

### OAuth2 Credentials

Searches for `client_secrets.json` and `request.token` in current directory, then `~/.config/youtubeuploader/`. OAuth port defaults to 8080 with redirect URI `http://localhost:8080/oauth2callback`.

## Go Version

Requires Go 1.23.0+ (toolchain go1.24.4).
