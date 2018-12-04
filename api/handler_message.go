package api

type HandlerMessageType int

const (
	TypeAddUser HandlerMessageType = iota
	TypeUserAdded
	TypeUserExisted
	TypeDeleteUser
	TypeUserDeleted
	TypeUserNotFound
	TypeGetUser
	TypeCurrentUser
	TypePutElbow
	TypePutWristAngle
	TypePutWristRotation
	TypePutGripper
	TypeResetRobot
	TypeActionPerformed
	TypeInvalidToken
	TypeInvalidCommand
	TypeSomethingWentWrong
)

type HandlerMessage struct {
	Type  HandlerMessageType
	Value []interface{}
}
