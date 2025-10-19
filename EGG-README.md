# Mineplus Paper Autostart Egg

This repository ships a Pterodactyl egg that bundles Paper with the Mineplus proxy wrapper. The proxy keeps the panel console clean, automatically starts the Paper server when a player connects, and stops it again when the server is idle.

## Download & Import

1. Download `egg-paper-mineplus-autostart.json` from the repository root (or from the latest release once available).
2. In your Pterodactyl admin panel open **Nests → Import Egg** and upload the JSON file.
3. Create a server using the new **Mineplus Paper Autostart** egg.

## What the Installer Does

- Downloads the requested Paper build (or the latest available build) from the PaperMC API.
- Fetches the matching Mineplus proxy binary for your architecture from this fork’s releases.
- Drops a default `mineplus-config.json`, writes a `log4j2.xml` that formats the Paper console as `[Mineplus] message`, and creates an executable `mineplus-start.sh` launcher.
- Keeps the familiar `server.properties` defaults from the upstream Paper egg.

## Runtime Behaviour

- `mineplus-proxy` listens on the allocation port exposed by Pterodactyl.
- When a player pings or joins, Mineplus starts `java -jar server.jar`. If `AUTO_START_STOP` is disabled (hidden egg variable) the panel simply runs Java directly.
- When the server stays empty for the configured idle timeout, Mineplus sends `/stop` (or suspends the process if suspend mode is enabled) and returns to standby.

## Key Variables

Most variables are hidden to keep the panel uncluttered. The important ones remain editable:

| Name | Visible | Purpose |
| --- | --- | --- |
| `SERVER_JARFILE` | ✅ | Paper jar to launch (`server.jar` by default). |
| `MINECRAFT_VERSION` / `BUILD_NUMBER` | ✅ | Paper release selection (blank = latest). |
| `JAVA_FLAGS` | ✅ | Applied only when Mineplus is bypassed (direct Java mode). |
| `AUTO_START_STOP` | ❌ | Hidden toggle that forces the launcher to choose Java directly (`0`) or the proxy (`1`). |
| `IDLE_SECONDS` | ❌ | Hidden idle timeout (default `30`). |
| `FORCE_STOP_AFTER` | ❌ | Hidden force-kill timeout (default `10`). |
| `USE_PROCESS_SUSPEND` | ❌ | Hidden toggle to suspend the Paper process instead of stopping it. |

All hidden variables can still be edited in the JSON before import if you need different defaults.

## Requirements

- Pterodactyl Panel v1.11+ (PTDL_v2 eggs).
- Java 17+ image from `ghcr.io/pterodactyl/yolks` (already wired in the egg).
- x86_64 or arm64 compute nodes (matching Mineplus binaries are provided by this fork).

## Support

This egg and the Mineplus proxy are maintained in the `Teddy563/mss` fork. Please open GitHub issues in this repository for bug reports or feature requests.
