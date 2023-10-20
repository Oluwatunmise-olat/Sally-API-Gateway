package gateway

import (
	"context"
	"errors"
	"fmt"
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/exception"
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

	ctx := shutDown()

	fmt.Println("Listening")
	// go func() {
	if err := serverInstance.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		// Log Error
		// }
	}
	// ()

	<-ctx.Done()
	// Close Open Listeners and Drain Active Connections
	if err := serverInstance.Shutdown(context.TODO()); err != nil {
		// 	Error
	}
}

func startServer(transformedConfig *validator.TransformedConfig) *mux.Router {
	router := mux.NewRouter()
	registerRoutes(router, transformedConfig)

	return router
}

func registerRoutes(r *mux.Router, transformedConfig *validator.TransformedConfig) {
	for listeningPath, allowedMethods := range *transformedConfig {
		pathMethods := make([]string, 0)
		pathUpstreams := make(map[string]string)

		for method, extras := range allowedMethods {
			pathMethods = append(pathMethods, method)
			pathUpstreams[method] = extras.Url
		}

		r.PathPrefix(listeningPath).Methods(pathMethods...).HandlerFunc(proxy(pathUpstreams))
	}
}

func proxy(upstream map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method := strings.ToLower(r.Method)
		targetUpstream, ok := upstream[method]

		if !ok {
			//	Log Error, 404 Error
			fmt.Println(">>", method, upstream)
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

				fmt.Println(parsedUrl, parsedUrl.Scheme, targetUpstream)

				r.Header.Set("X-Forwarded-For", clientIP)
				r.Header.Set("User-Agent", r.UserAgent())
				r.URL.Scheme = "https"
				r.Host = parsedUrl.Host
				r.URL.Host = parsedUrl.Host
				r.URL.Path = parsedUrl.Path
			}

			proxy.ServeHTTP(w, r)
		}

		// Throw Error Parsing Url, Err
	}
}

func shutDown() context.Context {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	return ctx
}
