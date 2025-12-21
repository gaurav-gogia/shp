# shp - Simple Container Runtime

A lightweight toy container that uses Linux namespaces and filesystem isolation to run commands in a containerized environment.

## Overview

shp (simple host process) creates isolated process environments using Linux containers. It supports both `pivot_root` (preferred) and `chroot` for filesystem isolation, with automatic fallback if `pivot_root` is not available.

## Features

- Linux namespace isolation (UTS, PID, Mount)
- Filesystem isolation using `pivot_root` with fallback to `chroot`
- Custom rootfs path support
- Automatic syscall detection and fallback

## Usage

```bash
shp run <rootfs_path> <cmd> [options]
```

### Parameters

- `<rootfs_path>`: Absolute or relative path to a Linux rootfs directory
- `<cmd>`: Command to execute inside the container
- `[options]`: Arguments to pass to the command

### Example

```bash
# Extract an Ubuntu rootfs
tar -xf ubuntu-rootfs.tar.gz -C /tmp/

# Run a command inside the container
./shp run /tmp/ubuntu bash

# Run a specific command
./shp run /tmp/ubuntu ls -la /
```

## How It Works

1. **Namespace Isolation**: Creates new UTS, PID, and Mount namespaces for isolation
2. **Root Detection**: Automatically tries `pivot_root` first, then falls back to `chroot` if unavailable
3. **Mount Points**: Automatically mounts the proc filesystem inside the container
4. **Command Execution**: Executes the specified command with full namespace isolation

## Downloading Linux Rootfs

### Ubuntu

```bash
# Download and extract Ubuntu root filesystem
mkdir -p ~/containers/ubuntu
cd ~/containers/ubuntu

# Using debootstrap (recommended - most reliable)
sudo apt-get install debootstrap
sudo debootstrap --arch=amd64 jammy . http://archive.ubuntu.com/ubuntu/

# OR download pre-built container image from Docker Hub
# Extract Ubuntu container image as rootfs
docker export $(docker create ubuntu:22.04) | tar -x

# OR use CD image from Ubuntu
wget https://cdimage.ubuntu.com/ubuntu-base/releases/24.04/release/ubuntu-base-24.04.3-base-amd64.tar.gz
tar -xf rootfs.tar.xz
```

### Fedora

```bash
# Download and extract Fedora root filesystem
mkdir -p ~/containers/fedora
cd ~/containers/fedora

# Using dnf (requires Fedora host - most reliable)
sudo dnf install --installroot=$(pwd) --releasever=39 bash coreutils

# OR download pre-built container image from Docker Hub
# Extract Fedora container image as rootfs
docker export $(docker create fedora:39) | tar -x

tar -xf rootfs.tar.xz
```

### Arch Linux

```bash
# Download and extract Arch Linux root filesystem
mkdir -p ~/containers/arch
cd ~/containers/arch

# Using pacstrap (recommended - most reliable)
sudo pacman -S arch-install-scripts
sudo pacstrap -c . base bash

# OR download pre-built container image from Docker Hub
# Extract Arch Linux container image as rootfs
docker export $(docker create archlinux:latest) | tar -x

tar -xf rootfs.tar.xz
```

## Building from Source

```bash
go build -o shp shp.go
```

## Requirements

- Linux kernel with namespace support
- Go 1.20 or later
- Root or appropriate Linux capabilities for namespace creation

## Syscall Behavior

### pivot_root
- **Preferred method** for filesystem isolation
- Requires the new root to be on a different filesystem
- More efficient and cleaner than chroot
- Supported on most modern Linux systems

### chroot (Fallback)
- Used if `pivot_root` fails
- Changes the root directory for the process
- Less efficient but widely supported
- Falls back automatically if `pivot_root` is unavailable

## Environment Variables

The container inherits environment variables from the host. You can override them by passing them before the command:

```bash
./shp run /tmp/ubuntu bash -c "echo $PATH"
```

## Limitations

- Requires Linux host
- Network isolation not implemented
- Resource limits (cgroups) not implemented
- Does not set up user namespaces (requires elevated privileges)

## Example Workflow

```bash
# 1. Create container directory
mkdir -p /tmp/mycontainer

# 2. Download and extract rootfs
cd /tmp/mycontainer
sudo debootstrap --arch=amd64 focal . http://archive.ubuntu.com/ubuntu/

# 3. Build shp
cd ~/shp
go build -o shp shp.go

# 4. Run container
sudo ./shp run /tmp/mycontainer bash
```

## Troubleshooting

### "permission denied" error
- Run with sudo: `sudo ./shp run /path/to/rootfs bash`

### "rootfs path does not exist"
- Verify the path is correct and properly extracted
- Use absolute paths: `/tmp/mycontainer` instead of `../mycontainer`

### "pivot_root failed"
- This is expected if the rootfs is on the same filesystem
- The tool will automatically fall back to chroot
- To use pivot_root, the rootfs must be on a different mount point

## License

MIT
