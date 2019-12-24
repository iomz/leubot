/*
 * Leubot
 *
 * This program provides a simple API for
 * PhantomX AX-12 Reactor Robot Arm with ArmLink Serial interface
 *
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

	//"github.com/Interactions-HSG/leubot/api"
	//"github.com/Interactions-HSG/leubot/armlink"
	"github.com/badoux/checkmail"
	"github.com/iomz/leubot/api"
	"github.com/iomz/leubot/armlink"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Environmental variables
var (
	// Current Version
	version = "1.0.4"

	// app
	app = kingpin.
		New("leubot", "Provide a Web API for the PhantomX AX-12 Reactor Robot Arm.")

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

	slackappenabled = app.
			Flag("slackappenabled", "Enable Slack app for user previleges.").
			Default("false").
			Bool()
	slackwebhookurl = app.
			Flag("slackwebhookurl", "The webhook url for posting the json payloads.").
			Default("https://hooks.slack.com/services/...").
			String()

	userTimeout = app.
			Flag("userTimeout", "The timeout duration for users in seconds.").
			Default("900").
			Int()
)

// RobotPose stores the rotations of each joint
type RobotPose struct {
	Base          uint16
	Shoulder      uint16
	Elbow         uint16
	WristAngle    uint16
	WristRotation uint16
	Gripper       uint16
}

// BuildArmLinkPacket creates a new ArmLinkPacket
func (rp *RobotPose) BuildArmLinkPacket() *armlink.ArmLinkPacket {
	return armlink.NewArmLinkPacket(rp.Base, rp.Shoulder, rp.Elbow, rp.WristAngle, rp.WristRotation, rp.Gripper, 128, 0, 0)
}

// String returns a string rep for the rp
func (rp *RobotPose) String() string {
	return fmt.Sprintf("Base: %v, Shoulder: %v, Elbow: %v, WristAngle: %v, WristRotation: %v, Gripper: %v", rp.Base, rp.Shoulder, rp.Elbow, rp.WristAngle, rp.WristRotation, rp.Gripper)
}

// Controller is the main thread for this API provider
type Controller struct {
	ArmLinkSerial     *armlink.ArmLinkSerial
	CurrentRobotPose  *RobotPose
	CurrentUser       *api.User
	HandlerChannel    chan api.HandlerMessage
	LastArmLinkPacket *armlink.ArmLinkPacket
	UserActChannel    chan bool
	UserTimer         *time.Timer
	UserTimerFinish   chan bool
}

// Shutdown processes the graceful termination of the program
func (controller *Controller) Shutdown() {
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

// NewController creates a new instance of Controller
func NewController(als *armlink.ArmLinkSerial) *Controller {
	hmc := make(chan api.HandlerMessage)
	controller := Controller{
		ArmLinkSerial: als,
		CurrentRobotPose: &RobotPose{
			Base:          512,
			Shoulder:      512,
			Elbow:         400,
			WristAngle:    580,
			WristRotation: 512,
			Gripper:       128,
		},
		CurrentUser:       &api.User{},
		HandlerChannel:    hmc,
		LastArmLinkPacket: &armlink.ArmLinkPacket{},
		UserActChannel:    make(chan bool),
		UserTimer:         time.NewTimer(time.Second * 10),
		UserTimerFinish:   make(chan bool),
	}
	controller.UserTimer.Stop()

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

			log.Printf("[CurrentRobotPose] %v", controller.CurrentRobotPose.String())

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
				if *miioenabled {
					cmd := exec.Command(*miiocli, "yeelight", "--ip", *miioip, "--token", *miiotoken, "on")
					cmd.Run()
				}
				// set the robot in Joint mode and go to home
				alp := armlink.ArmLinkPacket{}
				alp.SetExtended(armlink.ExtendedReset)
				controller.ArmLinkSerial.Send(alp.Bytes())
				// post to Slack
				if *slackappenabled {
					var jsonStr = []byte(fmt.Sprintf(`{"text":"<!here> User %v (%v) started using Leubot."}`, userInfo.Name, userInfo.Email))
					req, err := http.NewRequest("POST", *slackwebhookurl, bytes.NewBuffer(jsonStr))
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
									Base:          512,
									Shoulder:      512,
									Elbow:         400,
									WristAngle:    580,
									WristRotation: 512,
									Gripper:       128,
								}
								// set the robot in sleep mode
								alp := armlink.ArmLinkPacket{}
								alp.SetExtended(armlink.ExtendedSleep)
								controller.ArmLinkSerial.Send(alp.Bytes())
								// turn off the light
								if *miioenabled {
									cmd := exec.Command(*miiocli, "yeelight", "--ip", *miioip, "--token", *miiotoken, "off")
									cmd.Run()
								}
								// post to Slack
								if *slackappenabled {
									var jsonStr = []byte(fmt.Sprintf(`{"text":"<!here> User %v (%v) was inactive for %v seconds, releasing Leubot."}`, controller.CurrentUser.Name, controller.CurrentUser.Email, *userTimeout))
									req, err := http.NewRequest("POST", *slackwebhookurl, bytes.NewBuffer(jsonStr))
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
				} // End if *userTimeout != 0

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
				if token != controller.CurrentUser.Token && token != *mastertoken {
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
					Base:          512,
					Shoulder:      512,
					Elbow:         400,
					WristAngle:    580,
					WristRotation: 512,
					Gripper:       128,
				}
				// set the robot in sleep mode
				alp := armlink.ArmLinkPacket{}
				alp.SetExtended(armlink.ExtendedSleep)
				controller.ArmLinkSerial.Send(alp.Bytes())
				// turn off the light
				if *miioenabled {
					cmd := exec.Command(*miiocli, "yeelight", "--ip", *miioip, "--token", *miiotoken, "off")
					cmd.Run()
				}
				// post to Slack
				if *slackappenabled {
					var jsonStr = []byte(fmt.Sprintf(`{"text":"<!here> User %v (%v) stopped using Leubot."}`, controller.CurrentUser.Name, controller.CurrentUser.Email))
					req, err := http.NewRequest("POST", *slackwebhookurl, bytes.NewBuffer(jsonStr))
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
			case api.TypePutBase:
				// receive the robotCommand
				robotCommand, ok := msg.Value[0].(api.RobotCommand)
				if !ok {
					hmc <- api.HandlerMessage{
						Type: api.TypeSomethingWentWrong,
					}
					break
				}
				// check if the token is valid
				if robotCommand.Token != controller.CurrentUser.Token && robotCommand.Token != *mastertoken {
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
				fmt.Printf("Base: %v", controller.CurrentRobotPose.Base)
				controller.CurrentRobotPose.Base = robotCommand.Value
				fmt.Printf("Base: %v", controller.CurrentRobotPose.Base)
				// perform the move
				alp := controller.CurrentRobotPose.BuildArmLinkPacket()
				controller.ArmLinkSerial.Send(alp.Bytes())
				log.Printf("[ArmLinkPacket] %v", alp.String())

				hmc <- api.HandlerMessage{
					Type: api.TypeActionPerformed,
				}
			case api.TypePutShoulder:
				// receive the robotCommand
				robotCommand, ok := msg.Value[0].(api.RobotCommand)
				if !ok {
					hmc <- api.HandlerMessage{
						Type: api.TypeSomethingWentWrong,
					}
					break
				}
				// check if the token is valid
				if robotCommand.Token != controller.CurrentUser.Token && robotCommand.Token != *mastertoken {
					hmc <- api.HandlerMessage{
						Type: api.TypeInvalidToken,
					}
					break
				}
				// check the value is valid
				if robotCommand.Value < 205 || 810 < robotCommand.Value {
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
				controller.CurrentRobotPose.Shoulder = robotCommand.Value
				// perform the move
				alp := controller.CurrentRobotPose.BuildArmLinkPacket()
				controller.ArmLinkSerial.Send(alp.Bytes())
				log.Printf("[ArmLinkPacket] %v", alp.String())

				hmc <- api.HandlerMessage{
					Type: api.TypeActionPerformed,
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
				if robotCommand.Token != controller.CurrentUser.Token && robotCommand.Token != *mastertoken {
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
				alp := controller.CurrentRobotPose.BuildArmLinkPacket()
				controller.ArmLinkSerial.Send(alp.Bytes())
				log.Printf("[ArmLinkPacket] %v", alp.String())

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
				if robotCommand.Token != controller.CurrentUser.Token && robotCommand.Token != *mastertoken {
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
				alp := controller.CurrentRobotPose.BuildArmLinkPacket()
				controller.ArmLinkSerial.Send(alp.Bytes())
				log.Printf("[ArmLinkPacket] %v", alp.String())

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
				if robotCommand.Token != controller.CurrentUser.Token && robotCommand.Token != *mastertoken {
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
				alp := controller.CurrentRobotPose.BuildArmLinkPacket()
				controller.ArmLinkSerial.Send(alp.Bytes())
				log.Printf("[ArmLinkPacket] %v", alp.String())

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
				if robotCommand.Token != controller.CurrentUser.Token && robotCommand.Token != *mastertoken {
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
				alp := controller.CurrentRobotPose.BuildArmLinkPacket()
				controller.ArmLinkSerial.Send(alp.Bytes())
				log.Printf("[ArmLinkPacket] %v", alp.String())

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
				if robotCommand.Token != controller.CurrentUser.Token && robotCommand.Token != *mastertoken {
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
					Base:          512,
					Shoulder:      512,
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
