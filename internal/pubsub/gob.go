package pubsub

import (
	"bytes"
	"encoding/gob"
)

func EncodeGob[T any](v T) ([]byte, error) {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func DecodeGob[T any](data []byte) (T, error) {
	var v T
	buff := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buff)
	err := decoder.Decode(&v)
	if err != nil {
		return v, err
	}
	return v, nil
}
