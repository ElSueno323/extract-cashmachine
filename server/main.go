package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"cashmachine/internal/config"
	"cashmachine/internal/files"
	"cashmachine/internal/httpapi"
)

//go:embed static
var staticFS embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	api := httpapi.New(cfg, files.New())
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/sites", api.Sites)
	mux.HandleFunc("POST /api/export", api.Export)
	mux.HandleFunc("GET /api/download/{token}", api.Download)
	mux.Handle("/", staticHandler())

	log.Printf("grand livre disponible sur http://%s (%d caisses)", cfg.ListenAddr, len(cfg.Sites))
	go func() {
		time.Sleep(400 * time.Millisecond)
		openBrowser("http://" + cfg.ListenAddr)
	}()
	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
		fmt.Fprintln(os.Stderr, "serveur arrêté")
		os.Exit(1)
	}
}

func openBrowser(address string) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", address).Start()
	case "darwin":
		_ = exec.Command("open", address).Start()
	default:
		_ = exec.Command("xdg-open", address).Start()
	}
}

func staticHandler() http.Handler {
	disk := filepath.Join("server", "static")
	if info, err := os.Stat(filepath.Join(disk, "index.html")); err == nil && !info.IsDir() {
		return http.FileServer(http.Dir(disk))
	}
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return http.NotFoundHandler()
	}
	return http.FileServer(http.FS(sub))
}
