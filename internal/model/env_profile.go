package model

// EnvVar represents a single environment variable.
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// EnvProfile represents a named set of environment variables.
type EnvProfile struct {
	Name string   `json:"name"`
	Vars []EnvVar `json:"vars"`
	Path string   `json:"path"` // file path to the .env file
}
