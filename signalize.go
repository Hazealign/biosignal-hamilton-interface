package main

import (
	"errors"
	"os"
	"time"

	"biosignal-hamilton-interface/packet"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"go.bug.st/serial.v1"
)

var Options struct {
	Debug bool   `short:"d" long:"debug" description:"Enable Debug Mode." optional:"true"`
	Port  string `short:"p" long:"port" description:"Port which connected with Device" required:"true"`
}

var log = logrus.New()

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

	ser.Write(packet.IncomePacket{
		Identifier: 0x56,
	}.ToBytes())

	res, err := ReadFromSerial(ser)
	if err != nil {
		log.Errorln("에러가 발생했습니다.")
		log.Errorln(err)
		os.Exit(1)
	}

	pkt, err := packet.ParseOutcomePacket(res)
	if err != nil {
		log.Errorln(res)
		log.Errorln("에러가 발생했습니다.")
		log.Errorln(err)
		os.Exit(1)
	}

	log.Debug("Data Input")
	log.Debug(res)
	log.Debug(pkt)

	for {
		ser.Write(packet.IncomePacket{
			Identifier: 120,
		}.ToBytes())

		result, err := ReadFromSerial(ser)
		if err != nil {
			log.Errorln("에러가 발생했습니다.")
			log.Errorln(err)
			os.Exit(1)
		}

		pkt, err := packet.ParseOutcomePacket(result)
		if err != nil {
			log.Errorln(pkt)
			log.Errorln("에러가 발생했습니다.")
			log.Errorln(err)
			os.Exit(1)
		}

		log.Debug("기기에서 전송된 데이터: ")
		log.Debug(pkt)
		log.Debug(result)
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
	tmp_buffer := make([]byte, 1024)
	n, err := serial.Read(tmp_buffer)
	if n == 0 {
		return tmp_buffer[:n], errors.New("Timeout or EOF")
	}

	return tmp_buffer[:n], err
}
