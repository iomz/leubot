package main

import (
	"time"

	"github.com/Interactions-HSG/ax12ctrl/armlink"
)

func main() {
	als := armlink.NewArmLinkSerial()
	defer als.Close()

	// Set Backhoe/Joint mode and go to home
	als.Send([]byte{
		0xff,
		0x0,
		0x0,
		0x0,
		0x0,
		0x0,
		0x0,
		0x0,
		0x0,
		0x0,
		0x0,
		0x0,
		0x0,
		0x0,
		0x0,
		0x40,
		0xbf,
	})
	time.Sleep(2 * time.Second)

	// Default
	alp := armlink.NewArmLinkPacket(
		uint16(512), // baseRotation
		uint16(512), // shoulderRotation
		uint16(512), // elbowRotation
		uint16(512), // wristAngle
		uint16(512), // wristRotation
		uint16(512), // gripper
		byte(128),   // deltaByte
		byte(0),     // buttonByte
		byte(0),     // extendedInstructionByte
	)
	als.Send(alp.Bytes())
	time.Sleep(2 * time.Second)

	// Something new
	alp = armlink.NewArmLinkPacket(
		uint16(512), // baseRotation
		uint16(512), // shoulderRotation
		uint16(700), // elbowRotation
		uint16(400), // wristAngle
		uint16(400), // wristRotation
		uint16(256), // gripper
		byte(128),   // deltaByte
		byte(0),     // buttonByte
		byte(0),     // extendedInstructionByte
	)
	als.Send(alp.Bytes())
}
