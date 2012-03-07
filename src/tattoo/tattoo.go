package main

import (
    "webapp"
    "fmt"
    "log"
    "time"
    "os"
    "path"
    "flag"
    )

var startUpTime int64

var useFCGI = flag.Bool("fcgi", false, "Use FastCGI")

func main() {
    flag.Parse()
    if err := GetConfig().Load(); err != nil {
        fmt.Println("Failed to load configure file")
        return
    }
    startUpTime = time.Now().Unix()
    rootPath, _ := os.Getwd()
    staticPath := path.Join(rootPath, "static")

    app := webapp.App{}
    app.SetStaticPath("/static/", staticPath)
    app.SetHandler("/", HandleRoot)
	// load all templates
	InitTemplates()
    // Load DB
    app.Log("Tattoo DB", "Load DB")
    TattooDB.Load(&app)
    // Start Server.
    if *useFCGI {
        log.Printf("Server Starts(FastCGI): Listen on port %d\n", GetConfig().Port)
        app.RunCGI(GetConfig().Port)
    } else {
        log.Printf("Server Starts: Listen on port %d\n", GetConfig().Port)
        app.Run(GetConfig().Port)
    }
}
