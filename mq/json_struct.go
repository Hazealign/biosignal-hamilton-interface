package mq

import (
	"encoding/json"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
)

type QueueModel struct {
	TIMESTAMP      time.Time
	TYPE           string
	KEY            string
	PORT           string
	HOST           string
	VALUE_UNIT     string
	UDID           string
	DEVICE         string
	NUMERIC_VALUE  float64
	WAVEFORM_VALUE []int
	PATIENT_ID     string
}

func (d *QueueModel) MarshalJSON() ([]byte, error) {
	d.DEVICE = "Hamilton"
	d.PATIENT_ID = "TEST_ID"

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
	d.DEVICE = "Hamilton"
	d.PATIENT_ID = "TEST_ID"

	config := nsq.NewConfig()
	producer, _ := nsq.NewProducer(str, config)

	jsonVal, _ := d.MarshalJSON()
	logrus.Println(string(jsonVal))
	err := producer.Publish("Biosignal", jsonVal)
	return err
}
