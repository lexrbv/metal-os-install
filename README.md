# metal-os-install
metal-os-install is a simple tool that provides an ability to paritaly automate OS installation with QEMU and rescue mode.

## Rescue mode
Many popular datacenters (e.g Heztner) provide a usefull feature called "rescue system mode". In this mode, your server booted into a lightweight OS distro (usually Debian) via PXE, which in turn allows you to establish a SSH connection and manage your server just as if you were using a LiveCD (but without GUI).

## QEMU OS installation
This dirty-simple hack allows you to install an OS from ISO image being alredy booted into LiveCD OS. We just need to download OS iso image and start QEMU VM with host disk drives mounted as raw devices and ISO mounted as cdrom (`-drive file=/dev/sda,format=raw,media=disk -cdrom=os.iso`). Once the VM is up and running, we can establish a VNC session to manually install the OS with standard installer.

## What does this tool do?
It simply establishes an SSH connection to the server (must be already booted into rescue OS), downloads OS ISO, starts QEMU virtual machine with the specified parameters, and performs VNC port forwarding to localhost for simplifying VNC connection.

## Quick start
1. Boot your server into rescue mode (e.g. https://docs.hetzner.com/robot/dedicated-server/troubleshooting/hetzner-rescue-system/)
2. Prepare a direct link to the OS image you want to install on the server (in this example we will use Ubuntu Server 22.04)
3. Launch metal-os-install tool `./metal-os-install -ssh-host <server-ip>:22 -ssh-user <ssh-user> -ssh-password <ssh-password> -os-iso-url https://releases.ubuntu.com/22.04.4/ubuntu-22.04.4-live-server-amd64.iso -qemu-drives /dev/sda,/dev/sdb -qemu-memory 1G -qemu-cpu 4`
4. Launch any VNC client (I personally recommend Remmina) and connect to `127.0.0.1:53001`
5. Install OS the standard way using VNC
6. Once the OS installation process is complete, you need to disable the rescue system mode in your hosting control panel and reboot the server.
7. Your server should now boot into the installed OS. Enjoy!

## Advanced configuration
### Using Legacy BIOS mode
* By default, the QEMU VM boots into UEFI mode. If you need to use legacy BIOS mode, specify the `-use-uefi=false` flag.

### SSH connection with key-based auth
* If you need to use key-based SSH authentication, simply specify the private key file with `-ssh-private-key` flag.

## Contributing

This is just a simple PoC that I created quick-and-dirty for myself. If you encountered a problem or have a feature idea, feel free to open Issue or Pull Request.

## TODO

1. Add cloud-init support for *nix-based OSes
