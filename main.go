package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	cfg := loadConfig()
	logger := newLogger(cfg.LogFormat, cfg.LogLevel)

	// Use a plain http.Server with the handler directly — no ServeMux.
	// ServeMux does path cleaning which collapses the // in https:// to /,
	// causing an unwanted 301 before we ever reach our handler.
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      loggingMiddleware(logger)(corsProxyHandler(cfg, logger)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting cors proxy", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "err", err)
	}
}

// corsProxyHandler extracts the target URL from the request path and proxies to it.
//
// Request format:
//
//	https://proxy.com/{upstream-url}
//	https://proxy.com/https://ipinfo.io/json
//
// NOTE: the server must be set as the root Handler directly (not via ServeMux),
// because ServeMux path-cleans double slashes and issues a 301 redirect.
func corsProxyHandler(cfg *Config, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w, r, cfg)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// r.URL.Path starts with "/", the target URL follows.
		// e.g. /https://ipinfo.io/json -> https://ipinfo.io/json
		rawTarget := strings.TrimPrefix(r.URL.Path, "/")
		if r.URL.RawQuery != "" {
			rawTarget += "?" + r.URL.RawQuery
		}

		target, err := url.ParseRequestURI(rawTarget)
		if err != nil || target.Host == "" {
			logger.Warn("invalid target url", "raw", rawTarget, "err", err)
			http.Error(w, "invalid target URL in path", http.StatusBadRequest)
			return
		}

		if len(cfg.AllowedHosts) > 0 && !isAllowedHost(target.Host, cfg.AllowedHosts) {
			logger.Warn("host not allowed", "host", target.Host)
			http.Error(w, "host not allowed", http.StatusForbidden)
			return
		}

		proxyBase := proxyBaseURL(r)

		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL = target
				req.Host = target.Host

				clientIP := req.RemoteAddr
				if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
					clientIP = prior + ", " + clientIP
				}
				req.Header.Set("X-Forwarded-For", clientIP)

				logger.Debug("proxying request",
					"method", req.Method,
					"target", target.String(),
				)
			},
			// Rewrite Location headers in redirect responses so the client
			// follows them through the proxy, not directly to the upstream.
			ModifyResponse: func(resp *http.Response) error {
				resp.Header.Del("Access-Control-Allow-Origin")
				resp.Header.Del("Access-Control-Allow-Methods")
				resp.Header.Del("Access-Control-Allow-Headers")
				resp.Header.Del("Access-Control-Allow-Credentials")
				resp.Header.Del("Access-Control-Expose-Headers")
				resp.Header.Del("Access-Control-Max-Age")

				if loc := resp.Header.Get("Location"); loc != "" {
					resp.Header.Set("Location", proxyBase+"/"+loc)
				}
				return nil
			},
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				logger.Error("proxy error", "err", err, "target", target.String())
				http.Error(w, "bad gateway", http.StatusBadGateway)
			},
		}

		proxy.ServeHTTP(w, r)
	})
}

// proxyBaseURL returns the scheme+host of the proxy itself for rewriting Location headers.
func proxyBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func setCORSHeaders(w http.ResponseWriter, r *http.Request, cfg *Config) {
	origin := r.Header.Get("Origin")

	if cfg.AllowOrigins == "*" {
		if !cfg.HideAllowOriginsHeader {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
	} else if origin != "" && isAllowedOrigin(origin, cfg.AllowOrigins) {
		if !cfg.HideAllowOriginsHeader {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Vary", "Origin")
	}

	w.Header().Set("Access-Control-Allow-Methods", cfg.AllowMethods)
	w.Header().Set("Access-Control-Allow-Headers", cfg.AllowHeaders)
	w.Header().Set("Access-Control-Max-Age", cfg.MaxAge)

	if cfg.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}

func isAllowedOrigin(origin, allowed string) bool {
	for _, o := range strings.Split(allowed, ",") {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

func isAllowedHost(host string, allowed []string) bool {
	hostOnly := host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		hostOnly = host[:idx]
	}
	for _, a := range allowed {
		a = strings.TrimSpace(a)
		if a == hostOnly || a == host {
			return true
		}
	}
	return false
}

// loggingMiddleware records method, path, status, and latency for every request.
func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rw, r)

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
