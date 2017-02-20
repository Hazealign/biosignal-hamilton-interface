package main

import (
	"os"
	"time"

	"biosignal-hamilton-interface/packet"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/tarm/serial"
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
			log.Level = logrus.InfoLevel
		} else {
			log.Level = logrus.ErrorLevel
		}
	}

	config := &serial.Config{
		Name:     Options.Port,
		Baud:     9600,
		Parity:   serial.ParityEven,
		StopBits: serial.Stop2,
	}

	serial := OpenPort(config)

	serial.Write(packet.IncomePacket{
		Identifier: 0x56,
	}.ToBytes())

	result, err := ReadFromSerial(serial)
	if err != nil {
		log.Errorln("에러가 발생했습니다.")
		log.Errorln(err)
		os.Exit(1)
	}

	pkt, err := packet.ParseOutcomePacket(result)
	if err != nil {
		log.Errorln("에러가 발생했습니다.")
		log.Errorln(err)
		os.Exit(1)
	}

	log.Debug("Data Input")
	log.Debug(result)
	log.Debug(pkt)

	for {
		serial.Write(packet.IncomePacket{
			Identifier: 120,
		}.ToBytes())

		result, err := ReadFromSerial(serial)
		if err != nil {
			log.Errorln("에러가 발생했습니다.")
			log.Errorln(err)
			os.Exit(1)
		}

		pkt, err := packet.ParseOutcomePacket(result)
		if err != nil {
			log.Errorln("에러가 발생했습니다.")
			log.Errorln(err)
			os.Exit(1)
		}

		log.Debug("기기에서 전송된 데이터: ")
		log.Debug(pkt)
		log.Debug(result)
	}
}

func OpenPort(config *serial.Config) (socket *serial.Port) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("3초 후 다시 시도합니다.")
			time.Sleep(3 * time.Millisecond)
			OpenPort(config)
		}
	}()

	socket, err := serial.OpenPort(config)
	if err != nil {
		log.Errorln("Port를 열지 못했습니다.")
		log.Errorln(err)
		panic(err)
	}

	return socket
}

func ReadFromSerial(serial *serial.Port) (buf []byte, err error) {
	tmp_buffer := make([]byte, 1024)
	n, err := serial.Read(tmp_buffer)
	if err != nil {
		return []byte{}, err
	}

	return tmp_buffer[:n], nil
}
