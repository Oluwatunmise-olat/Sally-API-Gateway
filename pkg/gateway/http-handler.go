package gateway

import (
	"encoding/json"
	"errors"
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/files"
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/logger"
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/validator"
	"github.com/gorilla/mux"
	"net/http"
)

func registerBaseRoutes(router *mux.Router) {
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http404Handler(w)
	})

	router.Path("/gw-upload").Methods(http.MethodPost).HandlerFunc(configFileUploadHandler)

	loadValidateAndRegisterConfigFileIfExists(router)
}

func constructHttpResponse(w http.ResponseWriter, statusCode int, isSuccessful bool, message string) {
	payload := map[string]interface{}{
		"status":  isSuccessful,
		"message": message,
	}

	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

func http404Handler(w http.ResponseWriter) {
	constructHttpResponse(w, http.StatusNotFound, false, "Sigh 🥴, We Could Not Find The Route You Are Trying To Access")
}

func configFileUploadHandler(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("config_file")

	if err != nil {
		constructHttpResponse(w, http.StatusBadRequest, false, "An error occurred parsing config file")
		return
	}
	defer file.Close()
	err = files.CreateOrUpdateConfigFile(file)

	if err != nil {
		constructHttpResponse(w, http.StatusBadRequest, false, "An error occurred")
		return
	}

	config, err := validator.ValidateConfigurationFile("config/app.yaml")
	if err != nil {
		constructHttpResponse(w, http.StatusBadRequest, false, "An error occurred validating config file")
		return
	}

	registerConfigFileRoutes(router, &config)

	constructHttpResponse(w, http.StatusOK, true, "Config File Uploaded Successfully")
	return
}

// Register Config Routes with mux
func registerConfigFileRoutes(r *mux.Router, transformedConfig *validator.TransformedConfig) {

	for listeningPath, allowedMethods := range *transformedConfig {
		pathMethods := make([]string, 0)
		pathUpstreams := make(map[string]string)

		for method, extras := range allowedMethods {
			pathMethods = append(pathMethods, method)
			pathUpstreams[method] = extras.Url
		}

		isRegexPath, listeningPath := files.CheckIfIsRegexPathAndTransformRoute(listeningPath)

		logger.Log.WithField(listeningPath, pathMethods).Infoln("Mapping Routes")

		r.Path(listeningPath).Methods(pathMethods...).HandlerFunc(proxy(pathUpstreams, isRegexPath))
	}
}

func loadValidateAndRegisterConfigFileIfExists(router *mux.Router) error {
	if files.CheckIfPathExists("config/app.yaml") {
		config, err := validator.ValidateConfigurationFile("config/app.yaml")

		if err != nil {
			logger.Log.Errorln("An error occurred validating config file")
			return errors.New("An error occurred validating config file")
		}

		registerConfigFileRoutes(router, &config)
	}

	return nil
}
