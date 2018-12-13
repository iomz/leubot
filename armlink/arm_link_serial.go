package armlink

import (
	"encoding/hex"
	"io"
	"log"

	"github.com/jacobsa/go-serial/serial"
)

// Serial holds the serial interface to the Arbotix Robocontroller
type Serial struct {
	port io.ReadWriteCloser
}

// NewSerial is a constructor for Serial
func NewSerial() *Serial {
	als := &Serial{}

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

// Close closes the serial connection
func (als *Serial) Close() {
	als.port.Close()
}

// Send sends the given byte array
func (als *Serial) Send(b []byte) {
	log.Println(hex.Dump(b))
	_, err := als.port.Write(b)
	if err != nil {
		log.Fatalf("port.Write: %v", err)
	}
}
