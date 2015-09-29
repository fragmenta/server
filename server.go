// Package server offers a simple server with logging and configs
package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// TODO: Graceful restart would be nice - http://grisha.org/blog/2014/06/03/graceful-restart-in-golang/

// Logger interface for a simple logger (the stdlib log pkg and the fragmenta log pkg conform)
type Logger interface {
	Printf(format string, args ...interface{})
}

// Server holds the config and logger for the app
type Server struct {
	// Our internal logger instance
	Logger Logger

	// Which port to serve on
	port int

	// Which env mode we're in, read from ENV variable
	production bool

	// Server configs - access with Config(string)
	configProduction  map[string]string
	configDevelopment map[string]string
	configTest        map[string]string
}

// New creates a new server instance
func New() (*Server, error) {

	// Check environment variable to see if we are in production mode
	prod := false
	if os.Getenv("FRAG_ENV") == "production" {
		prod = true
	}

	// Set up a new server
	s := &Server{
		port:              3000,
		production:        prod,
		configProduction:  make(map[string]string),
		configDevelopment: make(map[string]string),
		configTest:        make(map[string]string),
		Logger:            log.New(os.Stderr, "fragmenta: ", log.LstdFlags), // default to a stderr logger
	}

	err := s.readConfig()
	if err != nil {
		return s, err
	}
	err = s.readArguments()
	if err != nil {
		return s, err
	}

	return s, err
}

// Logf logs the message with the given arguments to our internal logger
func (s *Server) Logf(format string, v ...interface{}) {
	s.Logger.Printf(format, v...)
}

// Log logs the message to our internal logger
func (s *Server) Log(message string) {
	s.Logf(message)
}

// Fatalf the message with the given arguments to our internal logger, and then exits with status 1
func (s *Server) Fatalf(format string, v ...interface{}) {
	s.Logger.Printf(format, v...)

	// Now exit
	os.Exit(1)
}

// Fatal logs the message, and then exits with status 1
func (s *Server) Fatal(format string) {
	s.Fatalf(format)
}

// Timef logs a time since starting, when used with defer at the start of a function to time
// Usage: defer s.Timef("Completed %s in %s",time.Now(),args...)
func (s *Server) Timef(format string, start time.Time, v ...interface{}) {
	end := time.Since(start).String()
	var args []interface{}
	args = append(args, end)
	args = append(args, v...)
	s.Logf(format, args...)
}

// Start starts the http server on our given port
func (s *Server) Start() error {
	p := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(p, nil)
}
