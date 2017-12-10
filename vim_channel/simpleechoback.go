package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	l, e := net.Listen("tcp", ":6868")
	if e != nil {
		return
	}
	for {
		c, e := l.Accept()
		if e != nil {
			continue
		}
		var b [500]byte
		n, e := c.Read(b[:])
		if e != nil {
			break
		}
		fmt.Println(string(b[:n]))
		c.Write(b[:n])
		time.Sleep(100 * time.Millisecond)
		c.Close()
		break
	}
}
