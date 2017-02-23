package packet

import (
	"errors"
	"math"
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
	PPatientLow      byte
	PPatientHigh     byte
	POptionalLow     byte
	POptionalHigh    byte
	FlowLow          byte
	FlowHigh         byte
	VolumeLow        byte
	VolumeHigh       byte
	PCO2Low          byte
	PCO2High         byte
}

func (packet OutcomePacket) ToBytes() (result []byte) {
	var r_error = []byte{
		0x02, 0x52, 0x45, 0x52, 0x52, 0x4F, 0x52, 0x03, 0x0D,
	}

	switch packet.ResponseType {
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
		var retVal = append([]byte{0x02}, packet.DeviceIdentifier...)
		retVal = append(retVal, packet.Values...)
		retVal = append(retVal, []byte{0x03, 0x0D}...)
		return retVal
	case RESP_TYPE_B_FORMAT_3:
		var retVal = append([]byte{0x02}, packet.DeviceIdentifier...)
		retVal = append(retVal, packet.Values...)
		retVal = append(retVal, []byte{0x03, 0x0D}...)
		return retVal
	case RESP_TYPE_C_34:
		return []byte{
			0x02, packet.Identifier, packet.VentilatorStatus,
			packet.PPatientLow, packet.PPatientHigh,
			packet.FlowLow, packet.FlowHigh,
			packet.VolumeLow, packet.VolumeHigh,
			packet.PCO2Low, packet.PCO2High,
			0x03, 0x0D,
		}
	case RESP_TYPE_C_120:
		return []byte{
			0x02, packet.Identifier, packet.VentilatorStatus,
			packet.PPatientLow, packet.PPatientHigh,
			packet.POptionalLow, packet.POptionalHigh,
			packet.FlowLow, packet.FlowHigh,
			packet.VolumeLow, packet.VolumeHigh,
			0x03, 0x0D,
		}
	}

	return r_error
}

func ParseOutcomePacket(raw []byte) (result OutcomePacket, err error) {
	var r_error = []byte{
		0x02, 0x52, 0x45, 0x52, 0x52, 0x4F, 0x52, 0x03, 0x0D,
	}

	// 0. Check packet is rerror
	if reflect.DeepEqual(r_error, raw) {
		return OutcomePacket{
			ResponseType: RESP_TYPE_RERROR,
		}, nil
	}

	// 1. check raw packet's length
	if len(raw) == 8 {
		return OutcomePacket{
			ResponseType:     RESP_TYPE_B_FORMAT_3,
			DeviceIdentifier: []byte{raw[1]},
			Values:           raw[2:6],
		}, nil
	}

	// 2. check parameter identifier
	if len(raw) == 13 {
		if int(raw[1]) == 34 {
			return OutcomePacket{
				ResponseType:     RESP_TYPE_C_34,
				Identifier:       raw[1],
				VentilatorStatus: raw[2],
				PPatientLow:      raw[3],
				PPatientHigh:     raw[4],
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
				PPatientLow:      raw[3],
				PPatientHigh:     raw[4],
				POptionalLow:     raw[5],
				POptionalHigh:    raw[6],
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

		if identifier == int(0x41) || identifier == int(0x56) || identifier == int(0x42) ||
			identifier == int(0x52) || identifier == int(0x43) {
			flag = RESP_TYPE_B_FORMAT_2
		} else if identifier >= 30 && identifier <= 33 {
			flag = RESP_TYPE_A
		} else if identifier >= 35 && identifier <= 119 {
			flag = RESP_TYPE_A
		} else if identifier >= 121 && identifier <= 123 {
			flag = RESP_TYPE_A
		} else if identifier >= 124 && identifier <= 127 {
			flag = RESP_TYPE_B_FORMAT_1
		} else {
			return OutcomePacket{}, errors.New("Invalid Outcome Packet!")
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
				Values:           raw[3:7],
			}, nil
		}
	}

	return OutcomePacket{}, errors.New("Invalid Outcome Packet!")
}

func ConvertBitWaveform(high byte, low byte) []uint8 {
	var retVal = []uint8{}

	for i := uint(0); i < 6; i++ {
		retVal = append(retVal, low&(1<<i)>>i)
	}

	for i := uint(0); i < 6; i++ {
		retVal = append(retVal, high&(1<<i)>>i)
	}

	// Reverse Array
	for i, j := 0, len(retVal)-1; i < j; i, j = i+1, j-1 {
		retVal[i], retVal[j] = retVal[j], retVal[i]
	}

	return retVal
}

func BitArrayToInteger(bitArray []uint8) int {
	var retVal = int(0)

	for i := len(bitArray) - 1; i >= 0; i-- {
		if bitArray[i] == 1 {
			retVal = retVal + int(math.Pow(2, float64(len(bitArray)-i-1)))
		}
	}

	return retVal
}
