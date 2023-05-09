package globalflow

import "globalflow/config"

// Container is a struct that serves as a simple IoC container.
type Container struct {
	// Configuration is the global configuration.
	Configuration *config.Configuration
}
