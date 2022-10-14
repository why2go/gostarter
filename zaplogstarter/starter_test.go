package zaplogstarter

import (
	"encoding/json"
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestZap(t *testing.T) {
	ymlText := `
zaplog:
  level: debug
  encoding: json
  encoderConfig:
    timeEncoder: rfc3339
`
	cfg := &zapConfig{}
	err := yaml.Unmarshal([]byte(ymlText), cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	b, _ := json.Marshal(cfg)

	fmt.Printf("cfg: %s\n", string(b))
}
