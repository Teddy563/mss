# Minecraft Server Color Logs Fix

## What Was Fixed

Previously, all Minecraft server logs appeared in **gray color** in MSH output, which made them difficult to read and removed the helpful color coding that Minecraft uses.

### Before (Gray Logs)
```
2025/01/18 21:00:00.000 [serv    ] [00:00:00] [Server thread/INFO]: Starting minecraft server version 1.21.4
```
All server output was wrapped in gray, making it hard to distinguish different log types.

### After (Original Colors Preserved)
```
2025/01/18 21:00:00.000 [serv    ] [00:00:00] [Server thread/INFO]: Starting minecraft server version 1.21.4
```
Server logs now display with their original ANSI color codes, making INFO (white), WARN (yellow), ERROR (red), etc. properly visible.

## Technical Details

### The Problem
In `lib/errco/errco.go`, the `TYPE_SER` (server) log type was wrapping all messages in `COLOR_GRAY`:

```go
case TYPE_SER:
    typ = fmt.Sprintf("%s%-6s%s", COLOR_GRAY, string(logMod.Typ), COLOR_RESET)
    ori = "\x00"
    mex = fmt.Sprintf("%s%s%s", COLOR_GRAY, StringGraphic(...), COLOR_RESET) // ❌ Wrapped in gray
    cod = "\x00"
```

### The Solution
Removed the color wrapping to preserve original ANSI codes:

```go
case TYPE_SER:
    typ = fmt.Sprintf("%s%-6s%s", COLOR_GRAY, string(logMod.Typ), COLOR_RESET)
    ori = "\x00"
    mex = StringGraphic(fmt.Sprintf(logMod.Mex, logMod.Arg...)) // ✅ Original colors preserved
    cod = "\x00"
```

## Files Modified

- `lib/errco/errco.go` - Line 131: Removed `COLOR_GRAY` wrapping from server messages
- All binaries rebuilt with the fix

## Benefits

1. **Better Readability**: Server logs are now color-coded as intended by Minecraft
2. **Easier Debugging**: Different log levels (INFO, WARN, ERROR) are visually distinct
3. **Professional Look**: Matches the appearance of running Minecraft directly
4. **No Breaking Changes**: Only affects visual output, no functional changes

## Rebuilt Binaries

All platform binaries have been rebuilt with the color fix:

- ✅ `msh-linux-amd64` - Linux x86_64
- ✅ `msh-linux-arm64` - Linux ARM64  
- ✅ `msh-darwin-amd64` - macOS Intel
- ✅ `msh-darwin-arm64` - macOS Apple Silicon
- ✅ `msh-windows-amd64.exe` - Windows x64

## How to Use

1. Download the latest binaries from the releases folder
2. Replace your existing `msh` binary
3. Restart your server
4. Enjoy colorful, readable logs!

## Example Output

With this fix, you'll now see:
- **Blue/Cyan** - Player join/leave messages
- **White** - INFO level logs
- **Yellow** - WARN level logs  
- **Red** - ERROR level logs
- **Green** - Success messages

The logs will look exactly like they do when running Minecraft server directly, making it much easier to monitor and debug your server.
