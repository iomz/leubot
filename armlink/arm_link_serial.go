package armlink

import (
	"encoding/hex"
	"io"
	"log"

	"github.com/jacobsa/go-serial/serial"
)

type ArmLinkSerial struct {
	port io.ReadWriteCloser
}

func NewArmLinkSerial() *ArmLinkSerial {
	als := &ArmLinkSerial{}

	// Set up options.
	options := serial.OpenOptions{
		PortName:        "/dev/ttyUSB0",
		BaudRate:        38400,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	als.port = port

	return als
}

func (als *ArmLinkSerial) Close() {
	als.port.Close()
}

func (als *ArmLinkSerial) Send(b []byte) {
	log.Println(hex.Dump(b))
	_, err := als.port.Write(b)
	if err != nil {
		log.Fatalf("port.Write: %v", err)
	}
}
