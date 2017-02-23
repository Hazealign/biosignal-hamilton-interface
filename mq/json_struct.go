package mq

import (
	"encoding/json"
	"time"

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
		TIMESTAMP: d.TIMESTAMP.Format("Mon Jan _2"),
	})
}

func (d *QueueModel) SendToNSQ(str string) error {
	config := nsq.NewConfig()
	producer, _ := nsq.NewProducer(str, config)

	err := producer.Publish("Biosignal", []byte(d.MarshalJSON()))
	return err
}
