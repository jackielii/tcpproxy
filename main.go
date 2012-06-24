package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
)

const (
	BUFSIZE = 10240
)

var (
	target      *string = flag.String("target", "dataservice:8080", "proxy destination")
	listen_port *string = flag.String("listen_port", "8080", "port to listen")
)

func pass_through(from, to net.Conn, done chan struct{}) {
	b := make([]byte, BUFSIZE)
	for {
		n, err := from.Read(b)
		if err != nil {
			break
		}
		if n > 0 {
			to.Write(b[:n])
		}
	}
	done <- struct{}{}
}

func start_tunnel(local net.Conn, target string) {
	local_addr := local.LocalAddr().String()
	Info("start to tunnel", local_addr, "to", target)
	remote, err := net.Dial("tcp", target)
	if err != nil {
		Fatal("unable to connect", target)
	}
	done := make(chan struct{})
	go pass_through(local, remote, done)
	go pass_through(remote, local, done)
	<-done
	<-done
	local.Close()
	remote.Close()
	Info("finished tunneling", local_addr, "to", target)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	if flag.NFlag() < 1 || *target == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	ln, err := net.Listen("tcp", ":"+*listen_port)
	if err != nil {
		Fatal("unable to listen to port", *listen_port)
	}
	Info("start to listen on port", *listen_port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			Error("failed to accept the connection, ignored")
			break
		}
		go start_tunnel(conn, *target)
	}
}

func Info(v ...interface{}) {
	log.Print("[Info]: ", fmt.Sprintln(v...))
}

func Error(v ...interface{}) {
	log.Println("[Error]: ", fmt.Sprintln(v...))
}

func Fatal(v ...interface{}) {
	log.Println("[Fatal]: ", fmt.Sprintln(v...))
	os.Exit(1)
}
