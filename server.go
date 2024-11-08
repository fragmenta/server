// Package server is a wrapper around the stdlib http server and x/autocert pkg.
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

// Server wraps the stdlib http server and x/autocert pkg with some setup.
type Server struct {

	// Which port to serve on - in 2.0 pass as argument for New()
	port int

	// Which mode we're in, read from ENV variable
	// Deprecated - due to be removed in 2.0
	production bool

	// Deprecated Logging - due to be removed in 2.0
	// Instead use the structured logging with server/log
	Logger Logger

	// Deprecated configs will be removed from the server object in 2.0
	// Use server/config instead to read the config from app.
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

	// Old style config read - this will be going away in Fragmenta 2.0
	// use server/config instead from the app
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

// Port returns the port of the server
func (s *Server) Port() int {
	return s.port
}

// PortString returns a string port suitable for passing to http.Server
func (s *Server) PortString() string {
	return fmt.Sprintf(":%d", s.port)
}

// Start starts an http server on the given port
func (s *Server) Start() error {
	server := &http.Server{
		// Set the port in the preferred string format
		Addr: s.PortString(),

		// The default server from net/http has no timeouts - set some limits
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       10 * time.Second, // IdleTimeout was introduced in Go 1.8

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

		// The default server from net/http has no timeouts - set some limits
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       10 * time.Second, // IdleTimeout was introduced in Go 1.8

		// This TLS config follows recommendations in the above article
		TLSConfig: &tls.Config{
			// VersionTLS12 would exclude many browsers
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
				tls.X25519, // Go 1.8 only
			},
		},
	}

	return server.ListenAndServeTLS(cert, key)
}

// StartTLSModern starts an https server on the given port
// with tls cert/key from config keys.
// TLS version is restricted to VersionTLS12 and cipher suites to known secure ones
// this attains an A+ score at https://www.ssllabs.com/
// while remaining compatible with most older clients
func (s *Server) StartTLSModern(cert, key string) error {

	// Set up a new http server
	server := &http.Server{
		// Set the port in the preferred string format
		Addr: s.PortString(),

		// The default server from net/http has no timeouts - set some limits
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       10 * time.Second, // IdleTimeout was introduced in Go 1.8

		// This TLS config follows recommendations in the above article
		TLSConfig: &tls.Config{
			// Require VersionTLS12 - this will exclude older browsers like Safari 5, IE 6
			// As of 2020 all browsers require this version because of vulnerabilities in 1.1
			MinVersion: tls.VersionTLS12,
			// Use Go's default ciphersuite preferences, which are tuned to avoid attacks.
			PreferServerCipherSuites: true,
			// Limit Curves to known secure ones
			CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			// Limit CipherSuites to known secure ones
			// TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256 is included for compatability reasons
			CipherSuites: []uint16{
				tls.TLS_CHACHA20_POLY1305_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
			},
		},
	}

	return server.ListenAndServeTLS(cert, key)
}

// StartTLSAuto starts an https server on the given port
// by requesting certs from an ACME provider using the http-01 challenge.
// it also starts a server on the port 80 to listen for challenges and redirect
// The server must be on a public IP which matches the
// DNS for the domains.
func (s *Server) StartTLSAuto(email, domains string) error {
	autocertDomains := strings.Split(domains, " ")
	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Email:      email,                                      // Email for problems with certs
		HostPolicy: autocert.HostWhitelist(autocertDomains...), // Domains to request certs for
		Cache:      autocert.DirCache("secrets"),               // Cache certs in secrets folder
	}
	// Handle all :80 traffic using autocert to allow http-01 challenge responses
	go func() {
		http.ListenAndServe(":80", certManager.HTTPHandler(nil))
	}()

	server := s.ConfiguredTLSServer(certManager)
	return server.ListenAndServeTLS("", "")
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

		// The default server from net/http has no timeouts - set some limits
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       10 * time.Second, // IdleTimeout was introduced in Go 1.8

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
				tls.X25519, // Go 1.8 only
			},
		},
	}

}

// StartRedirectAll starts redirecting all requests on the given port to the given host
// this should be called before StartTLS if redirecting http on port 80 to https
func (s *Server) StartRedirectAll(p int, host string) {
	port := fmt.Sprintf(":%d", p)
	// Listen and server on port p in a separate goroutine
	go func() {
		http.ListenAndServe(port, &redirectHandler{host: host})
	}()
}

// redirectHandler is useful if serving tls direct (not behind a proxy)
// and a redirect from port 80 is required.
type redirectHandler struct {
	host string
}

// ServeHTTP on this handler simply redirects to the main site
func (m *redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, m.host+r.URL.String(), http.StatusMovedPermanently)
}
