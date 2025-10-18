# Pterodactyl Egg for Minecraft Server Hibernation

This is the official Pterodactyl egg for running Paper Minecraft servers with automatic hibernation using MSH (Minecraft Server Hibernation).

## 📥 Installation

### Download Egg
The egg is available in two locations:
- **Repository Root**: `egg-paper-minecraft-hibernation.json`
- **Releases Directory**: `releases/egg-paper-minecraft-hibernation.json`
- **Direct Download**: [Latest Release on GitHub](https://github.com/Kartvya69/minecraft-server-hibernation/releases/latest)

### Upload to Pterodactyl
1. Go to your Pterodactyl Panel **Admin Area**
2. Navigate to **Nests** → **Minecraft** (or create a new nest)
3. Click **Import Egg**
4. Upload `egg-paper-minecraft-hibernation.json`
5. Configure the egg settings as needed

### Create Server
1. Go to **Servers** → **Create New Server**
2. Select the **Paper Minecraft Hibernation** egg
3. Configure server settings (RAM, CPU, disk space)
4. Click **Create Server**

## 🚀 Features

### Automatic Hibernation
- Server automatically hibernates when no players are online
- Wakes up instantly when a player tries to join
- Saves resources and reduces costs

### Supported Java Versions
- Java 21 (Minecraft 1.20.5+)
- Java 17 (Minecraft 1.18+)
- Java 16 (Minecraft 1.17)
- Java 11 (Minecraft 1.16)
- Java 8 (Minecraft <1.16)
- J9 variants available for reduced RAM usage

### Supported Architectures
- x86_64 (amd64) - Most common
- aarch64 (arm64) - ARM-based servers

## ⚙️ Configuration

### Server Variables
The egg includes several configurable variables:

#### Minecraft Settings
- **Minecraft Version**: Paper version to install (e.g., `1.21.4`, `latest`)
- **Server Jar File**: Name of the server JAR file (default: `server.jar`)
- **Build Number**: Paper build number (default: `latest`)

#### MSH Settings
- **Wait time for shutdown**: Time in seconds before hibernating (default: `120`)
- **TimeToKill**: Force kill timeout if graceful shutdown fails (default: `30`)
- **Debug Level**: Logging verbosity 1-3 (default: `2`)
- **SuspendAllow**: Use process suspension instead of stop (default: `1`)
- **SuspendRefresh**: Seconds between suspend refreshes (default: `30`)
- **wlimport**: Only allow whitelisted players to start server (default: `0`)

### Binary Download
The egg **automatically downloads** the latest MSH binaries from:
```
https://github.com/Kartvya69/minecraft-server-hibernation/releases/latest/download/
```

No configuration needed! The URL is hardcoded in the installation script.

## 🔧 How It Works

1. **Installation Phase**:
   - Downloads Paper server JAR from PaperMC API
   - Downloads MSH binary from GitHub releases
   - Verifies binary integrity (ELF format check)
   - Downloads default `server.properties` and `msh-config.json`

2. **Runtime**:
   - MSH acts as a proxy between players and the Minecraft server
   - When a player connects, MSH starts the server
   - After configured timeout with no players, server hibernates
   - MSH continues listening for new connections

## 🐛 Troubleshooting

### Binary Download Fails
If you see: `ERROR: Failed to download MSH binary`
- Check your server has internet access
- Verify GitHub releases are accessible: https://github.com/Kartvya69/minecraft-server-hibernation/releases
- Check firewall rules allow HTTPS connections

### Binary Verification Fails
If you see: `ERROR: Downloaded file is not a valid ELF binary`
- The downloaded file might be an error page
- Check the error output showing first 200 bytes
- Verify the release exists on GitHub

### Server Won't Start
1. Check Java version matches Minecraft version
2. Review server logs in Pterodactyl console
3. Verify `msh-config.json` is present and valid
4. Check `msh_server.bin` has execute permissions

## 📚 More Information

- **Main Repository**: https://github.com/Kartvya69/minecraft-server-hibernation
- **Original Project**: https://github.com/gekware/minecraft-server-hibernation
- **Discord Support**: Join our [discord server](https://discord.com/invite/guKB6ETeMF)

## 📝 License

This egg configuration is based on the original work by [BolverBlitz](https://github.com/gekware/minecraft-server-hibernation-pterodactyl-egg) and modified for the Kartvya69 fork.

MSH itself is licensed under the GPL-3.0 License.
