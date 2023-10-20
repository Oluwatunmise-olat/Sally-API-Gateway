package validator

type GatewayPlaybook struct {
	Openapi string                    `yaml:"openapi" required:"true"`
	Info    YamlConfigInfo            `yaml:"info" required:"true"`
	Paths   map[string]YamlConfigPath `yaml:"paths" required:"true"`
	Tags    []YamlConfigTag           `yaml:"tags"`
}

type YamlConfigInfo struct {
	Title   string `yaml:"title" required:"true"`
	Version string `yaml:"version" required:"true"`
}

type YamlConfigTag struct {
	Name        string `yaml:"name" required:"true"`
	Description string `yaml:"description" required:"true"`
	XTarget     string `yaml:"x-target" required:"true"`
}

type YamlConfigPath struct {
	Get     YamlConfigPathOperation `yaml:"get,omitempty"`
	Post    YamlConfigPathOperation `yaml:"post,omitempty"`
	Delete  YamlConfigPathOperation `yaml:"delete,omitempty"`
	Patch   YamlConfigPathOperation `yaml:"patch,omitempty"`
	Put     YamlConfigPathOperation `yaml:"put,omitempty"`
	Options YamlConfigPathOperation `yaml:"options,omitempty"`
}

type YamlConfigPathOperation struct {
	Summary    string                                     `yaml:"summary" required:"true"`
	Responses  map[string]YamlConfigPathOperationResponse `yaml:"responses,omitempty"`
	Tag        string                                     `yaml:"x-tag"`
	XTarget    string                                     `yaml:"x-target,omitempty"`
	Parameters []YamlConfigPathOperationParameters        `yaml:"parameters,omitempty"`
}

type YamlConfigPathOperationResponse struct {
	Description string `yaml:"description"`
}

type YamlConfigPathOperationParameters struct {
	Name     string                                  `yaml:"name" required:"true"`
	In       string                                  `yaml:"in" required:"true"`
	Required bool                                    `yaml:"required" required:"true"`
	Schema   YamlConfigPathOperationParametersSchema `yaml:"schema" required:"true"`
}

type YamlConfigPathOperationParametersSchema struct {
	Type string `yaml:"type"`
}

type TransformedBaseTag map[string]struct {
	XTarget string
}

type TransformedResourcePath map[string]struct {
	Url       string
	Operation YamlConfigPathOperation
}

type TransformedConfig map[string]TransformedResourcePath
