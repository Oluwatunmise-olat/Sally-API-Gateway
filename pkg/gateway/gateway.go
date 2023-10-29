package gateway

import (
	"context"
	"errors"
	"fmt"
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/logger"
	"github.com/gorilla/mux"
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

	logger.Log.Infoln("Shutting Down Server 🕹")
	// Close Open Listeners and Drain Active Connections
	if err := serverInstance.Shutdown(context.TODO()); err != nil {
		logger.Log.Errorln("An error occurred shutting down the server", err)
	}
}

func proxy(upstream map[string]string, isRegexPath bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method := strings.ToLower(r.Method)
		targetUpstream, ok := upstream[method]

		if !isRegexPath {
			targetUpstream = mapListenerPathAndUpstreamPath(mux.Vars(r), targetUpstream)
		} else {
			// Here we append any route called to the upstream server ie. /github/{account+} which was transformed to /github/account
			// would append any path after /account to the target upstream
			targetUpstream = targetUpstream + "/" + mux.Vars(r)["route"]
		}

		if !ok {
			logger.Log.Errorln(fmt.Sprintf("No Upstream Server Found For Path %s And Method %s", r.URL.Path, method))
			http404Handler(w)
		}

		parsedUrl, err := url.Parse(targetUpstream)

		if err == nil {
			proxy := httputil.NewSingleHostReverseProxy(parsedUrl)
			proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
				constructHttpResponse(w, http.StatusBadGateway, false, fmt.Sprintf("An error occurred accessing service with uri: %s", request.URL.String()))
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
