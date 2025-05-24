package main

import (
	_ "time/tzdata"

	"github.com/ashep/d5y/internal/app"
	"github.com/ashep/go-app/runner"
)

func main() {
	runner.New(app.New, app.Config{}).
		WithConsoleLogWriter().
		WithDefaultHTTPLogWriter().
		WithDefaultHTTPServer().
		WithDefaultMetricsHandler().
		Run()
}
