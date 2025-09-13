package foundation

import "time"

// Config is the type definition of the JSON config file.
type Config struct {
	HostPort string
	Message  string

	Startup       time.Time
	DevFileServer bool
}
