package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8898")
	if err != nil {
		fmt.Println("Can't resolve address: ", err)
		os.Exit(1)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Can't dial: ", err)
		os.Exit(1)
	}
	defer conn.Close()

	buff := [1024]byte{}
	for i := 0; ; i++ {
		str := fmt.Sprintf(`%03d %s`, i, "hello world")

		_, err = conn.Write([]byte(str))
		if err != nil {
			fmt.Println("failed:", err)
			os.Exit(1)
		}

		n := 0
		data := buff[:]
		n, err = conn.Read(data)
		if err != nil {
			fmt.Println("failed to read UDP msg because of ", err)
			os.Exit(1)
		}
		fmt.Println(n, string(data))
		time.Sleep(time.Millisecond * 50)
	}
}
