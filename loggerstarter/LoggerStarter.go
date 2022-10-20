package loggerstarter

var (
	LoggerConfig loggerConfig
)

type loggerConfig struct {
	Level        string
	Encoding     string   `json:"encoding" yaml:"encoding"`
	OutputPaths  []string `json:"outputPaths" yaml:"outputPaths"`
	TimeFormat   string   `yaml:"timeFormat" json:"timeFormat"`
	DurationUnit string   `yaml:"durationUnit" json:"durationUnit"`
}
