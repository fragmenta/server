// Package server offers a simple server with logging and configs
package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"
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
		Logger:            log.New(os.Stderr, "fragmenta: ", log.LstdFlags),
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

// PortString returns a string port suitable for passing to http.Server
func (s *Server) PortString() string {
	return fmt.Sprintf(":%d", s.port)
}

// Start starts an http server on the given port
func (s *Server) Start() error {
	server := &http.Server{
		Addr:         s.PortString(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return server.ListenAndServe()
}

// StartTLS starts an https server on the given port
// with tls cert/key from config keys.
func (s *Server) StartTLS(cert, key string) error {

	// Set up a new http server
	server := &http.Server{
		// Set the port in the preferred string format
		Addr: s.PortString(),
	}

	return server.ListenAndServeTLS(cert, key)
}

// StartTLSAutocert starts an https server on the given port
// by requesting certs from an ACME provider.
// The server must be on a public IP which matches the
// DNS for the domains.
func (s *Server) StartTLSAutocert(email string, domains string) error {
	autocertDomains := strings.Split(domains, " ")
	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Email:      email,                                      // Email for problems with certs
		HostPolicy: autocert.HostWhitelist(autocertDomains...), // Domains to request certs for
		Cache:      autocert.DirCache("secrets"),               // Cache certs in secrets folder
	}
	server := s.ConfiguredTLSServer(certManager)
	return server.ListenAndServeTLS("", "")
}

// ConfiguredTLSServer returns a TLS server instance with a secure config
// this server has read/write timeouts set to 20 seconds,
// prefers server cipher suites and only uses certain accelerated curves
// see - https://blog.gopheracademy.com/advent-2016/exposing-go-on-the-internet/
func (s *Server) ConfiguredTLSServer(certManager *autocert.Manager) *http.Server {

	return &http.Server{
		// Set the port in the preferred string format
		Addr: s.PortString(),
		/*
			// The default server from net/http has no timeouts
			ReadTimeout:  20 * time.Second,
			WriteTimeout: 20 * time.Second,
			IdleTimeout:  120 * time.Second,
		*/
		// This TLS config follows recommendations in the above article
		TLSConfig: &tls.Config{
			// Pass in a cert manager if you want one set
			// this will only be used if the server Certificates are empty
			GetCertificate: certManager.GetCertificate,

			// VersionTLS11 or VersionTLS12 would exclude many browsers
			// inc. Android 4.x, IE 10, Opera 12.17, Safari 6
			// So unfortunately not acceptable as a default yet
			// Current default here for clarity
			MinVersion: tls.VersionTLS10,

			// Causes servers to use Go's default ciphersuite preferences,
			// which are tuned to avoid attacks. Does nothing on clients.
			PreferServerCipherSuites: true,
			// Only use curves which have assembly implementations
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				//	tls.X25519, // Go 1.8 only
			},
		},
	}

}

// StartRedirectAll starts redirecting from port given to the given url
// this should be called before StartTLS if redirecting to https
func (s *Server) StartRedirectAll(p int, url string) {
	port := fmt.Sprintf(":%d", p)
	// Listen and server on port p in a separate goroutine
	go func() {
		http.ListenAndServe(port, &redirectHandler{redirect: url})
	}()
}

// redirectHandler is useful if serving tls direct (not behind a proxy)
// and a redirect from port 80 is required.
type redirectHandler struct {
	redirect string
}

// ServeHTTP on this handler simply redirects to the main site
func (m *redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, m.redirect, http.StatusMovedPermanently)
}
