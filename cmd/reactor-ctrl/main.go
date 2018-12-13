// Copyright (c) 2018 Iori Mizutani
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/Interactions-HSG/ax12ctrl/armlink"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Environmental variables
var (
	// Current Version
	version = "0.1.0"

	// app
	app = kingpin.
		New("reactor-ctrl", "Send a command to control the PhantomX AX-12 Reactor Robot Arm.")

	// flags
	reset = app.
		Flag("reset", "Set Backhoe/Joint mode and go to home.").
		Default("false").
		Bool()

	baseRotation = app.
			Flag("base", "Base rotation [0-1023].").
			Default("512").
			Uint16()

	shoulderRotation = app.
				Flag("shoulder", "Shoulder rotation [205-810].").
				Default("450").
				Uint16()

	elbowRotation = app.
			Flag("elbow", "Elbow rotation [210-900].").
			Default("400").
			Uint16()

	wristAngle = app.
			Flag("wristAngle", "Wrist angle [200-830].").
			Default("580").
			Uint16()

	wristRotation = app.
			Flag("wristRotation", "Wrist rotation [0-1023].").
			Default("512").
			Uint16()

	gripper = app.
		Flag("gripper", "Gripper [0-512].").
		Default("255").
		Uint16()

	delta = app.
		Flag("delta", "Delta [0-254].").
		Default("128").
		Uint16()

	button = app.
		Flag("button", "Button [0-127].").
		Default("0").
		Uint16()

	extended = app.
			Flag("extended", "Extended [0-254].").
			Default("0").
			Uint16()
)

func main() {
	app.Version(version)
	parse := kingpin.MustParse(app.Parse(os.Args[1:]))
	_ = parse

	als := armlink.NewSerial()
	defer als.Close()

	var alp *armlink.Packet

	if *reset {
		alp = armlink.NewPacket(
			0,
			0,
			0,
			0,
			0,
			0,
			0,
			0,
			byte(64), // change Mode to 3D Cylindrical / 90 ° Wrist & Go to Home
		)
	} else {
		// Construct ArmLink Packet based on the flags
		alp = armlink.NewPacket(
			uint16(*baseRotation),     // baseRotation
			uint16(*shoulderRotation), // shoulderRotation
			uint16(*elbowRotation),    // elbowRotation
			uint16(*wristAngle),       // wristAngle
			uint16(*wristRotation),    // wristRotation
			uint16(*gripper),          // gripper
			byte(*delta),              // deltaByte
			byte(*button),             // buttonByte
			byte(*extended),           // extendedInstructionByte
		)
	}
	als.Send(alp.Bytes())
}
