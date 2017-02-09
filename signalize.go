package main

import (
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/tarm/serial"
)

var Options struct {
	Debug bool   `short:"d" long:"debug" description:"Enable Debug Mode." optional:"true"`
	Port  string `short:"p" long:"port" description:"Port which connected with Device" required:"true"`
}

var log = logrus.New()

func init() {
	log.Formatter = new(logrus.TextFormatter)
	logrus.SetOutput(log.Writer())
}

func main() {
	if _, err := flags.ParseArgs(&Options, os.Args); err != nil {
		log.Errorln("포트 번호가 명시되지 않았습니다")
		log.Errorln(err)
		os.Exit(1)
	} else {
		if Options.Debug {
			logrus.SetLevel(logrus.InfoLevel)
		} else {
			logrus.SetLevel(logrus.ErrorLevel)
		}
	}

	config := &serial.Config{
		Name:     Options.Port,
		Baud:     9600,
		Parity:   serial.ParityEven,
		StopBits: serial.Stop2,
	}

	serial := OpenPort(config)
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
