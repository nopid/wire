package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
)

const BASE_PORT = 5000

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func launchdump(inter int, conn *net.TCPConn) {
	var f *os.File
	defer conn.Close()
	f, err := conn.File()
	checkError(err)
	defer f.Close()
	cmd := exec.Command("/usr/bin/tcpdump", "-s", "0", "-n", "-w", "-", "-U", "-i", fmt.Sprintf("eth%d", inter))
	cmd.Env = os.Environ()
	cmd.Stdout = f
	cmd.Stdin = nil
	cmd.Stderr = nil
	err = cmd.Start()
	checkError(err)
}

func dumper(inter int) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", BASE_PORT+inter))
	checkError(err)
	for {
		conn, err := ln.Accept()
		checkError(err)
		go launchdump(inter, conn.(*net.TCPConn))
	}
}

var eth = make(map[string]int)
var nxt = 1

func controler(conn net.Conn) {
	defer conn.Close()
	res, err := ioutil.ReadAll(conn)
	checkError(err)
	toks := strings.Split(strings.Trim(string(res), "\n\t "), " ")
	if len(toks) == 2 && toks[0] == "DUMP" {
		port, ok := eth[toks[1]]
		if ok {
			conn.Write([]byte(fmt.Sprintf("PORT %d\n", BASE_PORT+port)))
		} else {
			eth[toks[1]] = nxt
			conn.Write([]byte(fmt.Sprintf("NEW PORT %d\n", BASE_PORT+nxt)))
			go dumper(nxt)
			nxt = nxt + 1
		}
	} else {
		conn.Write([]byte("SYNTAX ERROR\n"))
	}
}

func main() {
	fmt.Println(">>> Starting wire")
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", BASE_PORT))
	checkError(err)
	for {
		conn, err := ln.Accept()
		checkError(err)
		go controler(conn)
	}
}
