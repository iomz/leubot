package armlink

type ArmLinkPacket struct {
	baseRotation            uint16
	shoulderRotation        uint16
	elbowRotation           uint16
	wristAngle              uint16
	wristRotation           uint16
	gripper                 uint16
	deltaByte               byte
	buttonByte              byte
	extendedInstructionByte byte
}

func NewArmLinkPacket(br, sr, er, wa, wr, g uint16, d, b, e byte) *ArmLinkPacket {
	alp := &ArmLinkPacket{
		baseRotation:            br,
		shoulderRotation:        sr,
		elbowRotation:           er,
		wristAngle:              wa,
		wristRotation:           wr,
		gripper:                 g,
		deltaByte:               d,
		buttonByte:              b,
		extendedInstructionByte: e,
	}
	return alp
}

func (alp *ArmLinkPacket) Bytes() []byte {
	baseRotationHighByte := byte(alp.baseRotation >> 8)
	baseRotationLowByte := byte(alp.baseRotation & 0x0F)
	shoulderRotationHighByte := byte(alp.shoulderRotation >> 8)
	shoulderRotationLowByte := byte(alp.shoulderRotation & 0x0F)
	elbowRotationHighByte := byte(alp.elbowRotation >> 8)
	elbowRotationLowByte := byte(alp.elbowRotation & 0x0F)
	wristAngleHighByte := byte(alp.wristAngle >> 8)
	wristAngleLowByte := byte(alp.wristAngle & 0x0F)
	wristRotationHighByte := byte(alp.wristRotation >> 8)
	wristRotationLowByte := byte(alp.wristRotation & 0x0F)
	gripperHighByte := byte(alp.gripper >> 8)
	gripperLowByte := byte(alp.gripper & 0x0F)

	checksum := ^((baseRotationHighByte +
		baseRotationLowByte +
		shoulderRotationHighByte +
		shoulderRotationLowByte +
		elbowRotationHighByte +
		elbowRotationLowByte +
		wristAngleHighByte +
		wristAngleLowByte +
		wristRotationHighByte +
		wristRotationLowByte +
		gripperHighByte +
		gripperLowByte +
		alp.deltaByte +
		alp.buttonByte +
		alp.extendedInstructionByte) % 0xff)

	return []byte{
		byte(0xff), // header
		baseRotationHighByte,
		baseRotationLowByte,
		shoulderRotationHighByte,
		shoulderRotationLowByte,
		elbowRotationHighByte,
		elbowRotationLowByte,
		wristAngleHighByte,
		wristAngleLowByte,
		wristRotationHighByte,
		wristRotationLowByte,
		gripperHighByte,
		gripperLowByte,
		alp.deltaByte,
		alp.buttonByte,
		alp.extendedInstructionByte,
		checksum,
	}
}
