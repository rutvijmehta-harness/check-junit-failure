package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	globSetting = "junit_globs"
	globEnv     = "PLUGIN_JUNIT_GLOBS"
)

func main() {
	app := &cli.App{
		Name:   "harness-check-junit-failure",
		Usage:  "Harness plugin to determine test failures",
		Action: run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    globSetting,
				EnvVars: []string{globEnv},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	p := Plugin{
		GlobPaths: c.String(globSetting),
	}
	return p.Exec()
}
