// Package config loads the settings the service needs to start.
//
// Values are read from the environment. A .env file is loaded first as
// a convenience for local development, and its absence is not an error.
// Required values have no fallback: a missing one fails the startup
// rather than being silently replaced.
package config
