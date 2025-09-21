package client

import (
	"fmt"
	"net/rpc"
	"os"
	"bufio"
	"strings"
	"os/signal"
    "syscall"
	"time"
	"sync"
	"strconv"
)

var totalMatches = 0

// tests valid and invalid grep requests
func TestGrep(client *rpc.Client) {
	cmds := []string{
		"grep \"package main\" ../server/server.go", 
		"grep -i \"func\" ../server/server.go",
		"grep -n \"func main\" ../server/server.go",
		"grep -A 2 \"import\" ../server/server.go",
		"grep -E \"error\" ../server/server.go",
		"grep -E \"&&\" ../server/server.go",

		"grep --noexist \"main\" ../server/server.go",
		"grep \"main\" noexist.go",         
		"grep \"main\" ../server/server.go | ls",       
		"grep \"[badregex\" ../server/server.go",              
		"gre \"package main\" ../server/server.go",        
	}

	for _,cmd := range cmds {
		var reply string 
		err := client.Call("VM.Grep", cmd, &reply)
		// -1 placeholder
		Printer(-1, cmd, reply, err)	
	}

}

// tests functionality on seperate doc
func TestSpec(client *rpc.Client) {
	cmd := "grep -n \"main\" ../log/log.txt"
	var reply string
	err := client.Call("VM.Grep", cmd, &reply)
	// -1 placeholder
	Printer(-1, cmd, reply, err)	
}

func Printer(vm_no int, cmd string, reply string, err error) {
	fmt.Print("\n------------------------------\n" + 
	"vm number: " + strconv.Itoa(vm_no) + "\n" +	
	cmd + 
	"\n------------------------------\n")

	if err != nil /* strings.HasPrefix(err.Error(), "error:") */ {
		fmt.Println(err.Error())
	} else {
		fmt.Println(reply)
	}
}

// calls the RPC grep function registered by the server
// once we set up the VMs we would call the RPC function on every server
func Call(vm_no int, cmd string, client *rpc.Client) error {
	err := CheckConnection(client)
	if err != nil {
		return err
	}

	var reply string
	err = client.Call("VM.Grep", cmd, &reply)
	Printer(vm_no, cmd, reply, err)
	
	var numMatches = 0
	lines := strings.Split(reply, "\n")
	if len(lines) > 0 && strings.Contains(lines[len(lines)-1], "MATCHES") {
    	numMatches, _ = strconv.Atoi(lines[len(lines)-1][9:])
	}
	totalMatches += numMatches


	return err
}

// calls the RPC confirm connection function registered by the server
// once we set up the VMs we would validiate there is a valid connection before calling grep
func CheckConnection(client *rpc.Client) error {
	client_name := "client" // make this unique per VM
	var reply string
	return client.Call("VM.ConfirmConnection", client_name, &reply)
}

// if is_signal -> closes client due to a signal
// else -> closes client due to user request
func Kill(is_signal bool) {
	if is_signal {
		fmt.Println("\nsignal recieved")
	}
	fmt.Println("exiting gracefully...")
	os.Exit(0)
}

func Connect() []*rpc.Client {
	// ip addresses for VMs 01-10
	ip_adds := []string{
		// ":4425", // localhost for testing
		"172.22.159.124:4425",
		"172.22.155.198:4425",
		"172.22.155.125:4425",
		"172.22.159.125:4425",
		"172.22.155.199:4425",
		"172.22.155.126:4425",
		"172.22.159.126:4425",
		"172.22.155.200:4425",
		"172.22.155.127:4425",
		"172.22.159.127:4425",
	}

	// create a splice to hold VM connections
	vms := make([]*rpc.Client, len(ip_adds))
	var conn = 0
	var dialNum = 0
	for conn != 1 {
		if (dialNum == 2) {
			fmt.Println("\nFAILED TO CONNECT TO ANY VM!\n")
			os.Exit(1)
		} else if dialNum == 1 {
			fmt.Println("\nATTEMPTING TO CONNECT AGAIN\n")
			time.Sleep(1 * time.Second)
		}

		for i, ip := range ip_adds {
			vm, err := rpc.DialHTTP("tcp", ip)
			if err != nil {
				fmt.Printf("error dialing on vm %02d: %v\n", i + 1, err)
				vms[i] = nil
				continue
			}
			vms[i] = vm
		}

		for _,vm := range vms {
			if vm != nil {
				conn = 1
				break
			}
		}

		time.Sleep(1 * time.Second)
		dialNum = dialNum + 1
	}

	return vms
}

func Client() {
	// create a channel which asynchronously checks for kill signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	
	// create a buffered io to read from stdin
	reader := bufio.NewReader(os.Stdin)
	
	// sleep to ensure server connection is secure
	
	for {
		vms := Connect()
		fmt.Print("\nenter a command: ")
		
		// channels for input
		inputChan := make(chan string, 1)
		errChan := make(chan error, 1)
		go func() {
			input, err := reader.ReadString('\n')
			if (err != nil) {
				errChan <- err
			} else {
				inputChan <- input
			}	
		}()

		select {
			// close client if signal recieved
		case <-signalChan:
			Kill(true)
			// client.Close()
			return

			// close client if failed to read in from stdin
		case err := <-errChan:
			Kill(false)
			fmt.Println("\nerror reading input:", err)
			return

			// process input
		case input := <-inputChan:
			input = strings.TrimSpace(input)
			if input == "exit" || input == "quit" {
				// client.Close()
				Kill(false)
			}

			var totalLatency time.Duration = 0
			totalMatches = 0
    		var wg sync.WaitGroup

			for i, vm := range vms {
				if vm == nil {
					continue;
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					start := time.Now()
					err := Call(i+1, input, vm)
					if err == nil {
						t := time.Now()
						elapsed := t.Sub(start)
						totalLatency += elapsed
						fmt.Printf("LATENCY: %s\n", elapsed)
					}
				}()
				wg.Wait()
			}
			fmt.Print("\n------------------------------\n" + "RESULTS" + "\n------------------------------\n")
			fmt.Println("AVERAGE LATENCY:", totalLatency / 10)
			fmt.Printf("TOTAL MATCHES: %d\n\n", totalMatches)
		}
	}
}
