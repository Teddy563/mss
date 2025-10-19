# Mineplus Autostart Proxy

Mineplus is a lightweight proxy that keeps your Paper (or any Java edition) server dormant until a player connects. It sits in front of the Minecraft server, wakes it up on demand, and shuts it down again when the server is idle. The console output is flattened to `[Mineplus] message`, so Pterodactyl panels and log tailers stay clean.

This repository also provides a ready-to-import Pterodactyl egg that bundles the proxy with Paper.

## Highlights

- **On-demand start/stop** – keep the server offline when nobody is playing and automatically start it when a player joins.
- **Clean console** – Mineplus rewrites its own logs and Paper’s Log4j output to a simple `[Mineplus]` prefix.
- **Suspend or stop** – optionally suspend the Java process instead of issuing `/stop` for near instant resumes.
- **Hidden egg controls** – the Pterodactyl egg exposes only standard Paper options; Mineplus tuning lives in hidden variables or the JSON template.

## Quick Start (Standalone)

1. Install Go 1.21+.
2. Clone the repository and build the proxy:
   ```shell
   go build -o mineplus-proxy .
   ```
3. Copy `mineplus-config.json` next to the proxy binary and adjust the paths/ports.
4. Run Mineplus:
   ```shell
   ./mineplus-proxy -prefix Mineplus -quiet
   ```
   Use `-quiet=false` to see the detailed debug output.

## Configuration

The proxy looks for `mineplus-config.json`. The JSON structure mirrors the original MSH config but the sections and keys are renamed so that any branding is removed.

```jsonc
{
  "Server": {
    "Folder": "/path/to/paper",
    "FileName": "server.jar",
    "Version": "1.21.4",
    "Protocol": 767
  },
  "Commands": {
    "StartServer": "java <Commands.StartServerParam> -jar <Server.FileName> nogui",
    "StartServerParam": "-Xms1024M -Xmx1024M",
    "StopServer": "stop",
    "StopServerAllowKill": 10
  },
  "Auto": {
    "LogLevel": 1,
    "InstanceID": "",
    "ProxyPort": 25555,
    "ProxyQueryPort": 25555,
    "EnableQuery": true,
    "IdleTimeout": 30,
    "UseSuspend": false,
    "SuspendRefreshSeconds": -1,
    "InfoIdle": "                   \\u00a7fserver status:\\n                   \\u00a7b\\u00a7lSTANDBY",
    "InfoWarming": "                   \\u00a7fserver status:\\n                    \\u00a76\\u00a7lWARMING UP",
    "NotifyUpdate": true,
    "NotifyMessage": true,
    "Whitelist": [],
    "WhitelistImport": false,
    "ShowResourceUsage": false,
    "ShowInternetUsage": false
  }
}
```

### Flags

- `-prefix Mineplus` – text printed before every log line (`[Mineplus] …`).  
- `-quiet` – suppress wrapper INFO/BYTE logs (error and Paper logs are always shown).  
- `-port`, `-timeout`, `-suspendallow`, `-suspendrefresh`, etc. map 1:1 to the configuration file and remain compatible with the original CLI.

## Pterodactyl Egg

- JSON file: `egg-paper-mineplus-autostart.json`
- Imports as **Mineplus Paper Autostart**
- Hidden variables keep the idle timeout, suspend settings, and debug level locked. Change them in the JSON before importing if you need different defaults.
- The installer downloads the corresponding Mineplus binary from the releases of this fork and generates a `log4j2.xml` that forces the `[Mineplus] %msg` pattern.

Refer to [EGG-README.md](EGG-README.md) for the panel specific walkthrough.

## Building Releases

Mineplus ships binaries for Linux x86_64 and arm64. To create them locally:

```shell
GOOS=linux GOARCH=amd64 go build -o releases/mineplus-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -o releases/mineplus-linux-arm64 .
```

Upload the two files to a GitHub release so the egg installer can fetch them.

## Credits & License

Mineplus is a heavily rebranded fork of [gekware/minecraft-server-hibernation](https://github.com/gekware/minecraft-server-hibernation) (originally by @gekigek99). The project remains licensed under the GPL-3.0; please review `LICENSE` for full terms and retain the appropriate notices when redistributing.
