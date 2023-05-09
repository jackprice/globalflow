package db

import "encoding/json"

type DataType int
type Time int64

const (
	DataTypeString DataType = iota
	DataTypeList
)

type Data struct {
	Type        DataType `json:"type"`
	StringValue string   `json:"stringValue"`
	ListValue   []string `json:"listValue"`
	ExpiresAt   Time     `json:"expiresAt"`
}

func (data Data) Encode() ([]byte, error) {
	return json.Marshal(data)
}

func (data *Data) Decode(encoded []byte) error {
	return json.Unmarshal(encoded, data)
}
