package config

// 定义可以处理的文件格式

type fileFormat struct {
	name           string
	fileSuffix     []string
	fieldTagPrefix string
	parser         configParser
}

var (
	jsonFormat = &fileFormat{
		name:           "JSON",
		fileSuffix:     []string{"json"},
		fieldTagPrefix: "json",
		parser:         &jsonParser{},
	}
	yamlFormat = &fileFormat{
		name:           "YAML",
		fileSuffix:     []string{"yml", "yaml"},
		fieldTagPrefix: "yaml",
		parser:         &yamlParser{},
	}

	allSupportedFileFormats = []*fileFormat{
		jsonFormat,
		yamlFormat,
	}
)
