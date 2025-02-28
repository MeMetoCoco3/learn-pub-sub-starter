package pubsub

import (
	"bytes"
	"encoding/gob"
	"time"
)

func encode(gl GameLog) ([]byte, error) {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(gl)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func decode(data []byte) (GameLog, error) {
	var g GameLog
	buff := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buff)
	err := decoder.Decode(&g)
	if err != nil {
		return g, err
	}
	return g, nil
}

// don't touch below this line

type GameLog struct {
	CurrentTime time.Time
	Message     string
	Username    string
}
