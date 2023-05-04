package commands

import (
	"github.com/urfave/cli/v2"
	"globalflow/globalflow"
)

var ServerCommand = &cli.Command{
	Name: "server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "node-name",
		},
	},
	Action: func(c *cli.Context) error {
		container := &globalflow.Container{
			Configuration: globalflow.NewConfiguration(),
		}

		if c.String("node-name") != "" {
			container.Configuration.NodeID = c.String("node-name")
		}

		server := globalflow.NewServer(container)

		return server.Run(c.Context)
	},
}
