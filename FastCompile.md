# Fast compile and autoreload with CompileDaemon

This file explains how to use a compile-watcher (CompileDaemon) to speed up development by automatically rebuilding and restarting your Go binary when source files change.

Note: examples assume Windows (PowerShell) environment and the project root where `main.go` lives.

## Install CompileDaemon

With modern Go toolchain use `go install`:

```powershell
go install github.com/githubnemo/CompileDaemon@latest
# Ensure $GOPATH\bin or $GOBIN is in your PATH so the CompileDaemon executable is available.
```

If you cannot use `go install`, you can `git clone` the repo and `go build` the tool locally.

## Basic usage

From the project root run CompileDaemon and tell it how to build and run your app. On Windows the run command should call the built `.exe`:

```powershell
CompileDaemon -build="go build -o auth.exe ." -command="./auth.exe"
```

Behavior:
- Watches for file changes in the current directory (and subfolders).
- Runs the `-build` command to compile the binary.
- If build succeeds, runs `-command` to start the binary (restarting it on subsequent successful builds).

## Common flags

- `-build="..."` : command used to compile the project (required).
- `-command="..."` : command used to run the compiled binary (required).
- `-directories="dir1,dir2"` : comma-separated list of directories to watch (default `.`).
- `-exclude-dir="vendor,node_modules"` : directories to ignore.
- `-delay=200` : milliseconds delay after change before rebuilding.

Example to watch only the `handlers` and `models` folders and ignore `vendor`:

```powershell
CompileDaemon -directories="handlers,models" -exclude-dir="vendor" -build="go build -o auth.exe ." -command="./auth.exe"
```

## Passing environment variables

Set environment variables before running CompileDaemon so the child process inherits them:

```powershell
$env:PORT = "8000"
$env:MONGODB_URI = "mongodb://localhost:27017"
$env:SECRET_KEY = "dev_secret"
CompileDaemon -build="go build -o auth.exe ." -command="./auth.exe"
```

Or use an env-file loader (eg. `godotenv` inside your app) and keep `.env` in the project root.

## Using with Windows PowerShell specifics

- Use `./auth.exe` to run the built binary.
- If you prefer a different binary name, change `-build` and `-command` accordingly.
- If you run into PATH issues for `CompileDaemon`, run it via its full path: `$env:GOBIN\CompileDaemon.exe` (or `$(go env GOPATH)\bin\CompileDaemon.exe`).

## Advanced: custom build flags and tags

Include build flags (race detector, tags) in the `-build` command:

```powershell
CompileDaemon -build="go build -tags=dev -o auth.exe ." -command="./auth.exe"
```

## Quick workflow

1. Start MongoDB locally.
2. Set env vars (or create `.env`).
3. Run CompileDaemon from project root.
4. Edit files; changes trigger rebuild + restart automatically.

## Troubleshooting

- If changes are not detected: ensure files are saved and not ignored by `-exclude-dir`.
- If CompileDaemon isn't found: add `$GOPATH\bin` or `$GOBIN` to `PATH` and re-open terminal.
- For build errors: CompileDaemon shows compiler output; fix errors in source and save to trigger a new build.

## Alternatives

If you prefer other tools, consider `air` (github.com/cosmtrek/air) or `reflex` — they offer similar file-watching and restart capabilities.

---

Save this file as `FastCompile.md` in the project root for quick reference.