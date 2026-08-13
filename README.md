# BedrockHelper

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go version](https://img.shields.io/github/go-mod/go-version/end2endzone/BedrockHelper?filename=src%2Fgo.mod)](src/go.mod)

`bedrock_helper` is a command line tool for installing, uninstalling and inspecting Minecraft Bedrock Edition add-on packs (`.mcaddon` / `.mcpack`) on a Minecraft Bedrock Dedicated Server (BDS).



## Status

Build:

| Platform             | Build                                                                                                                                                            | Tests                                                                                                                                                                                                          |
| -------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Windows Server        | [![Build on Windows](https://github.com/end2endzone/BedrockHelper/actions/workflows/build_windows.yml/badge.svg)](https://github.com/end2endzone/BedrockHelper/actions/workflows/build_windows.yml) | [![Tests on Windows](https://img.shields.io/endpoint?url=https%3A%2F%2Fgist.githubusercontent.com%2Fend2endzone%2F58cf6c72c08e706335337d5ef9ca48e8%2Fraw%2FBedrockHelper.main.Windows.json)](https://github.com/end2endzone/BedrockHelper/actions/workflows/build_windows.yml) |
| Ubuntu                | [![Build on Linux](https://github.com/end2endzone/BedrockHelper/actions/workflows/build_linux.yml/badge.svg)](https://github.com/end2endzone/BedrockHelper/actions/workflows/build_linux.yml)       | [![Tests on Linux](https://img.shields.io/endpoint?url=https%3A%2F%2Fgist.githubusercontent.com%2Fend2endzone%2F58cf6c72c08e706335337d5ef9ca48e8%2Fraw%2FBedrockHelper.main.Linux.json)](https://github.com/end2endzone/BedrockHelper/actions/workflows/build_linux.yml) |
| macOS                  | [![Build on macOS](https://github.com/end2endzone/BedrockHelper/actions/workflows/build_macos.yml/badge.svg)](https://github.com/end2endzone/BedrockHelper/actions/workflows/build_macos.yml)       | [![Tests on macOS](https://img.shields.io/endpoint?url=https%3A%2F%2Fgist.githubusercontent.com%2Fend2endzone%2F58cf6c72c08e706335337d5ef9ca48e8%2Fraw%2FBedrockHelper.main.macOS.json)](https://github.com/end2endzone/BedrockHelper/actions/workflows/build_macos.yml) |



## Purpose

Minecraft Bedrock Dedicated Server does not ship with any tooling for managing add-on packs. Installing a single `.mcaddon` file by hand means:

1. Unzipping the archive.
2. Figuring out which of the extracted folders is a behavior pack and which is a resource pack by reading their `manifest.json` files.
3. Finding the server's *active* world directory (which depends on the `level-name` property inside `server.properties`).
4. Moving each pack folder into the world's `behavior_packs/` or `resource_packs/` directory.
5. Hand-editing `world_behavior_packs.json` / `world_resource_packs.json` to add the pack's `uuid` and `version` so the server actually loads it on boot.
6. Repeating steps 1-5, in reverse, and error-free, whenever a pack needs to be removed.

This is tedious, error-prone, and hard to automate. `bedrock_helper` does all of the above from a single command, which makes it possible to script and automate add-on deployment (CI/CD pipelines, provisioning scripts, fleet management for multiple servers, etc.) instead of doing it by hand over ssh.



## Features

- Install and uninstall `.mcaddon` (multiple packs) and `.mcpack` (single pack) add-on files, as well as plain `.zip` archives with the same internal layout.
- Uninstall a pack by UUID alone, when the original add-on file is no longer available.
- Automatically locates the server's active world by reading `level-name` from `server.properties` (falling back to the first world found under `worlds/` if the property is absent).
- Automatically detects whether an extracted pack is a `BehaviorPack` or a `ResourcePack` from its `manifest.json` module types.
- Reads and writes `world_behavior_packs.json` / `world_resource_packs.json` for you, so packs are always correctly registered (or unregistered) with the server.
- Recursively scans a directory for files that look like add-on packs (`--find-addons`).
- Lists every pack currently registered on a server, resolving each UUID to a human-readable name and version (`--list-addons`).
- Finds which add-on file (if any) contains a pack matching a given UUID (`--resolve-pack`).
- Batch install or uninstall every add-on found under a server directory in one call (`--install-all` / `--uninstall-all`).
- Zip-slip-safe extraction: archive entries can never write outside of the intended destination directory.
- Single, dependency-free, statically linked binary (`CGO_ENABLED=0`) for Windows, Linux and macOS.



## Use cases

- Automating add-on deployment as part of a server provisioning script or CI/CD pipeline.
- Rolling out (or rolling back) an add-on across a fleet of Bedrock servers with one command per server.
- Auditing what is actually installed and registered on a server with `--list-addons`.
- Recovering from a lost/misplaced add-on file by uninstalling a pack using only its UUID.
- Bulk-migrating add-ons between two server installations or synchronizing add-ons between multiple servers.



## Installation


### Prerequisites

- [Go](https://go.dev/dl/). For exact version see [`src/go.mod`](src/go.mod) or the badge at the top of this document).


### Build from source

```console
git clone https://github.com/end2endzone/BedrockHelper.git
cd BedrockHelper/src
go build -o bedrock_helper ./cmd/bedrock_helper
```

This produces a single `bedrock_helper` (or `bedrock_helper.exe` on Windows) executable with no runtime dependencies. Copy it anywhere on your `PATH`.



### Build using the CI scripts

The scripts under [`ci/`](ci) are what the GitHub Actions workflows use, and can also be run locally. They embed the version, commit hash and build date into the binary via `-ldflags`, and place the result in `bin/` at the repository root.

They produce a single executable named `bedrock_helper-{os}-{cpuarch}` with no runtime dependencies. For example `bedrock_helper-darwin-amd64`, `bedrock_helper-windows-amd64.exe`, etc. Copy it anywhere on your `PATH`.


Linux / macOS:

```bash
./ci/linux/build.sh
```

Windows (PowerShell or Command Prompt):

```powershell
.\ci\windows\build.ps1
```

```batch
.\ci\windows\build.bat
```

`GOOS` and `GOARCH` can be set beforehand to cross-compile (e.g. `GOOS=linux GOARCH=arm64 ./ci/linux/build.sh` to build a Linux ARM64 binary from any host).



## Usage

```
bedrock_helper --install <path> [--server-location <dir>] [--no-header]
bedrock_helper --uninstall <path-or-uuid> [--server-location <dir>] [--no-header]
bedrock_helper --find-addons <path> [--no-header]
bedrock_helper --list-addons [--server-location <dir>] [--no-header]
bedrock_helper --resolve-pack <uuid> [--server-location <dir>] [--no-header]
bedrock_helper --install-all [--server-location <dir>] [--no-header]
bedrock_helper --uninstall-all [--server-location <dir>] [--no-header]
bedrock_helper --version
bedrock_helper --help
```

Only one command may be specified per invocation.

| Argument                    | Description                                                                                       |
| --------------------------- | ------------------------------------------------------------------------------------------------- |
| `--find-addons <path>`      | Search the directory at `<path>` recursively for files that look like add-on packs and list them. |
| `--install <path>`          | Install the `.mcaddon`/`.mcpack`/`.zip` add-on at `<path>`.                                       |
| `--install-all`             | Scan the target server directory for add-on files and install every one that is found.            |
| `--list-addons`             | List the add-on packs currently registered for the target server.                                 |
| `--resolve-pack <uuid>`     | Search the target server for an add-on file that contains a pack matching `<uuid>`.               |
| `--server-location <dir>`   | Target Minecraft Bedrock server directory. Optional; defaults to the current directory.           |
| `--uninstall <path\|uuid>`  | Uninstall the add-on at `<path>`, or by pack UUID if the original add-on file is unavailable.     |
| `--uninstall-all`           | Scan the target server directory for add-on files and uninstall every one that is found.          |
| `--no-header`               | Do not show the product header when running a command.                                            |
| `--version`                 | Show the product version.                                                                         |
| `--help`                    | Show the usage message.                                                                           |

### Example - installing an add-on

```console
$ bedrock_helper --install ./foobar.mcaddon --server-location /srv/bedrock_server
```

```
bedrock_helper - install and manage Minecraft Bedrock add-on packs.
Version 0.0.0 (1234567) compiled on 1900-01-01.
Installed the following packs:
  - Foobar BP version 1.0.0 (BehaviorPack) uuid=1f9498b2-9576-4053-8bad-77afc4221df2
  - Foobar RP version 1.0.0 (ResourcePack) uuid=24fbe7e6-9239-4e6d-a52b-9f67b17e74ce
in server D:\Projets\Programmation\Go\BedrockHelper\master/.debug-sandbox/testdata/servers/server_empty
```

`foobar.mcaddon` bundled one behavior pack and one resource pack; both were extracted, moved into place and registered with the server's active world in a single command.


### More examples

```txt
# Install a single add-on for a specific server
bedrock_helper --install $HOME/foobar.mcaddon --server-location $HOME/myserverinstalldir

# Uninstall that same add-on later
bedrock_helper --uninstall $HOME/foobar.mcaddon --server-location $HOME/myserverinstalldir

# Uninstall a pack when the original add-on file was lost, using its UUID instead
bedrock_helper --uninstall 2bda6085-9d71-4d8a-9b9f-74e07b30459c --server-location $HOME/myserverinstalldir

# Recursively search a directory for add-on files
bedrock_helper --find-addons $HOME/Downloads

# List everything currently registered on a server
bedrock_helper --list-addons --server-location $HOME/myserverinstalldir

# Find which add-on file on the server contains a given pack UUID
bedrock_helper --resolve-pack "2bda6085-9d71-4d8a-9b9f-74e07b30459c" --server-location $HOME/myserverinstalldir

# Install every add-on file found anywhere under the server directory
bedrock_helper --install-all --server-location $HOME/myserverinstalldir

# Uninstall every add-on file found anywhere under the server directory
bedrock_helper --uninstall-all --server-location $HOME/myserverinstalldir
```

`--list-addons` output looks like this:

```
KIND          NAME       VERSION  UUID
BehaviorPack  Foobar BP  1.0.0    2bda6085-9d71-4d8a-9b9f-74e07b30459c
ResourcePack  Foobar RP  1.0.0    33333333-3333-3333-3333-333333333333
```


## How it works

A Minecraft Bedrock Dedicated Server directory looks roughly like this once a couple of add-ons are installed:

```
bedrock_server/
├── bedrock_server(.exe)
├── server.properties          # level-name=Bedrock level
└── worlds/
    └── Bedrock level/         # the *active* world, named after level-name
        ├── behavior_packs/
        │   └── foobar_BP/
        │       └── manifest.json
        ├── resource_packs/
        │   └── foobar_RP/
        │       └── manifest.json
        ├── world_behavior_packs.json
        └── world_resource_packs.json
```



## Test using the CI scripts

The test scripts under [`ci/`](ci) are what the GitHub Actions workflows use for testing, and can also be run locally.

Run the test suite:

Linux / macOS:

```bash
./ci/linux/test.sh
```

Windows (PowerShell or Command Prompt):

```powershell
.\ci\windows\test.ps1
```

```batch
.\ci\windows\test.bat
```

The tests use [testify](https://github.com/stretchr/testify)'s `assert`/`require` packages and a set of fixture add-ons and fake server directories under `src/testdata/`, so they don't touch a real Minecraft installation.


### Debugging in VS Code

[`.vscode/launch.json`](.vscode/launch.json) defines one debug configuration per command line shown in `--help`, each pre-configured with fixture arguments so you can hit F5 and step straight into any code path:

* `--install`
* `--uninstall`
* `--find-addons`
* `--list-addons`
* `--resolve-pack`
* `--install-all`
* `--uninstall-all`

Configurations that install or uninstall packs run against a disposable copy of `src/testdata` under `.debug-sandbox/` (refreshed automatically before every run by the `debug: reset sandbox` task), so debugging never modifies the checked-in test fixtures.


## Platform support

`bedrock_helper` is a pure Go, `CGO_ENABLED=0` binary and is built and tested on:

- Windows (Windows 11, latest)
- Linux (Ubuntu, latest)
- macOS (latest)


## Versioning

This project uses [Semantic Versioning 2.0.0](https://semver.org/). The current version is tracked in the [`VERSION`](VERSION) file at the repository root, and is embedded into the binary at build time.


## Disclaimer

This is an independent, unofficial tool. It is not affiliated with, endorsed by, or associated with Mojang Studios or Microsoft. "Minecraft" is a trademark of Mojang Synergies AB.



## Author

- **Antoine Beauchamp** - [end2endzone](https://github.com/end2endzone)



## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
