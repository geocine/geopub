package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/geocine/geopub/internal/config"
	"github.com/geocine/geopub/internal/models"
)

// ExternalPreprocessor represents an external preprocessor that runs as a separate command
type ExternalPreprocessor struct {
	Name       string
	Command    string
	Renderers  []string
	Book       *models.Book
	Config     *config.Config
	Renderer   string
	ExtraProps map[string]interface{}
}

// RunExternal executes an external preprocessor and returns the modified book
func (ep *ExternalPreprocessor) RunExternal() error {
	// Resolve command if not specified
	command := ep.Command
	if command == "" {
		command = fmt.Sprintf("geopub-%s", ep.Name)
	}

	// Check if preprocessor should run for this renderer
	if len(ep.Renderers) > 0 {
		found := false
		for _, r := range ep.Renderers {
			if r == ep.Renderer {
				found = true
				break
			}
		}
		if !found {
			// Skip this preprocessor for this renderer
			return nil
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the preprocessor context
	ppCtx := NewPreprocessorContext(ep.Book, ep.Config, ep.Renderer)

	// Marshal to JSON
	inputJSON, err := json.Marshal(ppCtx)
	if err != nil {
		return fmt.Errorf("failed to marshal preprocessor context: %w", err)
	}

	// Parse command (handle shell commands like "node script.js")
	parts := strings.Fields(command)
	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.CommandContext(ctx, parts[0])
	} else {
		cmd = exec.CommandContext(ctx, parts[0], parts[1:]...)
	}

	// Set up stdin, stdout, stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdin = bytes.NewReader(inputJSON)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	if err := cmd.Run(); err != nil {
		stderrMsg := stderr.String()
		if stderrMsg != "" {
			return fmt.Errorf("preprocessor '%s' failed: %w\nstderr: %s", ep.Name, err, stderrMsg)
		}
		return fmt.Errorf("preprocessor '%s' failed: %w", ep.Name, err)
	}

	// Parse output JSON
	var outputCtx PreprocessorContext
	if err := json.Unmarshal(stdout.Bytes(), &outputCtx); err != nil {
		return fmt.Errorf("preprocessor '%s' returned invalid JSON: %w\noutput: %s", ep.Name, err, stdout.String())
	}

	// Apply mutations to book
	if err := JsonToBook(outputCtx.Book, ep.Book); err != nil {
		return fmt.Errorf("failed to apply preprocessor mutations: %w", err)
	}

	return nil
}

// ResolveCommand resolves the command for a preprocessor
// If command is specified, uses it as-is (may contain spaces for shell commands)
// Otherwise, resolves to "geopub-<name>" on PATH
func ResolveCommand(name, command string) (string, error) {
	if command != "" {
		return command, nil
	}

	defaultCommand := fmt.Sprintf("geopub-%s", name)

	// Try to find on PATH
	path, err := exec.LookPath(defaultCommand)
	if err == nil {
		return path, nil
	}

	// If not found but command exists, try it anyway (will fail at runtime with clear error)
	return defaultCommand, nil
}

// ValidateCommandExists checks if a command can be resolved
func ValidateCommandExists(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	_, err := exec.LookPath(parts[0])
	return err
}

// ExecutePreprocessor runs a preprocessor (built-in or external) on a book
// If the preprocessor is external, it runs as a subprocess
// If it's built-in (like "index"), it runs in-process
func ExecutePreprocessor(name string, ppCfg *config.PreprocessorConfig, book *models.Book, cfg *config.Config, renderer string, verbose bool) error {
	// Check if it's a built-in preprocessor
	if isBuiltinPreprocessor(name) {
		if verbose {
			fmt.Printf("Running preprocessor: %s (built-in)\n", name)
		}
		// Let the built-in preprocessor pipeline handle it
		return nil
	}

	// External preprocessor
	if verbose {
		if ppCfg.Command != "" {
			fmt.Printf("Running preprocessor: %s (external): %s\n", name, ppCfg.Command)
		} else {
			fmt.Printf("Running preprocessor: %s (external): geopub-%s\n", name, name)
		}
	}

	ep := &ExternalPreprocessor{
		Name:       name,
		Command:    ppCfg.Command,
		Renderers:  ppCfg.Renderers,
		Book:       book,
		Config:     cfg,
		Renderer:   renderer,
		ExtraProps: ppCfg.Extra,
	}

	return ep.RunExternal()
}

// isBuiltinPreprocessor checks if a preprocessor is built-in (runs in-process)
func isBuiltinPreprocessor(name string) bool {
	// Built-in preprocessors: index and frontmatter
	builtins := map[string]bool{
		"index":       true,
		"frontmatter": true,
	}
	return builtins[name]
}

// GetBuiltinPreprocessors returns the list of built-in preprocessor names
// Only "index" is in the default list; frontmatter must be explicitly enabled
func GetBuiltinPreprocessors() []string {
	return []string{"index"}
}

// PrepareWorkingDirectory resolves paths in commands relative to the book root
// For example, "node preprocessors/example.js" becomes "node <bookroot>/preprocessors/example.js"
func PrepareWorkingDirectory(command string, bookRoot string) string {
	parts := strings.Fields(command)
	if len(parts) < 2 {
		return command
	}

	// Resolve paths that don't start with / or a drive letter (for Windows)
	// and aren't known commands
	resolved := []string{parts[0]}
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		// Check if it looks like a path (contains / or \)
		if strings.ContainsAny(part, `/\`) {
			// Check if it's relative
			if !filepath.IsAbs(part) {
				// Make it relative to bookRoot
				part = filepath.Join(bookRoot, part)
			}
		}
		resolved = append(resolved, part)
	}

	return strings.Join(resolved, " ")
}

// GetCommandSearchPaths returns directories to search for preprocessor commands
func GetCommandSearchPaths(bookRoot string) []string {
	paths := []string{}

	// Add common locations
	paths = append(paths, filepath.Join(bookRoot, "bin"))
	paths = append(paths, filepath.Join(bookRoot, "preprocessors"))

	// Add system PATH
	if pathEnv := os.Getenv("PATH"); pathEnv != "" {
		paths = append(paths, strings.Split(pathEnv, string(os.PathListSeparator))...)
	}

	return paths
}
