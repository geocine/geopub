package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/geocine/geopub/internal/cli"
	"github.com/geocine/geopub/internal/config"
	"github.com/geocine/geopub/internal/loader"
	"github.com/geocine/geopub/internal/models"
	"github.com/geocine/geopub/internal/preprocessor/runner"
	"github.com/geocine/geopub/internal/renderer"
)

func main() {
	// Define subcommands
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	buildDir := buildCmd.String("dest-dir", "", "Destination directory for build")
	buildNoExternals := buildCmd.Bool("no-externals", false, "Disable external preprocessors")
	buildVerbose := buildCmd.Bool("verbose", false, "Enable verbose output")

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initName := initCmd.String("name", "", "Book directory name (or pass as positional)")
	initTitle := initCmd.String("title", "", "Book title (defaults to name)")
	initSrc := initCmd.String("src", "src", "Source directory")
	initBuildDir := initCmd.String("build-dir", "book", "Build output directory")
	initCreateMissing := initCmd.Bool("create-missing", false, "Create missing files on build")
	initYes := initCmd.Bool("yes", false, "Skip interactive prompts and use provided/default values")

	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	servePort := serveCmd.Int("port", 3000, "Port to serve on")
	serveHost := serveCmd.String("hostname", "127.0.0.1", "Hostname to bind to")
	serveOpen := serveCmd.Bool("open", false, "Open in browser")
	serveDest := serveCmd.String("dest-dir", "", "Destination directory for build")
	serveNoExternals := serveCmd.Bool("no-externals", false, "Disable external preprocessors")
	serveVerbose := serveCmd.Bool("verbose", false, "Enable verbose output")

	cleanCmd := flag.NewFlagSet("clean", flag.ExitOnError)
	cleanDest := cleanCmd.String("dest-dir", "", "Destination directory to clean")

	if len(os.Args) < 2 {
		fmt.Println("Usage: geopub [command]")
		fmt.Println("Commands:")
		fmt.Println("  build      Build the book")
		fmt.Println("  init       Initialize a new book")
		fmt.Println("  serve      Serve the book")
		fmt.Println("  clean      Clean the build directory")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		buildCmd.Parse(os.Args[2:])
		handleBuild(*buildDir, *buildNoExternals, *buildVerbose)

	case "init":
		initCmd.Parse(os.Args[2:])
		handleInit(initCmd, *initName, *initTitle, *initSrc, *initBuildDir, *initCreateMissing, *initYes)

	case "serve":
		serveCmd.Parse(os.Args[2:])
		handleServe(*serveHost, *servePort, *serveOpen, *serveDest, *serveNoExternals, *serveVerbose)

	case "clean":
		cleanCmd.Parse(os.Args[2:])
		handleClean(*cleanDest)

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func handleBuild(destDir string, noExternals, verbose bool) {
	// Load config
	cfg, err := config.LoadFromFile("book.toml")
	if err != nil {
		log.Printf("Warning: could not load config file: %v. Using defaults.", err)
		cfg = config.NewDefaultConfig()
	}

	// Use config's build directory if destDir not specified
	outDir := destDir
	if outDir == "" {
		outDir = cfg.Build.BuildDir
	}

	// Create book loader
	bl := loader.NewBookLoader(".", cfg)

	// Load book
	book, err := bl.Load()
	if err != nil {
		log.Fatalf("Failed to load book: %v", err)
	}

	// Print book info
	fmt.Printf("Building book: %s\n", cfg.Book.Title)
	fmt.Printf("Chapters loaded: %d\n", len(book.Items))

	// Print chapter information
	for i, item := range book.Items {
		switch v := item.(type) {
		case *models.Chapter:
			fmt.Printf("  [%d] Chapter: %s\n", i, v.Name)
		case *models.Separator:
			fmt.Printf("  [%d] Separator\n", i)
		case *models.PartTitle:
			fmt.Printf("  [%d] Part: %s\n", i, v.Title)
		default:
			fmt.Printf("  [%d] Unknown item\n", i)
		}
	}

	// Run preprocessors
	fmt.Println("Running preprocessors...")
	pipelineRunner := runner.NewRunner(cfg, "html")
	pipelineRunner.SetVerbose(verbose)
	pipelineRunner.SetDisableExternals(noExternals)

	if err := pipelineRunner.Run(book); err != nil {
		log.Fatalf("Failed to run preprocessors: %v", err)
	}

	// Render to HTML
	fmt.Printf("Rendering to: %s\n", outDir)
	r := renderer.NewHtmlRenderer()
	ctx := &renderer.RenderContext{
		Root:                   ".",
		DestDir:                outDir,
		Book:                   book,
		Config:                 cfg,
		SourceDir:              filepath.Join(".", cfg.Book.Src),
		LiveReloadEndpointPath: "", // not serving
		AssetsFS:               embeddedFrontend,
	}

	if err := r.Render(ctx); err != nil {
		log.Fatalf("Failed to render book: %v", err)
	}

	fmt.Printf("Book built successfully to %s!\n", outDir)
}

func handleInit(initCmd *flag.FlagSet, name, title, src, buildDir string, createMissing, yes bool) {
	// Determine name: prefer positional arg if present, then --name, else default
	if name == "" {
		// Positional: first non-flag arg after parsing
		if initCmd.NArg() >= 1 {
			name = initCmd.Arg(0)
		} else {
			name = "my-book"
		}
	}

	fmt.Printf("Initializing new book: %s\n", name)

	opts := cli.InitOptions{
		Name:          name,
		CreateMissing: createMissing,
		SrcDir:        src,
		BuildDir:      buildDir,
		Title:         title,
	}

	if !yes {
		cli.FillInitOptionsInteractive(&opts)
	}

	if err := cli.Init(opts); err != nil {
		log.Fatalf("Failed to initialize book: %v", err)
	}

	fmt.Printf("\nSuccessfully created book in '%s'\n", name)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", name)
	fmt.Println("  geopub build     # build the book")
	fmt.Println("  geopub serve     # serve locally with live reload")
}

// handleServe builds the book, serves it with live reload, and rebuilds on changes.
func handleServe(host string, port int, open bool, destOverride string, noExternals, verbose bool) {
	addr := fmt.Sprintf("%s:%d", host, port)

	// Load config
	cfg, err := config.LoadFromFile("book.toml")
	if err != nil {
		log.Printf("Warning: could not load config file: %v. Using defaults.", err)
		cfg = config.NewDefaultConfig()
	}

	// Determine output directory
	outDir := destOverride
	if outDir == "" {
		outDir = cfg.Build.BuildDir
	}

	// Initial build
	if err := buildWithOptions(outDir, true, "/__livereload", noExternals, verbose); err != nil {
		log.Fatalf("Initial build failed: %v", err)
	}

	// Live reload broker (SSE)
	broker := newSSEBroker()

	// HTTP handlers
	mux := http.NewServeMux()
	// SSE endpoint
	mux.HandleFunc("/__livereload", func(w http.ResponseWriter, r *http.Request) {
		broker.serveSSE(w, r)
	})
	// Static files with 404 fallback
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Clean path and map to file in outDir
		upath := r.URL.Path
		if strings.HasSuffix(upath, "/") {
			upath = upath + "index.html"
		}
		if upath == "/" {
			upath = "/index.html"
		}
		// Prevent path traversal
		upath = filepath.Clean(upath)
		target := filepath.Join(outDir, upath)
		// Ensure target stays within outDir
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(outDir)) {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		if fi, err := os.Stat(target); err == nil && !fi.IsDir() {
			http.ServeFile(w, r, target)
			return
		}
		// Fallback to 404.html
		fourOFour := filepath.Join(outDir, "404.html")
		if _, err := os.Stat(fourOFour); err == nil {
			w.WriteHeader(http.StatusNotFound)
			http.ServeFile(w, r, fourOFour)
			return
		}
		http.NotFound(w, r)
	})

	server := &http.Server{Addr: addr, Handler: mux}

	// Start server
	go func() {
		log.Printf("Serving on http://%s\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Open browser if requested
	if open {
		go func() {
			url := fmt.Sprintf("http://%s", addr)
			time.Sleep(300 * time.Millisecond)
			_ = openBrowser(url)
		}()
	}

	// Watch and rebuild
	watchPaths := []string{"book.toml", cfg.Book.Src}
	watchPaths = append(watchPaths, cfg.Build.ExtraWatchDirs...)
	debounce := 150 * time.Millisecond
	var lastBuild time.Time
	var mu sync.Mutex

	// track mtimes
	lastHash := ""

	for {
		time.Sleep(300 * time.Millisecond)
		hash, err := snapshotModHash(watchPaths)
		if err != nil {
			// Log and continue
			log.Printf("watch error: %v", err)
			continue
		}
		if hash != lastHash && time.Since(lastBuild) > debounce {
			mu.Lock()
			// Recompute within lock to avoid duplicate builds
			hash2, _ := snapshotModHash(watchPaths)
			if hash2 == lastHash {
				mu.Unlock()
				continue
			}
			log.Println("Change detected, rebuilding...")
			if err := buildWithOptions(outDir, true, "/__livereload", noExternals, verbose); err != nil {
				log.Printf("Build failed: %v", err)
			} else {
				lastHash = hash2
				lastBuild = time.Now()
				broker.broadcast("reload")
				log.Println("Rebuilt. Reload signal sent.")
			}
			mu.Unlock()
		}
	}
}

// buildWithOptions loads the book and renders with optional live reload endpoint.
func buildWithOptions(outDir string, serve bool, liveReloadPath string, noExternals, verbose bool) error {
	cfg, err := config.LoadFromFile("book.toml")
	if err != nil {
		cfg = config.NewDefaultConfig()
	}
	bl := loader.NewBookLoader(".", cfg)
	book, err := bl.Load()
	if err != nil {
		return fmt.Errorf("failed to load book: %w", err)
	}

	// Run preprocessors
	pipelineRunner := runner.NewRunner(cfg, "html")
	pipelineRunner.SetVerbose(verbose)
	pipelineRunner.SetDisableExternals(noExternals)
	if err := pipelineRunner.Run(book); err != nil {
		return fmt.Errorf("failed to run preprocessors: %w", err)
	}

	htmlRenderer := renderer.NewHtmlRenderer()
	ctx := &renderer.RenderContext{
		Root:                   ".",
		DestDir:                outDir,
		Book:                   book,
		Config:                 cfg,
		SourceDir:              filepath.Join(".", cfg.Book.Src),
		LiveReloadEndpointPath: "",
		AssetsFS:               embeddedFrontend,
	}
	if serve {
		ctx.LiveReloadEndpointPath = liveReloadPath
	}
	if err := htmlRenderer.Render(ctx); err != nil {
		return fmt.Errorf("render failed: %w", err)
	}
	return nil
}

// snapshotModHash walks provided paths and returns a coarse hash based on mtimes and sizes.
func snapshotModHash(paths []string) (string, error) {
	var b strings.Builder
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			// If path doesn't exist, skip
			continue
		}
		if info.IsDir() {
			filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					return nil
				}
				// Ignore output dir to avoid loops
				if strings.Contains(path, string(os.PathSeparator)+"book"+string(os.PathSeparator)) {
					return nil
				}
				fmt.Fprintf(&b, "%s|%d|%d\n", path, info.ModTime().UnixNano(), info.Size())
				return nil
			})
		} else {
			fmt.Fprintf(&b, "%s|%d|%d\n", p, info.ModTime().UnixNano(), info.Size())
		}
	}
	return b.String(), nil
}

// openBrowser attempts to open the provided URL in a browser.
func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

// SSE broker for simple live reload.
type sseBroker struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

func newSSEBroker() *sseBroker {
	return &sseBroker{clients: make(map[chan string]struct{})}
}

func (b *sseBroker) serveSSE(w http.ResponseWriter, r *http.Request) {
	// Set headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}

	ch := make(chan string, 1)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()

	// Heartbeat to keep connection alive
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	// Initial comment
	fmt.Fprintf(w, ":ok\n\n")
	flusher.Flush()

	// Clean up on exit
	defer func() {
		b.mu.Lock()
		delete(b.clients, ch)
		b.mu.Unlock()
	}()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ":hb\n\n")
			flusher.Flush()
		case msg := <-ch:
			fmt.Fprintf(w, "event: reload\n")
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (b *sseBroker) broadcast(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func handleClean(destOverride string) {
	// Load config
	cfg, err := config.LoadFromFile("book.toml")
	if err != nil {
		log.Printf("Warning: could not load config file: %v. Using defaults.", err)
		cfg = config.NewDefaultConfig()
	}
	// Determine directory to clean
	outDir := destOverride
	if outDir == "" {
		outDir = cfg.Build.BuildDir
	}
	// If it doesn't exist, nothing to do
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		fmt.Printf("Nothing to clean; directory '%s' does not exist.\n", outDir)
		return
	}
	// Summarize contents
	var files, dirs int
	var bytes int64
	filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if path != outDir {
				dirs++
			}
			return nil
		}
		files++
		bytes += info.Size()
		return nil
	})
	// Remove
	if err := os.RemoveAll(outDir); err != nil {
		log.Fatalf("Failed to remove '%s': %v", outDir, err)
	}
	fmt.Printf("Removed %d files, %d directories, %s from '%s'.\n", files, dirs, humanBytes(bytes), outDir)
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	val := float64(n) / float64(div)
	suffix := []string{"KiB", "MiB", "GiB", "TiB"}
	if exp >= len(suffix) {
		return fmt.Sprintf("%.1f PiB", val/float64(unit))
	}
	return fmt.Sprintf("%.1f %s", val, suffix[exp])
}
