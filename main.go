package main

import (
	"embed"
	"os"

	"GoT0Emergency/internal/app"
	"GoT0Emergency/internal/infra/db"
	logger "GoT0Emergency/internal/pkg/log"
	"GoT0Emergency/internal/pkg/path"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 1. Initialize Path
	if err := path.Init(); err != nil {
		// Log init hasn't happened yet, so we have to use println or panic
		panic("Failed to init path: " + err.Error())
	}

	// 2. Initialize Log
	if err := logger.Init(); err != nil {
		panic("Failed to init log: " + err.Error())
	}
	logger.Info("Application starting...")

	// 3. Initialize DB
	if err := db.Init(); err != nil {
		logger.Error("Failed to init db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create an instance of the app structure
	application := app.NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "GoT0Emergency",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        application.Startup,
		OnShutdown:       application.Shutdown,
		Bind: []interface{}{
			application,
		},
	})

	if err != nil {
		logger.Error("App run error", "error", err)
	}
}
