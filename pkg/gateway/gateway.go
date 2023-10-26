package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/exception"
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/logger"
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/validator"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func Bootstrap(configFileName string) {
	config, err := validator.ValidateConfigurationFile(configFileName)

	if err != nil {
		_ = exception.ErrorHandler(&exception.ErrorExceptions{Message: "...Testing", Err: err})
	}

	router := startServer(&config)
	serverInstance := &http.Server{Addr: ":8080", Handler: router}

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

func startServer(transformedConfig *validator.TransformedConfig) *mux.Router {
	router := mux.NewRouter()

	registerRoutes(router, transformedConfig)

	return router
}

// Register Config Routes with mux
func registerRoutes(r *mux.Router, transformedConfig *validator.TransformedConfig) {
	for listeningPath, allowedMethods := range *transformedConfig {
		pathMethods := make([]string, 0)
		pathUpstreams := make(map[string]string)

		for method, extras := range allowedMethods {
			pathMethods = append(pathMethods, method)
			pathUpstreams[method] = extras.Url
		}

		logger.Log.WithField(listeningPath, pathMethods).Infoln("Mapping Routes")
		r.PathPrefix(listeningPath).Methods(pathMethods...).HandlerFunc(proxy(pathUpstreams))
	}

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handle404(w)
	})
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
		}

		logger.Log.Errorln("An error occurred while parsing upstream server url", err)
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
	w.Header().Set("Content-Type", "application/json")
	payload := map[string]interface{}{
		"status":  false,
		"message": "Sigh 🥴, We Could Not Find The Route You Are Trying To Access",
	}

	response, _ := json.Marshal(payload)
	w.WriteHeader(http.StatusNotFound)
	w.Write(response)
}
