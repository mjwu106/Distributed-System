package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
)

// runs a given string command on all 10 VMs, authenticated with client config
func Run(cmd string, config *ssh.ClientConfig) {
	var wg sync.WaitGroup

	for vmNumber := 1; vmNumber <= 10; vmNumber++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vm := fmt.Sprintf("fa25-cs425-b4%02d.cs.illinois.edu:22", vmNumber)
			client, err := ssh.Dial("tcp", vm, config)
			if err != nil {
				log.Printf("failed to dial: %v", err)
			}
			defer client.Close()

			session, err := client.NewSession()
			if err != nil {
				log.Printf("failed to create session: %v", err)
			}
			defer session.Close()

			session.Stdout = os.Stdout
			session.Stderr = os.Stderr

			err = session.Run(cmd)

		}()
	}
	wg.Wait()
}

func ClearLogs(config *ssh.ClientConfig) {
	var wg sync.WaitGroup

	for vmNumber := 1; vmNumber <= 10; vmNumber++ {
		wg.Add(1)
		go func(vmNumber int) {
			defer wg.Done()
			vm := fmt.Sprintf("fa25-cs425-b4%02d.cs.illinois.edu:22", vmNumber)
			client, err := ssh.Dial("tcp", vm, config)
			if err != nil {
				log.Printf("failed to dial %s: %v\n", vm, err)
				return
			}
			defer client.Close()

			for i := 1; i <= 10; i++ {
				if i == vmNumber {
					continue
				}

				session, err := client.NewSession()
				if err != nil {
					log.Printf("failed to create session: %v", err)
					continue
				}

				session.Stdout = os.Stdout
				session.Stderr = os.Stderr
				cmd := fmt.Sprintf("cd ~/mw128/log && rm -f vm%d.log", i)

				err = session.Run(cmd)

				session.Close()
			}
		}(vmNumber)
	}
	wg.Wait()
}

func main() {
	// custom path needed for each user
	// perhaps make the keyPath a cmd line arg
	privateKeyPath := "C:\\Users\\mjwu1\\.ssh\\id_ed25519"
	privateKey, err := os.ReadFile(privateKeyPath)

	if err != nil {
		log.Fatalf("error: unable to open file")
	}

	signer, err := ssh.ParsePrivateKey(privateKey)

	if err != nil {
		log.Fatalf("error: unable to process key")
	}

	config := &ssh.ClientConfig{
		User: "mw128",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// different cmd line args can be passed to execute cmds on all VMs
	if len(os.Args) == 2 {
		if os.Args[1] == "wake" {
			cmd := "pkill -9 server; cd ~/cs-425-mp-1/server && go run server.go"
			Run(cmd, config)
			return
		}
		if os.Args[1] == "pull" {
			cmd := "cd ~/cs-425-mp-1/server && git fetch origin && git reset --hard origin/main"
			Run(cmd, config)
			return
		}
		if os.Args[1] == "kill" {
			cmd := "pkill -9 server"
			Run(cmd, config)
			return
		}
		if os.Args[1] == "log" {
			ClearLogs(config)
			return
		}
	}
	fmt.Println("usage: go run startup.go [ wake | kill | pull | log]")
}
