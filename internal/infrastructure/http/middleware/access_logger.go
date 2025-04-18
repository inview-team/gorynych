package middleware

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// Create a request logging middleware handler called Logger
type Logger struct {
	handler http.Handler
}

// ServeHTTP handles the request by passing it to the real handler and logging the request details
func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	l.handler.ServeHTTP(w, r)
	log.Infof("%s %s %v", r.Method, r.URL.Path, r.Header, time.Since(start))
}

// NewLogger constructs a new Logger middleware handler
func NewLogger(handlerToWrap http.Handler) *Logger {
	return &Logger{handlerToWrap}
}
