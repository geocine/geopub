package runner

import (
	"fmt"
	"log"

	"github.com/geocine/geopub/internal/config"
	"github.com/geocine/geopub/internal/models"
	"github.com/geocine/geopub/internal/preprocessor/index"
)

// Runner manages the preprocessor pipeline
type Runner struct {
	cfg                  *config.Config
	renderer             string
	verbose              bool
	disableExternals     bool
	includeDefaults      bool
	builtinPreprocessors map[string]func(*models.Book) error
}

// NewRunner creates a new preprocessor runner
func NewRunner(cfg *config.Config, renderer string) *Runner {
	return &Runner{
		cfg:              cfg,
		renderer:         renderer,
		verbose:          false,
		disableExternals: false,
		includeDefaults:  cfg.Build.UseDefaultPreprocessors,
		builtinPreprocessors: map[string]func(*models.Book) error{
			"index": func(book *models.Book) error {
				idx := index.NewIndexPreprocessor()
				return idx.Process(book)
			},
		},
	}
}

// SetVerbose enables verbose output
func (r *Runner) SetVerbose(verbose bool) {
	r.verbose = verbose
}

// SetDisableExternals disables external preprocessor execution
func (r *Runner) SetDisableExternals(disable bool) {
	r.disableExternals = disable
}

// SetIncludeDefaults controls whether to include default preprocessors
func (r *Runner) SetIncludeDefaults(include bool) {
	r.includeDefaults = include
}

// Run executes the preprocessor pipeline on a book
func (r *Runner) Run(book *models.Book) error {
	// Get configured preprocessors
	configuredPreprocessors := r.cfg.GetPreprocessorConfigs()

	// Build the execution order
	customPreps := make(map[string]struct {
		Before []string
		After  []string
	})

	for name, ppCfg := range configuredPreprocessors {
		customPreps[name] = struct {
			Before []string
			After  []string
		}{
			Before: ppCfg.Before,
			After:  ppCfg.After,
		}
	}

	// Resolve order
	builtins := GetBuiltinPreprocessors()
	orderedNames, err := ResolvePreprocessorOrder(builtins, customPreps, r.includeDefaults)
	if err != nil {
		return fmt.Errorf("failed to resolve preprocessor order: %w", err)
	}

	if r.verbose {
		fmt.Printf("Preprocessor execution order: %v\n", orderedNames)
	}

	// Execute each preprocessor in order
	for _, name := range orderedNames {
		// Check if it's a built-in
		if isBuiltinPreprocessor(name) {
			if r.verbose {
				fmt.Printf("Running preprocessor: %s (built-in)\n", name)
			}

			// Run built-in
			if fn, ok := r.builtinPreprocessors[name]; ok {
				if err := fn(book); err != nil {
					return fmt.Errorf("preprocessor '%s' failed: %w", name, err)
				}
			}
		} else {
			// External preprocessor
			if r.disableExternals {
				if r.verbose {
					fmt.Printf("Skipping preprocessor: %s (external disabled)\n", name)
				}
				continue
			}

			ppCfg, ok := configuredPreprocessors[name]
			if !ok {
				log.Printf("Warning: preprocessor '%s' not found in config\n", name)
				continue
			}

			if r.verbose {
				if ppCfg.Command != "" {
					fmt.Printf("Running preprocessor: %s (external): %s\n", name, ppCfg.Command)
				} else {
					fmt.Printf("Running preprocessor: %s (external): geopub-%s\n", name, name)
				}
			}

			// Check renderer filter
			if len(ppCfg.Renderers) > 0 {
				found := false
				for _, rend := range ppCfg.Renderers {
					if rend == r.renderer {
						found = true
						break
					}
				}
				if !found {
					if r.verbose {
						fmt.Printf("  (skipped - renderer '%s' not in renderers list)\n", r.renderer)
					}
					continue
				}
			}

			// Run external
			ep := &ExternalPreprocessor{
				Name:       name,
				Command:    ppCfg.Command,
				Renderers:  ppCfg.Renderers,
				Book:       book,
				Config:     r.cfg,
				Renderer:   r.renderer,
				ExtraProps: ppCfg.Extra,
			}

			if err := ep.RunExternal(); err != nil {
				return fmt.Errorf("preprocessor '%s' failed: %w", name, err)
			}
		}
	}

	return nil
}

// GetExecutionOrder returns the preprocessors that will be executed (for testing/inspection)
func (r *Runner) GetExecutionOrder() ([]string, error) {
	configuredPreprocessors := r.cfg.GetPreprocessorConfigs()

	customPreps := make(map[string]struct {
		Before []string
		After  []string
	})

	for name, ppCfg := range configuredPreprocessors {
		customPreps[name] = struct {
			Before []string
			After  []string
		}{
			Before: ppCfg.Before,
			After:  ppCfg.After,
		}
	}

	builtins := GetBuiltinPreprocessors()
	return ResolvePreprocessorOrder(builtins, customPreps, r.includeDefaults)
}
