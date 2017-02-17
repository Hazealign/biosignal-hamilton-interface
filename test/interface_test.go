package signalize

import (
	"testing"

	"biosignal-hamilton-interface/packet"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSignalize(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BiosignalHamiltonInterface Suite")
}

var IncomePacket = Describe("Income Packet", func() {
	It("Encoding", func() {
		var bytes = packet.IncomePacket{
			Identifier: byte(0x40),
		}.ToBytes()

		Ω(bytes).Should(Equal([]byte{
			0x02, 0x40, 0x03, 0x0D,
		}))
	})

	It("Decoding - Valid", func() {
		var bytes = []byte{0x02, 0x40, 0x03, 0x0D}

		pkt, err := packet.ParseIncomePacket(bytes)

		Ω(err).Should(BeNil())
		Ω(pkt).Should(Equal(packet.IncomePacket{
			Identifier: 0x40,
		}))
	})

	It("Decoding - Invalid", func() {
		var bytes = []byte{0x02, 0x03, 0x0D}

		pkt, err := packet.ParseIncomePacket(bytes)

		Ω(err).ShouldNot(BeNil())
		Ω(pkt.Identifier).Should(BeZero())
	})
})

var OutcomePacket = Describe("Outcome Packet Decoder", func() {
	// Check Invalid Packet
	It("Decoding - Invalid", func() {
		var bytes = []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD}

		pkt, err := packet.ParseOutcomePacket(bytes)

		Ω(err).ShouldNot(BeNil())
		Ω(pkt.ResponseType).Should(BeZero())
	})

	// Check RERROR
	It("Decoding RERRROR", func() {
		var bytes = []byte{0x02, 0x52, 0x45, 0x52, 0x52, 0x4F, 0x52, 0x03, 0x0D}

		pkt, err := packet.ParseOutcomePacket(bytes)

		Ω(err).Should(BeNil())
		Ω(pkt.ResponseType).Should(Equal(packet.RESP_TYPE_RERROR))
	})

	It("Decoding Type A", func() {
		var bytes = []byte{0x02, 30, 0x00, 0x01, 0x02, 0x03, 46, 0x03, 0x0D}

		pkt, err := packet.ParseOutcomePacket(bytes)

		Ω(err).Should(BeNil())
		Ω(pkt.ResponseType).Should(Equal(packet.RESP_TYPE_A))
		Ω(pkt.Identifier).Should(Equal(byte(30)))
		Ω(pkt.Values).Should(Equal([]byte{0x00, 0x01, 0x02, 0x03, 46}))
	})

	It("Decoding Type B - Format 1", func() {
		var bytes = []byte{0x02, 0x7C, 0x47, 0x30, 0x32, 0x30, 0x30, 0x03, 0x0D}

		pkt, err := packet.ParseOutcomePacket(bytes)

		Ω(err).Should(BeNil())
		Ω(pkt.ResponseType).Should(Equal(packet.RESP_TYPE_B_FORMAT_1))
		Ω(pkt.Identifier).Should(Equal(byte(0x7C)))
		Ω(pkt.DeviceIdentifier).Should(Equal([]byte{0x47}))
		Ω(pkt.Values).Should(Equal([]byte{0x30, 0x32, 0x30, 0x30}))
	})

	It("Decoding Type B - Format 2", func() {
		var bytes = []byte{0x02, 0x41, 0x35, 0x33, 0x34, 0x32, 0x6A, 0x03, 0x0D}

		pkt, err := packet.ParseOutcomePacket(bytes)

		Ω(err).Should(BeNil())
		Ω(pkt.ResponseType).Should(Equal(packet.RESP_TYPE_B_FORMAT_2))
		Ω(pkt.Identifier).Should(BeZero())
		Ω(pkt.DeviceIdentifier).Should(Equal([]byte{0x41, 0x35}))
		Ω(pkt.Values).Should(Equal([]byte{0x33, 0x34, 0x32, 0x6A}))
	})

	It("Decoding Type B - Format 3", func() {
		var bytes = []byte{0x02, 0x43, 0x39, 0x39, 0x39, 0x39, 0x03, 0x0D}

		pkt, err := packet.ParseOutcomePacket(bytes)

		Ω(err).Should(BeNil())
		Ω(pkt.ResponseType).Should(Equal(packet.RESP_TYPE_B_FORMAT_3))
		Ω(pkt.Identifier).Should(BeZero())
		Ω(pkt.DeviceIdentifier).Should(Equal([]byte{0x43}))
		Ω(pkt.Values).Should(Equal([]byte{0x39, 0x39, 0x39, 0x39}))
	})

	It("Encoding RERROR", func() {
		var bytes = packet.OutcomePacket{
			ResponseType: packet.RESP_TYPE_RERROR,
		}.ToBytes()

		Ω(bytes).Should(Equal([]byte{
			0x02, 0x52, 0x45, 0x52, 0x52, 0x4F, 0x52, 0x03, 0x0D,
		}))
	})

	It("Encoding Type A", func() {
		var bytes = packet.OutcomePacket{
			ResponseType: packet.RESP_TYPE_A,
			Identifier:   byte(30),
			Values:       []byte{0x00, 0x01, 0x02, 0x03, 46},
		}.ToBytes()

		Ω(bytes).Should(Equal([]byte{
			0x02, 30, 0x00, 0x01, 0x02, 0x03, 46, 0x03, 0x0D,
		}))
	})

	It("Encoding Type B - Format 1", func() {
		var bytes = packet.OutcomePacket{
			ResponseType:     packet.RESP_TYPE_B_FORMAT_1,
			Identifier:       byte(0x7C),
			DeviceIdentifier: []byte{0x47},
			Values:           []byte{0x30, 0x32, 0x30, 0x30},
		}.ToBytes()

		Ω(bytes).Should(Equal([]byte{
			0x02, 0x7C, 0x47, 0x30, 0x32, 0x30, 0x30, 0x03, 0x0D,
		}))
	})

	It("Encoding Type B - Format 2", func() {
		var bytes = packet.OutcomePacket{
			ResponseType:     packet.RESP_TYPE_B_FORMAT_2,
			DeviceIdentifier: []byte{0x41, 0x35},
			Values:           []byte{0x33, 0x34, 0x32, 0x6A},
		}.ToBytes()

		Ω(bytes).Should(Equal([]byte{
			0x02, 0x41, 0x35, 0x33, 0x34, 0x32, 0x6A, 0x03, 0x0D,
		}))
	})

	It("Encoding Type B - Format 3", func() {
		var bytes = packet.OutcomePacket{
			ResponseType:     packet.RESP_TYPE_B_FORMAT_3,
			DeviceIdentifier: []byte{0x43},
			Values:           []byte{0x39, 0x39, 0x39, 0x39},
		}.ToBytes()

		Ω(bytes).Should(Equal([]byte{
			0x02, 0x43, 0x39, 0x39, 0x39, 0x39, 0x03, 0x0D,
		}))
	})
})
