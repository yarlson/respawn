# Windows Compatibility Guide

Turbine is designed to work cross-platform, including Windows. This guide covers platform-specific considerations and how to set up Turbine on Windows.

## Prerequisites

### Required Software

1. **Go 1.22 or later** - Download from https://go.dev/dl/
2. **Git for Windows** - Download from https://git-scm.com/download/win
3. **Claude CLI** or **OpenCode CLI** - Depending on your backend choice

### Optional Requirements for Full Features

- **Windows Terminal** - For better terminal experience (recommended)
- **PowerShell 7+** - For modern shell features
- **Developer Mode** - For true symlink support (see "Symlinks on Windows" below)

## Installation

### Install Turbine

```powershell
go install github.com/yarlson/turbine@latest
```

Or build from source:

```powershell
cd $env:TEMP
git clone https://github.com/yarlson/turbine.git
cd turbine
go build -o turbine.exe .
```

### Set Up Configuration

Create the config directory:

```powershell
$configDir = "$env:APPDATA\turbine"
mkdir -Force $configDir | Out-Null
```

Create configuration file at `$env:APPDATA\turbine\turbine.yaml`:

```yaml
defaults:
  backend: claude
  retry:
    rotations: 3
    strokes: 3

backends:
  claude:
    command: claude
    args:
      - -p
      - --dangerously-skip-permissions
    models:
      fast: claude-3-5-sonnet-latest
      slow: claude-4-5-opus-latest
```

## Symlinks on Windows

Turbine creates a `CLAUDE.md` symlink to `AGENTS.md` for tool compatibility. Windows has limited symlink support compared to Unix/Linux.

### Option 1: Enable Developer Mode (Recommended)

Starting with Windows 10, you can enable Developer Mode to allow unprivileged symlink creation:

1. Open **Settings** > **Update & Security** > **For developers**
2. Enable **Developer Mode**
3. Restart your computer

After enabling Developer Mode, symlinks will work without requiring admin privileges.

### Option 2: Run as Administrator

If you don't want to enable Developer Mode, run Turbine with administrator privileges:

```powershell
Start-Process powershell -Verb runAs
# Then run turbine commands as normal
```

### Option 3: Use File Copy Fallback

If symlinks cannot be created, the agent will automatically copy `AGENTS.md` to `CLAUDE.md` as a fallback. This works identically for tool purposes but uses more disk space.

## Known Differences

### Path Handling

Windows uses backslashes (`\`) as path separators. Turbine automatically handles this:

```powershell
# Both work the same way
turbine agents --prd "path\to\prd.md"
turbine agents --prd "path/to/prd.md"  # Forward slashes also work
```

### Environment Variables

On Windows, use `%APPDATA%` or PowerShell's `$env:` prefix:

```powershell
# PowerShell
$env:XDG_CONFIG_HOME = "C:\config"

# Command Prompt
set XDG_CONFIG_HOME=C:\config
```

### Line Endings

Git on Windows uses `CRLF` line endings by default. Turbine handles this transparently, but be aware:

- YAML files may have CRLF line endings
- Task verification scripts should account for this

Configure Git to use Unix line endings for committed files:

```powershell
git config core.autocrlf input
```

### Temporary Directories

Turbine uses system temp directories for artifacts:

```powershell
# Windows temp directory
$env:TEMP    # Usually C:\Users\<user>\AppData\Local\Temp
$env:TMP     # Alternative

# Turbine stores logs in .turbine/runs/ (relative to repo)
```

## Troubleshooting

### Issue: "command not found: turbine"

**Solution:** Ensure Go's bin directory is in your PATH:

```powershell
# Check if turbine is installed
go list -m github.com/yarlson/turbine

# Add Go bin to PATH if needed
$env:PATH += ";$env:GOPATH\bin"

# Make it permanent (add to profile)
```

### Issue: "Symlink creation failed"

**Solution:** This is non-fatal. The agent will automatically fall back to copying the file. Check `.turbine/runs/` logs for details.

To enable true symlinks:
1. Enable Developer Mode (recommended)
2. Or run as administrator
3. Or upgrade to Windows 11 with Developer Mode enabled

### Issue: Backend command not found

**Solution:** Install the backend (Claude CLI or OpenCode) and ensure it's in PATH:

```powershell
# Check if claude is available
where claude

# If not found, install from:
# https://github.com/anthropics/anthropic-sdk-python
```

### Issue: Git commands fail

**Solution:** Ensure Git for Windows is installed and in PATH:

```powershell
where git
git --version
```

If Git is installed but not in PATH, add it manually to your PATH environment variable.

### Issue: File permissions denied

**Solution:** On Windows, sometimes file handles remain open. Try:

```powershell
# Close any editors that might have files open from the repo
# Then retry the operation

# Or manually delete and retry
Remove-Item -Force -Recurse .turbine\runs
turbine  # Retry
```

## Best Practices

1. **Use Administrator PowerShell** for initial setup to enable full feature support
2. **Enable Developer Mode** if you plan to use Turbine regularly
3. **Keep Git for Windows updated** for better compatibility
4. **Use VS Code or similar** for editing - it handles Windows paths well
5. **Check `.turbine/runs/` logs** if tasks fail - they contain detailed error information

## Testing on Windows

To verify Turbine works correctly on your Windows system:

```powershell
# Check installation
turbine --help

# Test a simple command
$tempDir = New-Item -ItemType Directory -Force "temp-prd"
$prdContent = "# Test PRD`n`nSimple CLI tool in Go"
$prdContent | Out-File "$tempDir/PRD.md" -Encoding UTF8

# Run agents command (requires valid credentials)
cd $tempDir
git init
turbine agents --prd PRD.md --yes --backend claude
```

## Performance Considerations

Windows may be slightly slower than Unix/Linux due to:

- File system differences (NTFS vs ext4)
- Virus scanner interference (configure exclusions for `.turbine/` directory)
- Network file systems (avoid running on network drives)

To optimize performance:

1. Add `.turbine/` to antivirus exclusions
2. Keep your project on local SSD storage
3. Avoid storing projects in cloud-synced directories (OneDrive, Dropbox, etc.)

## Further Help

For issues specific to Windows, check:

- Windows documentation: https://docs.microsoft.com/en-us/windows/
- Git for Windows issues: https://github.com/git-for-windows/git/issues
- Turbine issues: https://github.com/yarlson/turbine/issues

Feel free to report Windows-specific issues with detailed error messages and your Windows version.
