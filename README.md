# ax12ctrl
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

# API Spec
TODO: Swagger

# Setting Instruction
TODO: overall diagram and system desgin

# Synopsis


# License
See the `LICENSE` file
