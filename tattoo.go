package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"time"
	"github.com/salviati/tattoo/webapp"
)

var startUpTime int64

var useFCGI = flag.Bool("fcgi", false, "Use FastCGI")

func LoadTheme(app *webapp.App, themeName string) error {
	cfg := GetConfig()
	app.Log("Use Theme", themeName)
	if err := LoadThemeTemplates(themeName); err != nil {
		return err
	}
	themeURL := path.Join(cfg.Path, "theme", themeName)
	themeStaticURL := path.Join(cfg.Path, "theme", themeName, "static")
	TattooDB.SetVar("ThemeURL", themeURL)
	TattooDB.SetVar("ThemeStaticURL", themeStaticURL)
	return nil
}

func main() {
	flag.Parse()
	if err := GetConfig().Load(); err != nil {
		fmt.Println("Failed to load configure file")
		return
	}
	cfg := GetConfig()
	startUpTime = time.Now().Unix()
	rootPath, _ := os.Getwd()
	rootURL := path.Join(cfg.Path, "/")
	systemStaticPath := path.Join(rootPath, "/sys/static")
	systemStaticURL := path.Join(cfg.Path, "/sys/static")

	themePath := path.Join(rootPath, "theme")
	themeURL := path.Join(cfg.Path, "/theme")

	app := webapp.App{}
	app.Log("App Starts", "OK")
	app.SetStaticPath(systemStaticURL, systemStaticPath)
	app.SetStaticPath(themeURL, themePath)
	app.SetHandler(rootURL, HandleRoot)

	// Load DB
	app.Log("Tattoo DB", "Load DB")
	TattooDB.Load(&app)

	TattooDB.SetVar("RootURL", rootURL)
	TattooDB.SetVar("SystemStaticURL", systemStaticURL)

	// load templates
	if err := LoadSystemTemplates(); err != nil {
		app.Log("Error", fmt.Sprintf("Failed to load system templates: %v", err))
		return
	}
	if err := LoadTheme(&app, GetConfig().ThemeName); err != nil {
		app.Log("Error", fmt.Sprintf("Failed to load theme: %v", err))
	}

	// Start Server.
	if *useFCGI {
		log.Printf("Server Starts(FastCGI): Listen on port %d\n", GetConfig().Port)
		app.RunCGI(GetConfig().Port)
	} else {
		log.Printf("Server Starts: Listen on port %d\n", GetConfig().Port)
		app.Run(GetConfig().Port)
	}
}
