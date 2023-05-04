package globalflow

// Configuration is a struct that contains the global configuration.
type Configuration struct {
	// DatabasePath is the path to the database file.
	DatabasePath string

	// NodeID is the ID of the node.
	NodeID string
}

// NewConfiguration creates a new configuration with default values.
func NewConfiguration() *Configuration {
	return &Configuration{
		DatabasePath: ".",
	}
}
