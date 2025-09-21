package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os/exec"
	"log"
	"strconv"
	"strings"
	"os"
	"mvdan.cc/sh/v3/shell"
)

type VM struct{
	listener net.Listener
}

// checks input for malicious commands
func Validate(tokens []string) error {
	lower := strings.ToLower(tokens[0])

	// early return for trivial request
	if lower == "grep" && (len(tokens) == 1) {
		return nil
	}

	// validate user actually made a grep request
	if lower != "grep" {
		errMsg := fmt.Sprintf("error command not found: %s", lower)
		return errors.New(errMsg)
	}

	return nil

}

// this is an RPC function that can be called remotely
//
// turns a cmd e.g. grep [flags] "pattern" filename into a splice
// calls exec on the splice and returns either the output or error
func (vm *VM) Grep(str string, reply *string) error {
	tokens, err := shell.Fields(str, nil)

	if err != nil {
		return err
	}

	log.Print(tokens)

	err = Validate(tokens)

	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	
	switch {
		case strings.Contains(hostname, "b401"): 
			tokens = append(tokens, "../log/vm1.log")
		case strings.Contains(hostname, "b402"):
			tokens = append(tokens, "../log/vm2.log")
		case strings.Contains(hostname, "b403"): 
			tokens = append(tokens, "../log/vm3.log")
		case strings.Contains(hostname, "b404"): 
			tokens = append(tokens, "../log/vm4.log")
		case strings.Contains(hostname, "b405"): 
			tokens = append(tokens, "../log/vm5.log")
		case strings.Contains(hostname, "b406"): 
			tokens = append(tokens, "../log/vm6.log")
		case strings.Contains(hostname, "b407"): 
			tokens = append(tokens, "../log/vm7.log")
		case strings.Contains(hostname, "b408"): 
			tokens = append(tokens, "../log/vm8.log")
		case strings.Contains(hostname, "b409"): 
			tokens = append(tokens, "../log/vm9.log")
		case strings.Contains(hostname, "b410"):
			tokens = append(tokens, "../log/vm10.log")
		default:
			tokens = append(tokens, "../log/log.txt")
	}
	
	cmd := exec.Command(tokens[0], tokens[1:]...)

	// CombinedOutputs merges both the stdout and stderr, as well as an error code
	out, err := cmd.CombinedOutput()
	log.Println(string(out))

	if err != nil {
		log.Println(err.Error())

		// exit status 1 -> no match found
		// exit status 2 -> invalid input
		// else -> something else went wrong
		
		if err.Error() == "exit status 1" {
			err = errors.New("error: no match found")
			return err
		}  

		if err.Error() != "exit status 2" {
			err = errors.New("error: catastophic failure")
			return err
		}
	}			

	output := string(out)

	counter_args := append([]string{"-c"}, tokens[1:]...)
	cmd_counter := exec.Command(tokens[0], counter_args...)

	out_counter, err := cmd_counter.CombinedOutput()

	if err == nil {
		output = output + "\nMATCHES: " + strings.TrimSpace(string(out_counter))
	}

	*reply = output
	return nil
}

// this is an RPC function which can be called remotely
// 
// verifies a connection is made to a client
func (vm *VM) ConfirmConnection(str string, reply *string) error {
	*reply = fmt.Sprintf("status: connected to %s", str)
	log.Println(*reply)
	return nil
}

func main() {
	vm := &VM{}
	const portno int = 4425
	
	 if err := rpc.Register(vm); err != nil {
		log.Fatalf("error registering %v", err)	
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(portno))
	if err != nil {
		log.Fatalf("error listening: %v", err)
	}

	vm.listener = listener

	log.Printf("listening on port %d\n", portno)
	   if err := http.Serve(listener, nil); err != nil {
	       log.Fatalf("error serving: %v", err)
	   }
}
