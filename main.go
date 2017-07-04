package main

import (
	"bufio"
	"fmt"
	"net"
)

//const node string = "[fe80::e046:9aff:fe4e:912e%wlp1s0]:33123"

//const node string = "[fe80::e046:9aff:fe4e:912e%enp2s0]:33123"

//const node string = "[fe80::1e8f:814e:9731:dec6%enp2s0]:33123"

const node string = "[::1]:33123"

func main() {

	text := "dump\n"
	conn, err := net.Dial("tcp6", node)
	if err != nil {
		fmt.Println("node")
		fmt.Println(err)

		return
	}

	defer conn.Close()
	fmt.Fprintf(conn, text)
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("no")
			break
		}
		fmt.Println(message)
		if message == "ok" || message == "bad" || message == "no" {
			break
		}
	}

}
