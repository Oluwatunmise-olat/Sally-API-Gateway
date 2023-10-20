package validator

import (
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/exception"
	"gopkg.in/yaml.v2"
	"io"
	"os"
)

func ValidateConfigurationFile(fileName string) (TransformedConfig, error) {
	configPlaybook, err := validateStruct(fileName)

	if err != nil {
		return TransformedConfig{}, exception.ErrorHandler(&exception.ErrorExceptions{
			Message: err.Error(), Err: err,
		})
	}

	transformedPlaybook, err := validateResourcePathsAndTransform(configPlaybook)

	if err != nil {
		return TransformedConfig{}, exception.ErrorHandler(&exception.ErrorExceptions{
			Message: err.Error(), Err: err,
		})
	}

	return transformedPlaybook, nil
}

func validateStruct(fileName string) (GatewayPlaybook, error) {
	yamlFile, err := os.Open(fileName)

	if err != nil {
		return GatewayPlaybook{}, exception.ErrorHandler(&exception.ErrorExceptions{
			Message: "Failed to open config file.",
			Err:     err,
		})
	}
	defer yamlFile.Close()

	contents, err := io.ReadAll(yamlFile)
	if err != nil {
		e := exception.ErrorHandler(&exception.ErrorExceptions{
			Message: "Failed to read contents of config file.",
			Err:     err,
		})

		return GatewayPlaybook{}, e
	}

	var config GatewayPlaybook
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return GatewayPlaybook{}, exception.ErrorHandler(&exception.ErrorExceptions{
			Message: "Invalid config file.",
			Err:     err,
		})
	}

	return config, nil
}

// A file might be parsed successfully, but might be error-prone
func validateResourcePathsAndTransform(config GatewayPlaybook) (TransformedConfig, error) {
	baseTag := transformBaseTags(config.Tags)
	transformedConfig := TransformedConfig{}

	for urlPath, resources := range config.Paths {
		transformedResource, err := getDefinedHttpResourcesForUrlPath(resources, baseTag)

		if err != nil {
			return TransformedConfig{}, err
		}
		transformedConfig[urlPath] = transformedResource

	}

	return transformedConfig, nil
}

func getDefinedHttpResourcesForUrlPath(payload YamlConfigPath, baseTags TransformedBaseTag) (TransformedResourcePath, error) {
	httpResources := make(TransformedResourcePath)

	if !isEmptyString(payload.Get.Summary) {
		target, err := validateResource(payload.Get.XTarget, payload.Get.Tag, baseTags)

		if err != nil {
			return TransformedResourcePath{}, err
		}

		httpResources["get"] = struct {
			Url       string
			Operation YamlConfigPathOperation
		}{Url: target, Operation: payload.Get}
	}

	if !isEmptyString(payload.Delete.Summary) {
		target, err := validateResource(payload.Delete.XTarget, payload.Delete.Tag, baseTags)

		if err != nil {
			return TransformedResourcePath{}, err
		}

		httpResources["delete"] = struct {
			Url       string
			Operation YamlConfigPathOperation
		}{Url: target, Operation: payload.Delete}
	}

	if !isEmptyString(payload.Post.Summary) {
		target, err := validateResource(payload.Post.XTarget, payload.Post.Tag, baseTags)

		if err != nil {
			return TransformedResourcePath{}, err
		}

		httpResources["post"] = struct {
			Url       string
			Operation YamlConfigPathOperation
		}{Url: target, Operation: payload.Post}
	}

	if !isEmptyString(payload.Put.Summary) {
		target, err := validateResource(payload.Put.XTarget, payload.Put.Tag, baseTags)

		if err != nil {
			return TransformedResourcePath{}, err
		}

		httpResources["put"] = struct {
			Url       string
			Operation YamlConfigPathOperation
		}{Url: target, Operation: payload.Put}
	}

	if !isEmptyString(payload.Options.Summary) {
		target, err := validateResource(payload.Options.XTarget, payload.Options.Tag, baseTags)

		if err != nil {
			return TransformedResourcePath{}, err
		}

		httpResources["options"] = struct {
			Url       string
			Operation YamlConfigPathOperation
		}{Url: target, Operation: payload.Options}
	}

	return httpResources, nil
}

func isEmptyString(value string) bool {
	return len(value) == 0
}

func transformBaseTags(tags []YamlConfigTag) TransformedBaseTag {
	base := make(TransformedBaseTag)

	for _, tag := range tags {
		base[tag.Name] = struct{ XTarget string }{XTarget: tag.XTarget}
	}

	return base
}

func validateResource(target, resourceTag string, baseTags TransformedBaseTag) (string, error) {
	// Since resource target is empty, opt out for base tag target
	if isEmptyString(target) && isEmptyString(resourceTag) {
		return "", exception.ErrorHandler(&exception.ErrorExceptions{Message: "No resource target specified", Err: nil})
	}

	if isEmptyString(target) && !isEmptyString(resourceTag) {
		if _, ok := baseTags[resourceTag]; !ok {
			return "", exception.ErrorHandler(&exception.ErrorExceptions{Message: "No resource target specified", Err: nil})
		}

		target = baseTags[resourceTag].XTarget
	}

	return target, nil
}

// Once config file is passed, validate in background and return response later
