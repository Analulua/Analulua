package http

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"newdemo1/resource/jaeger"

	"github.com/gorilla/handlers"
	"github.com/rs/cors"
	"newdemo1/resource/jaeger/common/telemetry"
	telhttp "newdemo1/resource/jaeger/common/telemetry/instrumentation/http"
)

const (
	cacheMaxAge            = 86400
	defaultHealthCheckPath = "/health"
	defaultMatricPath      = "/metrics"
)

func CORS(handler http.Handler) http.Handler {
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           cacheMaxAge,
	})

	return corsHandler.Handler(handler)
}

func Recover(handler http.Handler) http.Handler {
	return handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(handler)
}

func HealthCheck(handler http.Handler, path string, providers ...jaeger.ChecksProvider) http.Handler {
	hc := jaeger.New(
		&jaeger.Health{
			Version:   "1",
			ReleaseID: "1.0.0",
		}, providers...)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == path {
			hc.Handler(w, r)
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}

type Option func(http.Handler) http.Handler

func WithRecovery() Option {
	return Recover
}

func WithCompression() Option {
	return handlers.CompressHandler
}

func WithCORS() Option {
	return CORS
}

func WithTelemetry(api *telemetry.API) Option {
	return func(h http.Handler) http.Handler {
		return telhttp.Telemetry(h, api)
	}
}

func WithHealthCheck(providers ...jaeger.ChecksProvider) Option {
	return WithHealthCheckPath(defaultHealthCheckPath, providers...)
}

func WithHealthCheckPathFilter(path string, filter map[string]string, providers ...jaeger.ChecksProvider) Option {
	return func(h http.Handler) http.Handler {
		hc := jaeger.New(
			&jaeger.Health{
				Version:   "1",
				ReleaseID: "1.0.0",
			}, providers...)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var pass bool
			if filter != nil {
				for k, v := range filter {
					if r.Header.Get(k) == v {
						pass = true
					} else {
						pass = false
						break
					}
				}
			}

			if r.URL.Path == path && pass {
				hc.Handler(w, r)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
}

func WithHealthCheckPath(path string, providers ...jaeger.ChecksProvider) Option {
	return func(h http.Handler) http.Handler {
		return HealthCheck(h, path, providers...)
	}
}

func WithDefault() Option {
	return func(h http.Handler) http.Handler {
		return handlers.CompressHandler(Recover(h))
	}
}

func NewHandler(handler http.Handler, options ...Option) http.Handler {
	h := handler
	for _, option := range options {
		h = option(h)
	}

	return h
}

func DefaultHandler(handler http.Handler) http.Handler {
	return NewHandler(handler, WithDefault())
}

func ReverseProxy(pathURL string, toAddr string, api *telemetry.API) func() {
	origin, err := url.Parse(pathURL)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	director := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", origin.Host)
		req.URL.Scheme = "http"
		req.URL.Host = origin.Host
	}

	proxy := &httputil.ReverseProxy{Director: director}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == defaultHealthCheckPath:
			proxy.ServeHTTP(w, r)
		case r.URL.Path == defaultMatricPath && api != nil:
			api.MetricExportHandler.ServeHTTP(w, r)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}

	})

	srv := &http.Server{
		Addr:    toAddr,
		Handler: nil,
	}

	go func() {
		_ = srv.ListenAndServe()
	}()

	return func() { _ = srv.Shutdown(context.Background()) }
}
