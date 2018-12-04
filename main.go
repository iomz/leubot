/*
 * API for ICSN 2018 Assignment 4
 *
 * This is a simple API for 52-5226
 *
 * API version: 1.0.0
 * Contact: iori.mizutani@unisg.ch
 */

package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/Interactions-HSG/ax12ctrl/api"
	"github.com/Interactions-HSG/ax12ctrl/armlink"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Environmental variables
var (
	// Current Version
	version = "0.1.0"

	// app
	app = kingpin.
		New("ax12ctrl", "Provide a Web API for the PhantomX AX-12 Reactor Robot Arm.")

	// flags
	miioenabled = app.
			Flag("miioenabled", "Enable Xiaomi yeelight device.").
			Default("false").
			Bool()

	miiocli = app.
		Flag("miiocli", "The path to miio cli.").
		Default("/opt/bin/miiocli").
		String()

	miiotoken = app.
			Flag("miiotoken", "The token for Xiaomi yeelight device.").
			Default("0000000000000000000000000000").
			String()

	miioip = app.
		Flag("miioip", "The IP address for Xiaomi yeelight device.").
		Default("192.168.1.2").
		String()
)

type Controller struct {
	ArmLinkSerial     *armlink.ArmLinkSerial
	CurrentUser       *api.UserInfo
	HandlerChannel    chan api.HandlerMessage
	LastArmLinkPacket *armlink.ArmLinkPacket
}

func NewController(als *armlink.ArmLinkSerial) *Controller {
	hmc := make(chan api.HandlerMessage)
	controller := Controller{
		ArmLinkSerial:     als,
		CurrentUser:       &api.UserInfo{},
		HandlerChannel:    hmc,
		LastArmLinkPacket: &armlink.ArmLinkPacket{},
	}

	go func() {
		for {
			msg, ok := <-hmc
			if !ok {
				break
			}

			switch msg.Type {
			case api.TypeAddUser:
				userInfo, ok := msg.Value[0].(api.UserInfo)
				if !ok {
					log.Fatalln("TypeAddUser contains an invalid Value.")
				}
				// check if there's a user already
				if *controller.CurrentUser != (api.UserInfo{}) {
					hmc <- api.HandlerMessage{
						Type: api.TypeUserExisted,
					}
					break
				}
				// register the user to the system
				controller.CurrentUser = &userInfo
				// generate and assign a token to the user
				controller.CurrentUser.Token = api.GenerateToken()
				// turn on the light
				if *miioenabled {
					cmd := exec.Command(*miiocli, "yeelight", "--ip", *miioip, "--token", *miiotoken, "on")
					cmd.Run()
				}
				// set the robot in Joint mode and go to home
				alp := armlink.ArmLinkPacket{}
				alp.SetExtended(armlink.ExtendedReset)
				controller.ArmLinkSerial.Send(alp.Bytes())

				hmc <- api.HandlerMessage{
					Type:  api.TypeUserAdded,
					Value: []interface{}{*controller.CurrentUser},
				}
			case api.TypeGetUser:
				hmc <- api.HandlerMessage{
					Type:  api.TypeCurrentUser,
					Value: []interface{}{*controller.CurrentUser},
				}
			case api.TypeDeleteUser:
				token, ok := msg.Value[0].(string)
				if !ok {
					log.Fatalln("TypeDeleteUser contains an invalid Value.")
				}
				// check if there's an existing user
				if controller.CurrentUser.Token != token {
					hmc <- api.HandlerMessage{
						Type: api.TypeUserNotFound,
					}
					break
				}
				// delete the current user
				controller.CurrentUser = &api.UserInfo{}
				// set the robot in sleep mode
				alp := armlink.ArmLinkPacket{}
				alp.SetExtended(armlink.ExtendedSleep)
				controller.ArmLinkSerial.Send(alp.Bytes())
				// turn off the light
				if *miioenabled {
					cmd := exec.Command(*miiocli, "yeelight", "--ip", *miioip, "--token", *miiotoken, "off")
					cmd.Run()
				}

				hmc <- api.HandlerMessage{
					Type: api.TypeUserDeleted,
				}
			}
		}
		log.Fatalln("HandlerChannel closed, dying...")
	}()

	return &controller
}

func main() {
	app.Version(version)
	parse := kingpin.MustParse(app.Parse(os.Args[1:]))
	_ = parse

	// initialize ArmLink serial interface to control the robot
	als := armlink.NewArmLinkSerial()
	defer als.Close()

	// create the controller with the serial
	controller := NewController(als)

	log.Printf("Server started")
	router := api.NewRouter(controller.HandlerChannel)
	log.Fatal(http.ListenAndServe(":6789", router))
}
