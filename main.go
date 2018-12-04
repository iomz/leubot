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
	mastertoken = app.
			Flag("mastertoken", "The master token for debug.").
			Default("sometoken").
			String()

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

type RobotPose struct {
	Elbow         uint16
	WristAngle    uint16
	WristRotation uint16
	Gripper       uint16
}

func (rp *RobotPose) BuildArmLinkPacket() *armlink.ArmLinkPacket {
	return armlink.NewArmLinkPacket(512, 450, rp.Elbow, rp.WristAngle, rp.WristRotation, rp.Gripper, 128, 0, 0)
}

type Controller struct {
	ArmLinkSerial     *armlink.ArmLinkSerial
	CurrentRobotPose  *RobotPose
	CurrentUserInfo   *api.UserInfo
	HandlerChannel    chan api.HandlerMessage
	LastArmLinkPacket *armlink.ArmLinkPacket
}

func (controller *Controller) Shutdown() {
	// init
	// set the robot in sleep mode
	alp := armlink.ArmLinkPacket{}
	alp.SetExtended(armlink.ExtendedSleep)
	controller.ArmLinkSerial.Send(alp.Bytes())
	// turn off the light
	if *miioenabled {
		cmd := exec.Command(*miiocli, "yeelight", "--ip", *miioip, "--token", *miiotoken, "off")
		cmd.Run()
	}
}

func NewController(als *armlink.ArmLinkSerial) *Controller {
	hmc := make(chan api.HandlerMessage)
	controller := Controller{
		ArmLinkSerial: als,
		CurrentRobotPose: &RobotPose{
			Elbow:         400,
			WristAngle:    580,
			WristRotation: 512,
			Gripper:       128,
		},
		CurrentUserInfo:   &api.UserInfo{},
		HandlerChannel:    hmc,
		LastArmLinkPacket: &armlink.ArmLinkPacket{},
	}

	// init
	// set the robot in sleep mode
	alp := armlink.ArmLinkPacket{}
	alp.SetExtended(armlink.ExtendedSleep)
	controller.ArmLinkSerial.Send(alp.Bytes())
	// turn off the light
	if *miioenabled {
		cmd := exec.Command(*miiocli, "yeelight", "--ip", *miioip, "--token", *miiotoken, "off")
		cmd.Run()
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
					hmc <- api.HandlerMessage{
						Type: api.TypeSomethingWentWrong,
					}
					break
				}
				// check if there's a user already
				if *controller.CurrentUserInfo != (api.UserInfo{}) {
					hmc <- api.HandlerMessage{
						Type: api.TypeUserExisted,
					}
					break
				}
				// register the user to the system
				controller.CurrentUserInfo = &userInfo
				// generate and assign a token to the user
				controller.CurrentUserInfo.Token = api.GenerateToken()
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
					Value: []interface{}{*controller.CurrentUserInfo},
				}
			case api.TypeGetUser:
				hmc <- api.HandlerMessage{
					Type:  api.TypeCurrentUser,
					Value: []interface{}{*controller.CurrentUserInfo},
				}
			case api.TypeDeleteUser:
				// receive the token
				token, ok := msg.Value[0].(string)
				if !ok {
					hmc <- api.HandlerMessage{
						Type: api.TypeSomethingWentWrong,
					}
					break
				}
				// check if the token is valid
				if token != controller.CurrentUserInfo.Token && token != *mastertoken {
					hmc <- api.HandlerMessage{
						Type: api.TypeUserNotFound,
					}
					break
				}
				// delete the current user
				controller.CurrentUserInfo = &api.UserInfo{}
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
			case api.TypePutElbow:
				// receive the robotCommand
				robotCommand, ok := msg.Value[0].(api.RobotCommand)
				if !ok {
					hmc <- api.HandlerMessage{
						Type: api.TypeSomethingWentWrong,
					}
					break
				}
				// check if the token is valid
				if robotCommand.Token != controller.CurrentUserInfo.Token && robotCommand.Token != *mastertoken {
					hmc <- api.HandlerMessage{
						Type: api.TypeInvalidToken,
					}
					break
				}
				// check the value is valid
				if robotCommand.Value < 400 || 650 < robotCommand.Value {
					hmc <- api.HandlerMessage{
						Type: api.TypeInvalidCommand,
					}
					break
				}
				// set the value to CurrentRobotPose
				controller.CurrentRobotPose.Elbow = robotCommand.Value
				// perform the move
				controller.ArmLinkSerial.Send(controller.CurrentRobotPose.BuildArmLinkPacket().Bytes())

				hmc <- api.HandlerMessage{
					Type: api.TypeActionPerformed,
				}
			case api.TypePutWristAngle:
				// receive the robotCommand
				robotCommand, ok := msg.Value[0].(api.RobotCommand)
				if !ok {
					hmc <- api.HandlerMessage{
						Type: api.TypeSomethingWentWrong,
					}
					break
				}
				// check if the token is valid
				if robotCommand.Token != controller.CurrentUserInfo.Token && robotCommand.Token != *mastertoken {
					hmc <- api.HandlerMessage{
						Type: api.TypeInvalidToken,
					}
					break
				}
				// check the value is valid
				if robotCommand.Value < 200 || 830 < robotCommand.Value {
					hmc <- api.HandlerMessage{
						Type: api.TypeInvalidCommand,
					}
					break
				}
				// set the value to CurrentRobotPose
				controller.CurrentRobotPose.WristAngle = robotCommand.Value
				// perform the move
				controller.ArmLinkSerial.Send(controller.CurrentRobotPose.BuildArmLinkPacket().Bytes())

				hmc <- api.HandlerMessage{
					Type: api.TypeActionPerformed,
				}
			case api.TypePutWristRotation:
				// receive the robotCommand
				robotCommand, ok := msg.Value[0].(api.RobotCommand)
				if !ok {
					hmc <- api.HandlerMessage{
						Type: api.TypeSomethingWentWrong,
					}
					break
				}
				// check if the token is valid
				if robotCommand.Token != controller.CurrentUserInfo.Token && robotCommand.Token != *mastertoken {
					hmc <- api.HandlerMessage{
						Type: api.TypeInvalidToken,
					}
					break
				}
				// check the value is valid
				if robotCommand.Value < 0 || 1023 < robotCommand.Value {
					hmc <- api.HandlerMessage{
						Type: api.TypeInvalidCommand,
					}
					break
				}
				// set the value to CurrentRobotPose
				controller.CurrentRobotPose.WristRotation = robotCommand.Value
				// perform the move
				controller.ArmLinkSerial.Send(controller.CurrentRobotPose.BuildArmLinkPacket().Bytes())

				hmc <- api.HandlerMessage{
					Type: api.TypeActionPerformed,
				}
			case api.TypePutGripper:
				// receive the robotCommand
				robotCommand, ok := msg.Value[0].(api.RobotCommand)
				if !ok {
					hmc <- api.HandlerMessage{
						Type: api.TypeSomethingWentWrong,
					}
					break
				}
				// check if the token is valid
				if robotCommand.Token != controller.CurrentUserInfo.Token && robotCommand.Token != *mastertoken {
					hmc <- api.HandlerMessage{
						Type: api.TypeInvalidToken,
					}
					break
				}
				// check the value is valid
				if robotCommand.Value < 0 || 512 < robotCommand.Value {
					hmc <- api.HandlerMessage{
						Type: api.TypeInvalidCommand,
					}
					break
				}
				// set the value to CurrentRobotPose
				controller.CurrentRobotPose.Gripper = robotCommand.Value
				// perform the move
				controller.ArmLinkSerial.Send(controller.CurrentRobotPose.BuildArmLinkPacket().Bytes())

				hmc <- api.HandlerMessage{
					Type: api.TypeActionPerformed,
				}
			case api.TypePutReset:
				// receive the robotCommand
				robotCommand, ok := msg.Value[0].(api.RobotCommand)
				if !ok {
					hmc <- api.HandlerMessage{
						Type: api.TypeSomethingWentWrong,
					}
					break
				}
				// check if the token is valid
				if robotCommand.Token != controller.CurrentUserInfo.Token && robotCommand.Token != *mastertoken {
					hmc <- api.HandlerMessage{
						Type: api.TypeInvalidToken,
					}
					break
				}
				// reset CurrentRobotPose
				controller.CurrentRobotPose = &RobotPose{
					Elbow:         400,
					WristAngle:    580,
					WristRotation: 512,
					Gripper:       128,
				}
				// perform the reset
				alp := armlink.ArmLinkPacket{}
				alp.SetExtended(armlink.ExtendedReset)
				controller.ArmLinkSerial.Send(alp.Bytes())

				hmc <- api.HandlerMessage{
					Type: api.TypeActionPerformed,
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
	defer controller.Shutdown()

	log.Printf("Server started")
	router := api.NewRouter(controller.HandlerChannel)
	log.Fatal(http.ListenAndServe(":6789", router))
}
