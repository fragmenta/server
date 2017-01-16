# server
A wrapper for the net/http server offering a few other features:

* Config loading from a json config file (for use in setup/handlers)
* Levelled, Structured logging using a flexible set of loggers - easy to add your own custom loggers 
* Optional logging middleware for requests
* Scheduling of tasks at specific times of day and intervals

## Config 

The config package offers access to json config files containing dev/production/test configs. 

## Logging

The logging package offers structured, levelled logging which can be configured to send to a file, stdout, and/or other services like an influxdb server with additional plugin loggers. You can add as many loggers which log events as you want, and because logging is structured, each logger can decide which information to act on.

## Scheduling

A simplistic scheduling facility so that you can schedule actions (like sending a tweet) on app startup. 
