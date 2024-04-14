package main

import (
	"flag"
	"fmt"

	"github.com/lexrbv/metal-os-install/build"

	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/SommerEngineering/SSHTunnel/Tunnel"
	"golang.org/x/crypto/ssh"
)

func executeSSHCommand (cmd string, sshClient *ssh.Client) error {
	session, err := sshClient.NewSession()

	if err != nil {
		log.Fatalf("unable to start new session: %v", err)
	}
	defer session.Close()
	err = session.Run(cmd)
	return err
}

func main() {
	var sshUser = flag.String("ssh-user", "root", "SSH user used to connect")
	var sshPassword = flag.String("ssh-password", "", "SSH password used to connect")
	var sshHost = flag.String("ssh-host", "", "SSH target host with port for OS install (e.g mysrv.example.com:22) (required)")
	var sshPrivateKeyPath = flag.String("ssh-private-key", "", "Path to SSH private key used to connect")
	var osIsoURL = flag.String("os-iso-url", "", "Direct URL to OS image (required)")
	var sshWorkDir = flag.String("ssh-work-dir","/tmp/os-install", "Working dir for OS installation without last /")
	var qemuDrives = flag.String("qemu-drives", "", "Comma-separated list of drives to connect to QEMU VM (eg. /dev/sda,/dev/sdb) (required)")
	var qemuCpu = flag.Int("qemu-cpu", 1, "CPU cores count for QEMU VM")
	var qemuMemory = flag.String("qemu-memory", "512m", "Amount of memory for QEMU VM (e.g. 16gb)")
	var packagesInstallCommand = flag.String("packages-install-command", "apt-get install -y curl qemu-system-x86 ovmf", "Command used to install packages in rescue OS")
	var sshTunnelLocalEndpoint = flag.String("ssh-tunnel-local-endpoint", "127.0.0.1:53001" , "The local end-point of the SSH tunnel")
	var sshTunnelRemoteEndpoint = flag.String("ssh-tunnel-remote-endpoint", "127.0.0.1:5901" , "The remote end-point of the SSH tunnel")
	var useUefi = flag.Bool("use-uefi", true, "Specify whether QEMU should boot in UEFI mode")
	var v = flag.Bool("v", false, "Print version")
	var sshTunnelConfig ssh.ClientConfig

	flag.Parse()

	if *v {
		fmt.Println("version:", build.Version)
		return
	}

	// Allow Go to use all CPUs:
	runtime.GOMAXPROCS(runtime.NumCPU())

	if len(*osIsoURL) == 0 || len(*sshHost) == 0 || len(*qemuDrives) == 0 {
        fmt.Println("Usage: ./metal-os-install -ssh-host myserver.example.com -os-image-url https://releases.ubuntu.com/22.04.4/ubuntu-22.04.4-live-server-amd64.iso -qemu-drives /dev/sda,dev/sdb")
        flag.PrintDefaults()
        os.Exit(1)
    }

	if len(*sshPrivateKeyPath) != 0 {
		log.Printf("Loading provided SSH private key...")

		key, err := os.ReadFile(*sshPrivateKeyPath)
		if err != nil {
			log.Fatalf("unable to read private key file:  %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			log.Fatalf("unable to parse private key: %v", err)
		}

		sshTunnelConfig = ssh.ClientConfig{
			User: *sshUser,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	}

	if len(*sshPrivateKeyPath) == 0 && len(*sshPassword) != 0 {
		log.Printf("Using provided SSH password...")
		sshTunnelConfig = ssh.ClientConfig{
			User: *sshUser,
			Auth: []ssh.AuthMethod{
				ssh.Password(*sshPassword),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	}

	log.Printf("Establishing a connection to the server...")
	
	client, err := ssh.Dial("tcp", *sshHost, &sshTunnelConfig)
	executeSSHCommand("ls -l", client)

	if err != nil {
		log.Fatalf("unable to connect: %v", err)
	}
	defer client.Close()

	log.Printf("Creating SSH port-forwarding for VNC...")
	localListener := Tunnel.CreateLocalEndPoint(*sshTunnelLocalEndpoint)
	go Tunnel.AcceptClients(localListener, &sshTunnelConfig, *sshHost, *sshTunnelRemoteEndpoint)

	log.Printf("Ensure working dir exists on the remote server..")
	err = executeSSHCommand("mkdir -p "+*sshWorkDir, client)
	if err != nil {
		log.Fatalf("unable to create working dir: %v", err)
	}

	log.Printf("Ensure packages installed..")
	err = executeSSHCommand(*packagesInstallCommand, client)
	if err != nil {
		log.Fatalf("unable to install packages: %v", err)
	}

	log.Printf("Downloading OS ISO. Please be patient....")
	err = executeSSHCommand("curl -o "+*sshWorkDir+"/os.iso "+*osIsoURL, client)
	if err != nil {
		log.Fatalf("unable to download OS ISO: %v", err)
	}

	log.Printf("Staring QEMU VM....")
	var qemuCommand string = "qemu-system-x86_64 -enable-kvm -cdrom "+*sshWorkDir+"/os.iso" + " -boot d -vnc 127.0.0.1:1 -smp " + 
	strconv.Itoa(*qemuCpu) + " -m " + *qemuMemory

	if *useUefi {
		qemuCommand += " -bios /usr/share/ovmf/OVMF.fd"
	}

	for drives, i := strings.Split(*qemuDrives, ","), 0; i<len(drives); i++ {
		qemuCommand += " -drive file="+drives[i]+",format=raw,media=disk"
	}

	// For debugging
	//log.Printf("Formatted QEMU command: %s", qemuCommand)
	log.Printf("Now you can connect to VM VNC: "+*sshTunnelLocalEndpoint)
	err = executeSSHCommand(qemuCommand, client)
	if err != nil {
		log.Fatalf("unable to start QEMU VM: %v", err)
	}

}
