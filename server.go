// Package server offers a simple server with logging and configs
package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

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
		Logger:            log.New(os.Stderr, "fragmenta: ", log.LstdFlags), // default to a stderr logger
	}

	err := s.readConfig()
	if err != nil {
		// Run the setup, till we have collected enough information
		s.runSetup()
		return s, err
	}
	err = s.readArguments()
	if err != nil {
		return s, err
	}

	return s, err
}

// Log this format and arguments
func (s *Server) Log(format string, v ...interface{}) {
	// Call our internal logger with these arguments
	s.Logger.Printf(format, v...)
}

// Fatal logs this format and arguments, and then exits with status 1
func (s *Server) Fatal(format string, v ...interface{}) {
	// Call our internal logger with these arguments
	s.Logger.Printf(format, v...)

	// Now exit
	os.Exit(1)
}

// Start the http server
func (s *Server) Start() error {
	p := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(p, nil)

}
