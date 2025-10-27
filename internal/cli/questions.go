package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// FillInitOptionsInteractive prompts the user to confirm or override defaults.
// If stdin is not interactive, it will keep the provided defaults.
func FillInitOptionsInteractive(opts *InitOptions) {
	reader := bufio.NewReader(os.Stdin)

	// Name (directory)
	fmt.Printf("Directory name [%s]: ", opts.Name)
	if s, _ := reader.ReadString('\n'); strings.TrimSpace(s) != "" {
		opts.Name = strings.TrimSpace(s)
	}

	// Title
	defTitle := opts.Title
	if defTitle == "" {
		defTitle = opts.Name
	}
	fmt.Printf("Book title [%s]: ", defTitle)
	if s, _ := reader.ReadString('\n'); strings.TrimSpace(s) != "" {
		opts.Title = strings.TrimSpace(s)
	} else if opts.Title == "" {
		opts.Title = defTitle
	}

	// SrcDir
	fmt.Printf("Source directory [%s]: ", opts.SrcDir)
	if s, _ := reader.ReadString('\n'); strings.TrimSpace(s) != "" {
		opts.SrcDir = strings.TrimSpace(s)
	}

	// BuildDir
	fmt.Printf("Build directory [%s]: ", opts.BuildDir)
	if s, _ := reader.ReadString('\n'); strings.TrimSpace(s) != "" {
		opts.BuildDir = strings.TrimSpace(s)
	}

	// CreateMissing
	defCreate := "n"
	if opts.CreateMissing {
		defCreate = "y"
	}
	fmt.Printf("Create missing chapter files on build? (y/N) [%s]: ", defCreate)
	if s, _ := reader.ReadString('\n'); strings.TrimSpace(s) != "" {
		v := strings.ToLower(strings.TrimSpace(s))
		opts.CreateMissing = v == "y" || v == "yes"
	}
}
