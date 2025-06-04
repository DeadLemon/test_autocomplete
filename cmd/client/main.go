package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	host := flag.String("host", "0.0.0.0", "server host")
	port := flag.Int("port", 10000, "server port")
	flag.Parse()

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text() + "\n"
		if _, err := conn.Write([]byte(text)); err != nil {
			log.Printf("write error: %v", err)
			return
		}
	}
}
