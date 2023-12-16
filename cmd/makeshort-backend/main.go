package main

import (
	_ "backend/docs"
	"backend/internal/config"
	"backend/internal/pkg/app"
)

// @title                        URL Shortener App API
// @version                      0.1
// @description                  API Server for URL Shortener Application
// @host                         localhost:8081
// @BasePath                     /api
// @securityDefinitions.apikey   AccessToken
// @in                           header
// @name                         Authorization
func main() {
	cfg := config.MustLoad()
	a := app.New(cfg)

	a.Run()
}
