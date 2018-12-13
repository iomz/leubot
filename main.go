/*
 * API for ICSN 2018 Assignment 4
 *
 * This is a simple API for PhantomX AX-12 Reactor Robot Arm
 * Currently only supports the Backhoe/Joint mode
 *
 * API version: 1.0.0
 * Contact: iori.mizutani@unisg.ch
 */

package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/Interactions-HSG/ax12ctrl/api"
	"github.com/Interactions-HSG/ax12ctrl/armlink"
	"github.com/badoux/checkmail"
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
	masterToken = app.
			Flag("masterToken", "The master token for debug.").
			Default("sometoken").
			String()

	miioEnabled = app.
			Flag("miioEnabled", "Enable Xiaomi yeelight device.").
			Default("false").
			Bool()

	miiocli = app.
		Flag("miiocli", "The path to miio cli.").
		Default("/opt/bin/miiocli").
		String()

	miioToken = app.
			Flag("miioToken", "The token for Xiaomi yeelight device.").
			Default("0000000000000000000000000000").
			String()

	miioIP = app.
		Flag("miioIP", "The IP address for Xiaomi yeelight device.").
		Default("192.168.1.2").
		String()

	slackAppEnabled = app.
			Flag("slackAppEnabled", "Enable Slack app for user previleges.").
			Default("false").
			Bool()
	slackWebHookURL = app.
			Flag("slackWebHookURL", "The webhook url for posting the json payloads.").
			Default("https://hooks.slack.com/services/...").
			String()

	userTimeout = app.
			Flag("userTimeout", "The timeout duration for users in seconds.").
			Default("900").
			Int()
)

// RobotPose stores the current pose of the robot
type RobotPose struct {
	Elbow         uint16
	WristAngle    uint16
	WristRotation uint16
	Gripper       uint16
}

// BuildPacket builds the corresponding Packet from the RobotPose
func (rp *RobotPose) BuildPacket() *armlink.Packet {
	return armlink.NewPacket(512, 450, rp.Elbow, rp.WristAngle, rp.WristRotation, rp.Gripper, 128, 0, 0)
}

// Controller is a main controller instance for the ax12ctrl
type Controller struct {
	ArmLinkSerial    *armlink.Serial
	CurrentRobotPose *RobotPose
	CurrentUser      *api.User
	HandlerChannel   chan api.HandlerMessage
	UserActChannel   chan bool
	UserTimer        *time.Timer
	UserTimerFinish  chan bool
}

// Shutdown performs the grace shutting down process
func (controller *Controller) Shutdown() {
	// init
	// set the robot in sleep mode
	alp := armlink.Packet{}
	alp.SetExtended(armlink.ExtendedSleep)
	controller.ArmLinkSerial.Send(alp.Bytes())
	// turn off the light
	if *miioEnabled {
		cmd := exec.Command(*miiocli, "yeelight", "--ip", *miioIP, "--token", *miioToken, "off")
		cmd.Run()
	}
}

// NewController yields a new Controller instance
func NewController(als *armlink.Serial) *Controller {
	hmc := make(chan api.HandlerMessage)
	controller := Controller{
		ArmLinkSerial: als,
		CurrentRobotPose: &RobotPose{
			Elbow:         400,
			WristAngle:    580,
			WristRotation: 512,
			Gripper:       128,
		},
		CurrentUser:     &api.User{},
		HandlerChannel:  hmc,
		UserActChannel:  make(chan bool),
		UserTimer:       time.NewTimer(time.Second * 10),
		UserTimerFinish: make(chan bool),
	}
	controller.UserTimer.Stop()

	// init
	// set the robot in sleep mode
	alp := armlink.Packet{}
	alp.SetExtended(armlink.ExtendedSleep)
	controller.ArmLinkSerial.Send(alp.Bytes())
	// turn off the light
	if *miioEnabled {
		cmd := exec.Command(*miiocli, "yeelight", "--ip", *miioIP, "--token", *miioToken, "off")
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
				// check if the email is valid
				if err := checkmail.ValidateFormat(userInfo.Email); err != nil {
					hmc <- api.HandlerMessage{
						Type: api.TypeInvalidUserInfo,
					}
					break
				}
				// check if there's no user in the system
				if controller.CurrentUser.ToUserInfo() != (api.UserInfo{}) && userInfo.Email != controller.CurrentUser.Email {
					hmc <- api.HandlerMessage{
						Type: api.TypeUserExisted,
					}
					break
				}
				// reissue the token for the existing user an return
				if userInfo.Email == controller.CurrentUser.Email {
					controller.CurrentUser = api.NewUser(&userInfo)
					log.Printf("[User] Token reissued for %v", userInfo.Name)
					controller.UserTimer.Reset(time.Second * time.Duration(*userTimeout))
					log.Println("[UserTimer] Timer resetted")
					// skip the rest and return the response with the new token
					hmc <- api.HandlerMessage{
						Type:  api.TypeUserAdded,
						Value: []interface{}{*controller.CurrentUser},
					}
					break
				}
				// register the user to the system with the new token
				controller.CurrentUser = api.NewUser(&userInfo)
				// turn on the light
				if *miioEnabled {
					cmd := exec.Command(*miiocli, "yeelight", "--ip", *miioIP, "--token", *miioToken, "on")
					cmd.Run()
				}
				// set the robot in Joint mode and go to home
				alp := armlink.Packet{}
				alp.SetExtended(armlink.ExtendedReset)
				controller.ArmLinkSerial.Send(alp.Bytes())
				// post to Slack
				if *slackAppEnabled {
					var jsonStr = []byte(fmt.Sprintf(`{"text":"<!here> User %v (%v) started using Leubot."}`, userInfo.Name, userInfo.Email))
					req, err := http.NewRequest("POST", *slackWebHookURL, bytes.NewBuffer(jsonStr))
					req.Header.Set("Content-Type", "application/json")
					r, err := (&http.Client{}).Do(req)
					if err != nil {
						panic(err)
					}
					r.Body.Close()
				}
				// start the timer
				if *userTimeout != 0 {
					controller.UserTimer.Reset(time.Second * time.Duration(*userTimeout))
					log.Printf("[UserTimer] Started for %v", userInfo.Name)
					go func() {
						for {
							select {
							case <-controller.UserActChannel: // Upon any activity, reset the timer
								log.Println("[UserTimer] Activity detected, resetting the timer")
								controller.UserTimer.Reset(time.Second * time.Duration(*userTimeout))
							case <-controller.UserTimer.C: // Inactive, logout
								log.Printf("[UserTimer] Timeout, deleting the user %v", controller.CurrentUser.Name)
								// reset CurrentRobotPose
								controller.CurrentRobotPose = &RobotPose{
									Elbow:         400,
									WristAngle:    580,
									WristRotation: 512,
									Gripper:       128,
								}
								// set the robot in sleep mode
								alp := armlink.Packet{}
								alp.SetExtended(armlink.ExtendedSleep)
								controller.ArmLinkSerial.Send(alp.Bytes())
								// turn off the light
								if *miioEnabled {
									cmd := exec.Command(*miiocli, "yeelight", "--ip", *miioIP, "--token", *miioToken, "off")
									cmd.Run()
								}
								// post to Slack
								if *slackAppEnabled {
									var jsonStr = []byte(fmt.Sprintf(`{"text":"<!here> User %v (%v) was inactive for %v seconds, releasing Leubot."}`, controller.CurrentUser.Name, controller.CurrentUser.Email, *userTimeout))
									req, err := http.NewRequest("POST", *slackWebHookURL, bytes.NewBuffer(jsonStr))
									req.Header.Set("Content-Type", "application/json")
									r, err := (&http.Client{}).Do(req)
									if err != nil {
										panic(err)
									}
									r.Body.Close()
								}
								// delete the current user; assign an empty User
								controller.CurrentUser = &api.User{}
								// exiting timer channel listener
								return
							case <-controller.UserTimerFinish:
								log.Println("[UserTimer] User deleted, terminating the timer")
								return
							}
						}
					}()
				}

				hmc <- api.HandlerMessage{
					Type:  api.TypeUserAdded,
					Value: []interface{}{*controller.CurrentUser},
				}
			case api.TypeGetUser:
				hmc <- api.HandlerMessage{
					Type:  api.TypeCurrentUser,
					Value: []interface{}{controller.CurrentUser.ToUserInfo()},
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
				if token != controller.CurrentUser.Token && token != *masterToken {
					hmc <- api.HandlerMessage{
						Type: api.TypeUserNotFound,
					}
					break
				}
				// stop the timer
				if *userTimeout != 0 {
					controller.UserTimer.Stop()
					controller.UserTimerFinish <- true
				}
				// reset CurrentRobotPose
				controller.CurrentRobotPose = &RobotPose{
					Elbow:         400,
					WristAngle:    580,
					WristRotation: 512,
					Gripper:       128,
				}
				// set the robot in sleep mode
				alp := armlink.Packet{}
				alp.SetExtended(armlink.ExtendedSleep)
				controller.ArmLinkSerial.Send(alp.Bytes())
				// turn off the light
				if *miioEnabled {
					cmd := exec.Command(*miiocli, "yeelight", "--ip", *miioIP, "--token", *miioToken, "off")
					cmd.Run()
				}
				// post to Slack
				if *slackAppEnabled {
					var jsonStr = []byte(fmt.Sprintf(`{"text":"<!here> User %v (%v) stopped using Leubot."}`, controller.CurrentUser.Name, controller.CurrentUser.Email))
					req, err := http.NewRequest("POST", *slackWebHookURL, bytes.NewBuffer(jsonStr))
					req.Header.Set("Content-Type", "application/json")
					r, err := (&http.Client{}).Do(req)
					if err != nil {
						panic(err)
					}
					r.Body.Close()
				}
				// delete the current user; assign an empty User
				controller.CurrentUser = &api.User{}

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
				if robotCommand.Token != controller.CurrentUser.Token && robotCommand.Token != *masterToken {
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
				// ack the timer
				if *userTimeout != 0 {
					controller.UserActChannel <- true
				}
				// set the value to CurrentRobotPose
				controller.CurrentRobotPose.Elbow = robotCommand.Value
				// perform the move
				controller.ArmLinkSerial.Send(controller.CurrentRobotPose.BuildPacket().Bytes())

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
				if robotCommand.Token != controller.CurrentUser.Token && robotCommand.Token != *masterToken {
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
				// ack the timer
				if *userTimeout != 0 {
					controller.UserActChannel <- true
				}
				// set the value to CurrentRobotPose
				controller.CurrentRobotPose.WristAngle = robotCommand.Value
				// perform the move
				controller.ArmLinkSerial.Send(controller.CurrentRobotPose.BuildPacket().Bytes())

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
				if robotCommand.Token != controller.CurrentUser.Token && robotCommand.Token != *masterToken {
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
				// ack the timer
				if *userTimeout != 0 {
					controller.UserActChannel <- true
				}
				// set the value to CurrentRobotPose
				controller.CurrentRobotPose.WristRotation = robotCommand.Value
				// perform the move
				controller.ArmLinkSerial.Send(controller.CurrentRobotPose.BuildPacket().Bytes())

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
				if robotCommand.Token != controller.CurrentUser.Token && robotCommand.Token != *masterToken {
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
				// ack the timer
				if *userTimeout != 0 {
					controller.UserActChannel <- true
				}
				// set the value to CurrentRobotPose
				controller.CurrentRobotPose.Gripper = robotCommand.Value
				// perform the move
				controller.ArmLinkSerial.Send(controller.CurrentRobotPose.BuildPacket().Bytes())

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
				if robotCommand.Token != controller.CurrentUser.Token && robotCommand.Token != *masterToken {
					hmc <- api.HandlerMessage{
						Type: api.TypeInvalidToken,
					}
					break
				}
				// ack the timer
				if *userTimeout != 0 {
					controller.UserActChannel <- true
				}
				// reset CurrentRobotPose
				controller.CurrentRobotPose = &RobotPose{
					Elbow:         400,
					WristAngle:    580,
					WristRotation: 512,
					Gripper:       128,
				}
				// perform the reset
				alp := armlink.Packet{}
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
	als := armlink.NewSerial()
	defer als.Close()

	// create the controller with the serial
	controller := NewController(als)
	defer controller.Shutdown()

	log.Printf("Server started")
	router := api.NewRouter(controller.HandlerChannel)
	log.Fatal(http.ListenAndServe(":6789", router))
}
