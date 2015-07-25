// Package server offers a simple server with logging and configs
package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func (s *Server) runSetup() {

	http.Handle("/", s)
	err := s.Start()
	if err != nil {
		s.Fatal("Error running setup", err)
	}

}

// Generate a random 32 byte key encoded in base64
func randomKey() string {
	k := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return ""
	}
	return hex.EncodeToString(k)
}

func (s *Server) setupWithData(form url.Values) {

	// Add the request form values
	for k, v := range form {
		s.configDevelopment[k] = v[0]
	}

	path, err := os.Getwd()
	if err != nil {
		s.Fatal("Error finding path", err)
	}

	s.configDevelopment["db_version"] = "12"
	s.configDevelopment["path"] = path // Use current path
	s.configDevelopment["port"] = "3000"
	s.configDevelopment["log"] = "log/development.log"
	s.configDevelopment["hmac_key"] = randomKey()
	s.configDevelopment["secret_key"] = randomKey()

	s.configProduction = s.configDevelopment
	s.configProduction["port"] = "4000"
	s.configProduction["log"] = "log/production.log"
	s.configProduction["hmac_key"] = randomKey()
	s.configProduction["secret_key"] = randomKey()

	// Save out the config for later
	s.saveConfig()

	// Run migrations

	// Restart the server? Need to somehow tell the app that we need to proceed with setup
	// how best to do that? can the app register a hook which will get a callback when we're done?

}

func (s *Server) saveConfig() {

	configs := map[string]map[string]string{
		"production":  s.configProduction,
		"development": s.configDevelopment,
		//     "test":s.configTest,
	}

	configsJSON, err := json.Marshal(configs)
	if err != nil {
		s.Fatal("Error finding path", err)
	}

	fmt.Printf("c:%s\n", configsJSON)

}

// A rudimentary dispatcher for web requests, we serve one setup page
// to get the config information we need to set up the app, then we start the app proper
func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	// We handle two paths - GET /setup and POST /setup
	switch request.Method {
	case "POST":
		// Process the posted form data in order to create our configuration and site
		err := request.ParseForm()
		if err != nil {
			s.Fatal("Error parsing setup form", err)
		}

		s.setupWithData(request.Form)

	default:
		// Serve our setup template file from the fragmenta templates dir
		http.ServeFile(writer, request, setupTemplatePath())
	}

}

func setupTemplatePath() string {
	return os.ExpandEnv("$GOPATH/src/github.com/fragmenta/fragmenta/templates/setup/setup.html")
}
