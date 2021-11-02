package main

import (
	"bufio"
	"fmt"
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
	eth := fmt.Sprintf("eth%d", inter)
	var f *os.File
	defer conn.Close()
	f, err := conn.File()
	checkError(err)
	defer f.Close()
	cmd := exec.Command("/usr/bin/dumpcap", "-s", "0", "-n", "-w", "-", "-i", eth)
	cmd.Env = os.Environ()
	cmd.Stdout = f
	cmd.Stdin = nil
	cmd.Stderr = nil
	err = cmd.Start()
	checkError(err)
}

func dumper(inter int, listening chan bool) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", BASE_PORT+inter))
	checkError(err)
	listening <- true
	for {
		conn, err := ln.Accept()
		checkError(err)
		go launchdump(inter, conn.(*net.TCPConn))
	}
}

var netname = make(chan string)
var ethnum = make(chan int)

func memory() {
	eth := make(map[string]int)
	listening := make(chan bool, 1)
	nxt := 1
	for {
		name := <-netname
		port, ok := eth[name]
		if ok {
			ethnum <- port
		} else {
			eth[name] = nxt
			go dumper(nxt, listening)
			<-listening
			ethnum <- -nxt
			nxt = nxt + 1
		}
	}
}

func controler(conn net.Conn) {
	defer conn.Close()
	res, err := bufio.NewReader(conn).ReadString('\n')
	checkError(err)
	toks := strings.Split(strings.Trim(string(res), "\n\t "), " ")
	if len(toks) == 2 && toks[0] == "DUMP" {
		netname <- toks[1]
		port := <-ethnum
		if port > 0 {
			conn.Write([]byte(fmt.Sprintf("PORT %d\n", BASE_PORT+port)))
		} else {
			conn.Write([]byte(fmt.Sprintf("NEW PORT %d\n", BASE_PORT-port)))
		}
	} else {
		conn.Write([]byte("SYNTAX ERROR\n"))
	}
}

func main() {
	go memory()
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", BASE_PORT))
	checkError(err)
	fmt.Println("READY")
	for {
		conn, err := ln.Accept()
		checkError(err)
		go controler(conn)
	}
}
