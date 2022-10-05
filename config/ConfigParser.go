package config

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

type configParser interface {
	Marshal(i interface{}) ([]byte, error)
	Unmarshal(b []byte, i interface{}) error
}

type yamlParser struct{}

func (parser *yamlParser) Marshal(i interface{}) ([]byte, error) {
	return yaml.Marshal(i)
}

func (parser *yamlParser) Unmarshal(b []byte, i interface{}) error {
	return yaml.Unmarshal(b, i)
}

type jsonParser struct{}

func (parser *jsonParser) Marshal(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

func (parser *jsonParser) Unmarshal(b []byte, i interface{}) error {
	return json.Unmarshal(b, i)
}
