package packet

import (
	"errors"
	"reflect"
)

type OutcomePacket struct {
	// Important Values
	ResponseType     int
	Identifier       byte
	DeviceIdentifier []byte
	Values           []byte

	// Optional Values
	VentilatorStatus byte
	pPatientLow      byte
	pPatientHigh     byte
	pOptionalLow     byte
	pOptionalHigh    byte
	FlowLow          byte
	FlowHigh         byte
	VolumeLow        byte
	VolumeHigh       byte
	PCO2Low          byte
	PCO2High         byte
}

func (packet OutcomePacket) ToBytes() (result []byte) {
	var rerror = []byte{
		0x02, 0x52, 0x45, 0x52, 0x52, 0x4F, 0x52, 0x03, 0x0D,
	}

	switch packet.Identifier {
	case RESP_TYPE_A:
		var retVal = []byte{0x02, packet.Identifier}
		retVal = append(retVal, packet.Values...)
		retVal = append(retVal, []byte{0x03, 0x0D}...)
		return retVal
	case RESP_TYPE_B_FORMAT_1:
		var retVal = []byte{
			0x02, packet.Identifier, packet.DeviceIdentifier[0],
		}
		retVal = append(retVal, packet.Values...)
		retVal = append(retVal, []byte{0x03, 0x0D}...)
		return retVal
	case RESP_TYPE_B_FORMAT_2:
	case RESP_TYPE_B_FORMAT_3:
		var retVal = append([]byte{0x02}, packet.DeviceIdentifier...)
		retVal = append(retVal, packet.Values...)
		retVal = append(retVal, []byte{0x03, 0x0D}...)
		return retVal
	case RESP_TYPE_C_34:
		return []byte{
			0x02, packet.Identifier, packet.VentilatorStatus,
			packet.pPatientLow, packet.pPatientHigh,
			packet.FlowLow, packet.FlowHigh,
			packet.VolumeLow, packet.VolumeHigh,
			packet.PCO2Low, packet.PCO2High,
			0x03, 0x0D,
		}
	case RESP_TYPE_C_120:
		return []byte{
			0x02, packet.Identifier, packet.VentilatorStatus,
			packet.pPatientLow, packet.pPatientHigh,
			packet.pOptionalLow, packet.pOptionalHigh,
			packet.FlowLow, packet.FlowHigh,
			packet.VolumeLow, packet.VolumeHigh,
			0x03, 0x0D,
		}
	}

	return rerror
}

func ParseOutcomePacket(raw []byte) (result OutcomePacket, err error) {
	var rerror = []byte{
		0x02, 0x52, 0x45, 0x52, 0x52, 0x4F, 0x52, 0x03, 0x0D,
	}

	// 0. Check packet is rerror
	if reflect.DeepEqual(rerror, raw) {
		return OutcomePacket{
			ResponseType: RESP_TYPE_RERROR,
		}, nil
	}

	// 1. check raw packet's length
	if len(raw) == 8 {
		return OutcomePacket{
			ResponseType:     RESP_TYPE_B_FORMAT_3,
			DeviceIdentifier: []byte{raw[1]},
			Values:           raw[2:5],
		}, nil
	}

	// 2. check parameter identifier
	if len(raw) == 13 {
		if int(raw[1]) == 34 {
			return OutcomePacket{
				ResponseType:     RESP_TYPE_C_34,
				Identifier:       raw[1],
				VentilatorStatus: raw[2],
				pPatientLow:      raw[3],
				pPatientHigh:     raw[4],
				FlowLow:          raw[5],
				FlowHigh:         raw[6],
				VolumeLow:        raw[7],
				VolumeHigh:       raw[8],
				PCO2Low:          raw[9],
				PCO2High:         raw[10],
			}, nil
		} else {
			return OutcomePacket{
				ResponseType:     RESP_TYPE_C_120,
				Identifier:       raw[1],
				VentilatorStatus: raw[2],
				pPatientLow:      raw[3],
				pPatientHigh:     raw[4],
				pOptionalLow:     raw[5],
				pOptionalHigh:    raw[6],
				FlowLow:          raw[7],
				FlowHigh:         raw[8],
				VolumeLow:        raw[9],
				VolumeHigh:       raw[10],
			}, nil
		}
	}

	var flag = 0
	if len(raw) == 9 {
		var identifier = int(raw[1])
		if identifier >= 30 && identifier <= 33 {
			flag = RESP_TYPE_A
		} else if identifier >= 35 && identifier <= 119 {
			flag = RESP_TYPE_A
		} else if identifier >= 121 && identifier <= 123 {
			flag = RESP_TYPE_A
		} else if identifier >= 124 && identifier <= 127 {
			flag = RESP_TYPE_B_FORMAT_1
		} else {
			flag = RESP_TYPE_B_FORMAT_2
		}

		switch flag {
		case RESP_TYPE_A:
			return OutcomePacket{
				ResponseType: RESP_TYPE_A,
				Identifier:   raw[1],
				Values:       raw[2:7],
			}, nil
		case RESP_TYPE_B_FORMAT_1:
			return OutcomePacket{
				ResponseType:     RESP_TYPE_B_FORMAT_1,
				Identifier:       raw[1],
				DeviceIdentifier: []byte{raw[2]},
				Values:           raw[3:7],
			}, nil
		case RESP_TYPE_B_FORMAT_2:
			return OutcomePacket{
				ResponseType:     RESP_TYPE_B_FORMAT_2,
				DeviceIdentifier: raw[1:3],
				Values:           raw[2:7],
			}, nil
		}
	}

	return OutcomePacket{}, errors.New("Invalid Outcome Packet!")
}
