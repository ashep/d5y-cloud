package main

import (
	_ "time/tzdata"

	"github.com/ashep/d5y/internal/app"
	"github.com/ashep/go-apprun/apprunner"
)

func main() {
	apprunner.New(app.Config{}, app.New).
		WithDefaultHTPServer().
		WithMetricsHandler().
		Run()
}
