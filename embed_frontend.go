package main

import (
	"embed"
)

// embeddedFrontend contains the contents of the frontend directory for templates and static assets.
//
//go:embed frontend
var embeddedFrontend embed.FS
