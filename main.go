package main

import (
	"fmt"
	"os"
	_ "time/tzdata"

	"github.com/ashep/d5y/internal/app"
	"github.com/ashep/go-app/runner"
)

func main() {
	res := runner.New(app.Run).
		AddConsoleLogWriter().
		AddHTTPLogWriter().
		LoadEnvConfig().
		LoadConfigFile("config.yml").
		Run()

	if res != nil {
		fmt.Println(res.Error())
		os.Exit(1)
	}
}
