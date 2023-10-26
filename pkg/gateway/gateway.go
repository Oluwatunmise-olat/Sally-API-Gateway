package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/logger"
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/validator"
	"github.com/gorilla/mux"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var router *mux.Router

func Bootstrap() {
	router = mux.NewRouter().StrictSlash(false)
	registerBaseRoutes(router)

	serverInstance := &http.Server{Addr: fmt.Sprintf(":%s", os.Getenv("PORT")), Handler: router}

	go func() {
		logger.Log.Infoln("Server started 🚀")

		if err := serverInstance.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Errorln("An error occurred starting the server", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	<-ctx.Done()

	logger.Log.Infoln("Shutting Down Server ...")
	// Close Open Listeners and Drain Active Connections
	if err := serverInstance.Shutdown(context.TODO()); err != nil {
		logger.Log.Errorln("An error occurred shutting down the server", err)
	}
}

func registerBaseRoutes(router *mux.Router) {
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handle404(w)
	})

	router.Path("/gw-upload").Methods(http.MethodPost).HandlerFunc(handleConfigUpload)
}

// Register Config Routes with mux
func registerConfigRoutes(r *mux.Router, transformedConfig *validator.TransformedConfig) {

	for listeningPath, allowedMethods := range *transformedConfig {
		pathMethods := make([]string, 0)
		pathUpstreams := make(map[string]string)

		for method, extras := range allowedMethods {
			pathMethods = append(pathMethods, method)
			pathUpstreams[method] = extras.Url
		}

		logger.Log.WithField(listeningPath, pathMethods).Infoln("Mapping Routes")

		r.Path(listeningPath).Methods(pathMethods...).HandlerFunc(proxy(pathUpstreams))
	}
}

func proxy(upstream map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method := strings.ToLower(r.Method)
		targetUpstream, ok := upstream[method]
		targetUpstream = mapListenerPathAndUpstreamPath(mux.Vars(r), targetUpstream)

		if !ok {
			logger.Log.Errorln(fmt.Sprintf("No Upstream Server Found For Path %s And Method %s", r.URL.Path, method))
			handle404(w)
		}

		parsedUrl, err := url.Parse(targetUpstream)

		if err == nil {
			proxy := httputil.NewSingleHostReverseProxy(parsedUrl)
			proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
				constructResponse(w, http.StatusBadGateway, false, fmt.Sprintf("An error occurred accessing service with uri: %s", request.URL.String()))
			}

			initialProxyDirector := proxy.Director
			proxy.Director = func(r *http.Request) {
				initialProxyDirector(r)

				clientIP := r.RemoteAddr
				// Append Ip address if header exists
				if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
					clientIP = ip + ", " + clientIP
				}

				r.Header.Set("X-Forwarded-For", clientIP)
				r.Header.Set("User-Agent", r.UserAgent())
				r.URL.Scheme = "https"
				r.Host = parsedUrl.Host
				r.URL.Host = parsedUrl.Host
				r.URL.Path = parsedUrl.Path
			}

			proxy.ServeHTTP(w, r)
		} else {
			logger.Log.WithField("targetUpstream", targetUpstream).Errorln("An error occurred while parsing upstream server url", err)
		}
	}
}

// Map listener parameters to proxy parameters i.e /users/{id} -> /proxy-server/users/{id} is mapped as /users/17 -> /proxy-server/users/17
func mapListenerPathAndUpstreamPath(params map[string]string, targetUpstream string) string {
	var parsedUrl string = targetUpstream

	for key, value := range params {
		parsedUrl = strings.Replace(targetUpstream, fmt.Sprintf("{%s}", key), value, -1)
	}

	return parsedUrl
}

func handle404(w http.ResponseWriter) {
	constructResponse(w, http.StatusNotFound, false, "Sigh 🥴, We Could Not Find The Route You Are Trying To Access")
}

func handleConfigUpload(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("config_file")

	if err != nil {
		constructResponse(w, http.StatusBadRequest, false, "An error occurred parsing config file")
	}
	defer file.Close()

	createOrUpdateConfigFile(w, file)

	config, err := validator.ValidateConfigurationFile("config/app.yaml")
	if err != nil {
		constructResponse(w, http.StatusBadRequest, false, "An error occurred validating config file")
	}

	registerConfigRoutes(router, &config)
	constructResponse(w, http.StatusOK, true, "Config File Uploaded Successfully")
}

func createOrUpdateConfigFile(w http.ResponseWriter, file multipart.File) {
	newConfigFile, err := os.OpenFile("config/app.yaml", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Log.Errorln("Error occurred copying uploaded config file")
		constructResponse(w, http.StatusBadRequest, false, "An error occurred")
	}
	defer newConfigFile.Close()

	_, err = io.Copy(newConfigFile, file)
	if err != nil {
		logger.Log.Errorln("Error occurred copying uploaded config file")
		constructResponse(w, http.StatusBadRequest, false, "An error occurred")
	}
}

func constructResponse(w http.ResponseWriter, statusCode int, isSuccessful bool, message string) {
	payload := map[string]interface{}{
		"status":  isSuccessful,
		"message": message,
	}

	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}
