package mq

import (
	"encoding/json"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
)

type QueueModel struct {
	TIMESTAMP  time.Time
	TYPE       string
	PORT       string
	HOST       string
	UNIT       string
	UDID       string
	P_PATIENT  []int
	P_OPTIONAL []int
	FLOW       []int
	VOLUME     []int
	PCO2       []int
}

func (d *QueueModel) MarshalJSON() ([]byte, error) {
	type Alias QueueModel
	return json.Marshal(&struct {
		*Alias
		TIMESTAMP string `json:"TIMESTAMP"`
	}{
		Alias:     (*Alias)(d),
		TIMESTAMP: d.TIMESTAMP.Format(time.RFC3339),
	})
}

func SendToNSQ(d QueueModel, str string) error {
	config := nsq.NewConfig()
	producer, _ := nsq.NewProducer(str, config)

	jsonVal, _ := d.MarshalJSON()
	logrus.Println(string(jsonVal))
	err := producer.Publish("Biosignal", jsonVal)
	return err
}
