package server

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
)

// Port returns the port of the server
func (s *Server) Port() int {
	return s.port
}

// Mode returns the mode (production or development)
func (s *Server) Mode() string {
	if s.production {
		return "Production"
	}
	return "Development"
}

// Production tells the caller if this server is in production mode or not?
func (s *Server) Production() bool {
	return s.production
}

// Configuration returns the map of configuration keys to values
func (s *Server) Configuration() map[string]string {
	if s.production {
		return s.configProduction
	}
	return s.configDevelopment

}

// Config returns a specific configuration value or "" if no value
func (s *Server) Config(key string) string {
	return s.Configuration()[key]
}

// ConfigInt returns the current configuration value as int64, or 0 if no value
func (s *Server) ConfigInt(key string) int64 {
	v := s.Config(key)
	if v != "" {
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return i
		}
	}
	return 0
}

// ConfigBool returns the current configuration value as bool (yes=true, no=false), or false if no value
func (s *Server) ConfigBool(key string) bool {
	v := s.Config(key)
	return (v == "yes")
}

// configPath returns our expected config file path
func (s *Server) configPath() string {
	return "secrets/fragmenta.json"
}

// Read our config file and set up the server accordingly
func (s *Server) readConfig() error {

	path := s.configPath()

	// Read the config json file
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Error opening config %s %v", path, err)
	}

	var data map[string]map[string]string
	err = json.Unmarshal(file, &data)
	if err != nil {
		return fmt.Errorf("Error reading config %s %v", path, err)
	}

	s.configDevelopment = data["development"]
	s.configProduction = data["production"]
	s.configTest = data["test"]

	// Update our port from the config port if we have it
	portString := s.Config("port")
	if portString != "" {
		s.port, err = strconv.Atoi(portString)
		if err != nil {
			return fmt.Errorf("Error reading port %s", err)
		}
	}

	return nil
}

// readArguments reads command line arguments
func (s *Server) readArguments() error {

	var p int
	flag.IntVar(&p, "p", p, "Port")
	flag.Parse()

	if p > 0 {
		s.port = p
	}

	return nil
}
