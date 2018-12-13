package api

// HandlerMessageType indicates the type of the request handler message
type HandlerMessageType int

const (
	// TypeAddUser requests adding a user
	TypeAddUser HandlerMessageType = iota
	// TypeUserAdded indicates the user is added
	TypeUserAdded
	// TypeUserExisted indicates there's a user already
	TypeUserExisted
	// TypeInvalidUserInfo indicates the request includes invalid information
	TypeInvalidUserInfo
	// TypeDeleteUser request deleting the user
	TypeDeleteUser
	// TypeUserDeleted indicates the user is deleted
	TypeUserDeleted
	// TypeUserNotFound indicates there's no such user
	TypeUserNotFound
	// TypeGetUser requests the current user's info
	TypeGetUser
	// TypeCurrentUser indicates returning the current user's info
	TypeCurrentUser
	// TypePutElbow requests the elbow to move
	TypePutElbow
	// TypePutWristAngle requests the wrist angle to move
	TypePutWristAngle
	// TypePutWristRotation requests the wrist rotaton to move
	TypePutWristRotation
	// TypePutGripper requests the gripper to move
	TypePutGripper
	// TypePutReset requests resetting the robot
	TypePutReset
	// TypeActionPerformed indicates the requested action is performed
	TypeActionPerformed
	// TypeInvalidToken indicates the provided token is invalid
	TypeInvalidToken
	// TypeInvalidCommand indicates the provided command is invalid
	TypeInvalidCommand
	// TypeSomethingWentWrong indicates there was something wrong...
	TypeSomethingWentWrong
)

// HandlerMessage is the message for request handling and control
type HandlerMessage struct {
	Type  HandlerMessageType
	Value []interface{}
}
