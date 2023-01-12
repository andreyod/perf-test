package config

type Config struct {
	// Setup configures the infrastructure.
	Setup *Setup `json:"setup,omitempty"`
}

type Setup struct {
	// Test is the test to perform: create;watch;delete
	Test string `json:"test"`
	// ObjectSizeKB is the size (in kilo bytes) of filler data to be placed in every object
	ObjectSizeKB int `json:"object_size_KB"`
	// Count is the number of objects to seed the API server with.
	ObjectCount int `json:"object_count,omitempty"`
	// Object name prefix
	NamePrefix string `json:"object_name_prefix"`
	// Parallelism determines how many parallel client connections to use
	Parallelism int `json:"parallelism,omitempty"`
}
