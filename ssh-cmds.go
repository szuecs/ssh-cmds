package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var Ch chan string = make(chan string)

func run_cmd(client *ssh.Client, cmd string) ([]string, error) {
	session, _ := client.NewSession()
	defer session.Close()
	var stderrBuf, stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf
	session.Run(cmd)
	s := stdoutBuf.String()
	err := stderrBuf.String()
	if err != "" {
		fmt.Printf("ERR: %+v\n")
	}
	return strings.Split(s, "\n"), nil
}

func dial(server string, port int, config *ssh.ClientConfig, wg *sync.WaitGroup, cmds []string) {
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", server, port), config)
	if err != nil {
		Ch <- fmt.Sprintf("ERR: Failed to dial to %s: %s\n", server, err)
		wg.Done()
		return
	}

	var tmpList, resultList []string

	// run commands
	for _, cmd := range cmds {
		tmpList, err = run_cmd(client, cmd)
		resultList = append(resultList, tmpList...)
	}

	// filter duplicates on one host
	var resultBuf bytes.Buffer
	if err == nil {
		for _, item := range resultList {
			if item != "" {
				resultBuf.Write([]byte(fmt.Sprintf("%s;%s\n", server, item)))
			}
		}
	}

	// send to write to STDOUT
	Ch <- resultBuf.String()
	wg.Done()
}

func out(wg *sync.WaitGroup) {
	for s := range Ch {
		fmt.Printf("%s", s)
		wg.Done()
	}
}

// implement []string for args
type cmdSlice []string

func (i *cmdSlice) String() string {
	return strings.Join(*i, ", ")
}

func (i *cmdSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var cmds cmdSlice
	var user string
	var port int
	flag.Var(&cmds, "cmd", "Commands to run.")
	flag.StringVar(&user, "user", "root", "User to access remote hosts, defaults to root.")
	flag.IntVar(&port, "port", 22, "Port to access remote hosts, defaults to 22.")
	flag.Parse()

	auth_socket := os.Getenv("SSH_AUTH_SOCK")
	if auth_socket == "" {
		log.Fatal(errors.New("no ENV $SSH_AUTH_SOCK defined"))
	}
	conn, err := net.Dial("unix", auth_socket)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ag := agent.NewClient(conn)
	auths := []ssh.AuthMethod{ssh.PublicKeysCallback(ag.Signers)}
	config := &ssh.ClientConfig{
		User: user,
		Auth: auths,
	}

	var wg sync.WaitGroup
	reader := bufio.NewReader(os.Stdin)

	go out(&wg) // print lines
	// connect to servers
	for {
		server, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		server = server[:len(server)-1] // chomp
		wg.Add(2)                       // dial and print
		go dial(server, port, config, &wg, cmds)
	}
	wg.Wait()
}
