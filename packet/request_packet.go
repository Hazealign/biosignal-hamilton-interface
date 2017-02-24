package packet

import "errors"

type RequestPacket struct {
	Identifier byte
}

func (packet RequestPacket) ToBytes() (result []byte) {
	retVal := []byte{0x02, packet.Identifier, 0x03, 0x0D}
	return retVal
}

func (packet RequestPacket) GetType() (result string) {
	return TypeIntString[int(packet.Identifier)]
}

func ParseIncomePacket(raw []byte) (result RequestPacket, err error) {
	if len(raw) != 4 {
		return RequestPacket{}, errors.New("Invalid Packet Length")
	}

	return RequestPacket{
		Identifier: raw[1],
	}, nil
}
