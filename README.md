# server
A simple http server with logging and configs, using http.ListenAndServe


### Usage 

```Go 
	// Setup server
	server, err := server.New()
	if err != nil {
		fmt.Printf("Error creating server %s", err)
		return
	}

	// Write to log 
	server.Logf("#info Starting server in %s mode on port %d", server.Mode(), server.Port())

	// Start the server
	err = server.Start()
	if err != nil {
		server.Fatalf("Error starting server %s", err)
	}
```
