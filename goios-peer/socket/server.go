package socket

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
)

var destinations = map[string]string{
	"usbmuxd": "UNIX:/var/run/usbmuxd",
	//"SOCK2": "UNIX:/tmp/sock2",
	//"SOCK3": "UNIX:/tmp/sock3",
	//"screen":  "TCP:127.0.0.1:3333",
	//"TCP2":  "TCP:127.0.0.1:9999",
}

func proxy(a, b net.Conn) {
	go func() {
		io.Copy(a, b)
		a.Close()
		b.Close()
	}()
	go func() {
		io.Copy(b, a)
		a.Close()
		b.Close()
	}()
}

func handleConn(tcpConn net.Conn) {
	defer tcpConn.Close()
	reader := bufio.NewReader(tcpConn)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Handshake read error:", err)
		return
	}
	key := strings.TrimSpace(line)
	dst, ok := destinations[key]
	if !ok {
		tcpConn.Write([]byte("Unknown destination key\n"))
		return
	}
	parts := strings.SplitN(dst, ":", 2)
	if len(parts) != 2 {
		tcpConn.Write([]byte("Invalid destination format\n"))
		return
	}
	var remoteConn net.Conn
	if parts[0] == "UNIX" {
		remoteConn, err = net.Dial("unix", parts[1])
	} else if parts[0] == "TCP" {
		remoteConn, err = net.Dial("tcp", parts[1])
	} else {
		tcpConn.Write([]byte("Unknown protocol\n"))
		return
	}
	if err != nil {
		log.Println("Dial remote error:", err)
		tcpConn.Write([]byte("Failed to connect to destination\n"))
		return
	}
	go io.Copy(remoteConn, reader) // Передаем остатки после handshake
	proxy(tcpConn, remoteConn)
}

func Start() {
	ln, err := net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	log.Println("Proxy server listening on :12345")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleConn(conn)
	}
}
