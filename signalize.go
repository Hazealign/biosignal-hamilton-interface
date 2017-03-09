package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"strconv"
	"time"

	"biosignal-hamilton-interface/mq"
	"biosignal-hamilton-interface/packet"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"go.bug.st/serial.v1"
)

var Options struct {
	Debug      bool   `short:"d" long:"debug" description:"Enable Debug Mode." optional:"true"`
	Port       string `short:"p" long:"port" description:"Port which connected with Device" required:"true"`
	NsqAddress string `short:"a" long:"address" description:"Address of NSQ Server" required:"true"`
}

var log = logrus.New()

// 가져와야할 Numeric Values
var list = []byte{
	40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 87, 104, 105, 106, 107, 108, 110, 111,
	35, 36, 37, 38, 39, 60, 61, 62, 63, 64, 65, 66, 67, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79,
	103, 113, 114, 115, 116, 117, 118, 120, 121, 122,
}

func main() {
	log.Formatter = new(logrus.TextFormatter)
	log.Out = os.Stdout

	if _, err := flags.ParseArgs(&Options, os.Args); err != nil {
		log.Errorln("포트 번호가 명시되지 않았습니다")
		log.Errorln(err)
		os.Exit(1)
	} else {
		if Options.Debug {
			log.Level = logrus.DebugLevel
		} else {
			log.Level = logrus.ErrorLevel
		}
	}

	config := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.EvenParity,
		StopBits: serial.TwoStopBits,
	}

	ser := OpenPort(Options.Port, config)

	ser.Write(packet.RequestPacket{
		Identifier: 0x56,
	}.ToBytes())

	res, err := ReadFromSerial(ser)
	if err != nil {
		log.Errorln("에러가 발생했습니다.")
		log.Errorln(err)
		os.Exit(1)
	}

	pkt, err := packet.ParseResponsePacket(res)
	if err != nil {
		log.Errorln(res)
		log.Errorln("에러가 발생했습니다.")
		log.Errorln(err)
		os.Exit(1)
	}

	log.Debug("Data Input")
	log.Debug(res)
	log.Debug(pkt)
	crypt := sha1.New()
	crypt.Write(pkt.Values)
	result := hex.EncodeToString(crypt.Sum(nil))
	var udid = string(result)
	var host = GetHostAddress() + ":" + Options.Port

	var index = 0
	for {
		ser.Write(packet.RequestPacket{
			Identifier: 120,
		}.ToBytes())

		ReceiveWaveforms(ser, udid, host)

		ser.Write(packet.RequestPacket{
			Identifier: list[index%len(list)],
		}.ToBytes())

		ReceiveNumerics(int(list[index%len(list)]), ser, udid, host)
		if index == len(list)*20 {
			index = 1
		} else {
			index += 1
		}
	}
}

func ReceiveWaveforms(ser serial.Port, udid string, host string) {
	result, err := ReadFromSerial(ser)
	if err != nil {
		log.Errorln("에러가 발생했습니다.")
		log.Errorln(err)
		os.Exit(1)
	}

	pkt, err := packet.ParseResponsePacket(result)
	if err != nil {
		log.Errorln(result)
		log.Errorln(pkt)
		log.Errorln("에러가 발생했습니다.")
		log.Errorln(err)
		os.Exit(1)
	}

	log.Debug("기기에서 전송된 데이터: ")
	log.Debug(pkt)
	log.Debug(result)

	err1 := mq.SendToNSQ(mq.QueueModel{
		TIMESTAMP:  time.Now(),
		KEY:        "P_PATIENT",
		TYPE:       "Waveform",
		HOST:       host,
		VALUE_UNIT: "",
		UDID:       udid,
		WAVEFORM_VALUE: []int{
			packet.BitArrayToInteger(packet.ConvertBitWaveform(pkt.PPatientHigh, pkt.PPatientLow)) - 2048,
		},
	}, Options.NsqAddress)

	err2 := mq.SendToNSQ(mq.QueueModel{
		TIMESTAMP:  time.Now(),
		KEY:        "P_OPTIONAL",
		TYPE:       "Waveform",
		HOST:       host,
		VALUE_UNIT: "",
		UDID:       udid,
		WAVEFORM_VALUE: []int{
			packet.BitArrayToInteger(packet.ConvertBitWaveform(pkt.POptionalHigh, pkt.POptionalLow)) - 2048,
		},
	}, Options.NsqAddress)

	err3 := mq.SendToNSQ(mq.QueueModel{
		TIMESTAMP:  time.Now(),
		KEY:        "FLOW",
		TYPE:       "Waveform",
		HOST:       host,
		VALUE_UNIT: "",
		UDID:       udid,
		WAVEFORM_VALUE: []int{
			packet.BitArrayToInteger(packet.ConvertBitWaveform(pkt.FlowHigh, pkt.FlowLow)) - 2048,
		},
	}, Options.NsqAddress)

	err4 := mq.SendToNSQ(mq.QueueModel{
		TIMESTAMP:  time.Now(),
		KEY:        "VOLUME",
		TYPE:       "Waveform",
		HOST:       host,
		VALUE_UNIT: "",
		UDID:       udid,
		WAVEFORM_VALUE: []int{
			packet.BitArrayToInteger(packet.ConvertBitWaveform(pkt.VolumeHigh, pkt.VolumeLow)) - 2048,
		},
	}, Options.NsqAddress)

	var errs = []error{err1, err2, err3, err4}
	for _, err := range errs {
		if err != nil {
			log.Errorln("NSQ에 보내는 중 오류가 발생하였습니다.")
			log.Errorln(err)
			panic(err)
		}
	}
}

func ReceiveNumerics(identifier int, ser serial.Port, udid string, host string) {
	result, err := ReadFromSerial(ser)
	if err != nil {
		log.Errorln("에러가 발생했습니다.")
		log.Errorln(err)
		os.Exit(1)
	}

	pkt, err := packet.ParseResponsePacket(result)
	if err != nil {
		log.Errorln(result)
		log.Errorln(pkt)
		log.Errorln("에러가 발생했습니다.")
		log.Errorln(err)
		os.Exit(1)
	}

	log.Debug("기기에서 전송된 데이터: ")
	log.Debug(pkt)
	log.Debug(result)

	var strVal = string(pkt.Values)
	floatVal, err := strconv.ParseFloat(strVal, 64)

	if err != nil {
		var model = mq.QueueModel{
			TIMESTAMP:     time.Now(),
			KEY:           packet.TypeIntString[identifier],
			TYPE:          "Numeric",
			HOST:          host,
			VALUE_UNIT:    "",
			UDID:          udid,
			NUMERIC_VALUE: floatVal,
		}

		err := mq.SendToNSQ(model, Options.NsqAddress)
		if err != nil {
			log.Errorln("NSQ에 보내는 중 오류가 발생하였습니다.")
			log.Errorln(err)
			panic(err)
		}
	}
}

func OpenPort(port string, config *serial.Mode) (socket serial.Port) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("3초 후 다시 시도합니다.")
			time.Sleep(3 * time.Second)
			OpenPort(port, config)
		}
	}()

	socket, err := serial.Open(port, config)
	if err != nil {
		log.Errorln("Port를 열지 못했습니다.")
		log.Errorln(err)
		panic(err)
	}

	return socket
}

func ReadFromSerial(serial serial.Port) (buf []byte, err error) {
	// sleep must be more than 36 millisecond
	time.Sleep(36 * time.Millisecond)
	tmp_buffer := make([]byte, 1024)
	n, err := serial.Read(tmp_buffer)
	if n == 0 {
		return tmp_buffer[:n], errors.New("Timeout or EOF")
	}

	return tmp_buffer[:n], err
}

func GetHostAddress() string {
	addresses, _ := net.InterfaceAddrs()
	for _, a := range addresses {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "ERROR"
}
