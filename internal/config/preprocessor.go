package config

// PreprocessorConfig holds configuration for a single preprocessor
type PreprocessorConfig struct {
	// Command is the executable to run for external preprocessors (optional)
	// If not specified, resolves to "geopub-<name>" on PATH
	Command string `toml:"command"`

	// Renderers is a list of renderer names this preprocessor applies to
	// If empty, applies to all renderers
	Renderers []string `toml:"renderers"`

	// Before is a list of preprocessor names that should run after this one
	Before []string `toml:"before"`

	// After is a list of preprocessor names that should run before this one
	After []string `toml:"after"`

	// Extra holds arbitrary extra configuration passed to the preprocessor
	Extra map[string]interface{}
}

// PreprocessorConfigs is a map of preprocessor name -> config
type PreprocessorConfigs map[string]*PreprocessorConfig
