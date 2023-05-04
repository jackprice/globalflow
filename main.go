package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"globalflow/commands"
	"os"
)

func main() {
	app := &cli.App{
		Name:        "globalflow",
		Description: "Globally distributed Redis",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-format",
				EnvVars: []string{"LOG_FORMAT"},
				Value:   "text",
				Usage:   "Log format to use - text or json",
				Action: func(c *cli.Context, value string) error {
					switch value {
					case "text":
						logrus.SetFormatter(&logrus.TextFormatter{})
					case "json":
						logrus.SetFormatter(&logrus.JSONFormatter{})
					default:
						return cli.Exit("invalid log format", 1)
					}

					return nil
				},
			},
			&cli.StringFlag{
				Name:    "log-level",
				EnvVars: []string{"LOG_LEVEL"},
				Value:   "info",
				Usage:   "Log level to use - debug, info, warn, error, fatal, panic",
				Action: func(c *cli.Context, value string) error {
					level, err := logrus.ParseLevel(value)
					if err != nil {
						return cli.Exit("invalid log level", 1)
					}

					logrus.SetLevel(level)

					return nil
				},
			},
		},
		Commands: []*cli.Command{
			commands.ServerCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
