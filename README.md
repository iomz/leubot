# leubot
A middleware provides a web API for control PhantomX AX-12 Reactor Robot Arm

# Supported devices
## PhantomX AX-12 Reactor Robot Arm
https://www.trossenrobotics.com/p/phantomx-ax-12-reactor-robot-arm.aspx

# Getting Started

1. Follow the link and install the `ArmLinkSerial` firmware on the robot: https://learn.trossenrobotics.com/36-demo-code/137-interbotix-arm-link-software.html#firmware
2. Add the user to `dialout` group so to be able to acesss the serial device.
3. Connect the robot to your device with FTDI-USB cable.
4. 
```
% go get github.com/Interactions-HSG/ax12ctrl
% cd `go list -f '{{.Dir}}' github.com/Interactions-HSG/ax12ctrl`
% go run main.go
```

# Reactor Arm Backhoe/Joint Positioning Limits
These values are taken from: https://learn.trossenrobotics.com/arbotix/arbotix-communication-controllers/31-arm-link-reference.html

| Parameter          | Lower Limit | Upper Limit | Default |
| ------------------ | ----------- | ----------- | ------- |
| Base Joint         | 0           | 1023        | 512     |
| Shoulder Joint     | 205         | 810         | 512     |
| Elbow Joint        | 210         | 900         | 512     |
| Wrist Angle Joint  | 200         | 830         | 512     |
| Wrist Rotate Joint | 0           | 1023        | 512     |
| Gripper Joint      | 0           | 512         | 512     |
| Delta              | 0           | 254         | 128     |
| Button             | 0           | 127         | 0       |
| Extended           | 0           | 254         | 0       |

# API Spec
See the documentation on SwaggerHub: https://app.swaggerhub.com/apis/iomz/leubot

# Setting Instruction
TODO: overall diagram and system desgin

# Synopsis


# License
See the `LICENSE` file
