package foundation

import "time"

// Config is the type definition of the JSON config file.
type Config struct {
	HostPort      string
	DBPath        string
	LitestreamYml string

	Startup       time.Time
	DevFileServer bool
}
