package main

import (
	"fmt"
	"github.com/tarm/serial"
	"io"
	"log"
)

func main() {
	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600}
	s, err := serial.OpenPort(c)

	if err != nil {
		log.Fatal(err)
	}
	fo := io.Discard

	var n int
	buf := make([]byte, 128)
	for {
		n, err = s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := fo.Write(buf[:n]); err != nil {
			panic(err)
		}
		fmt.Printf("%s", buf[:n])
	}
}
