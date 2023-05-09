package commands

import (
	"github.com/urfave/cli/v2"
	"globalflow/config"
	"globalflow/globalflow"
	"os"
	"os/signal"
	"syscall"
)

var ServerCommand = &cli.Command{
	Name: "server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "node-name",
		},
		&cli.IntFlag{
			Name: "node-port",
		},
		&cli.StringSliceFlag{
			Name: "node-peers",
		},
		&cli.StringFlag{
			Name: "node-zone",
		},
		&cli.StringFlag{
			Name: "node-region",
		},
		&cli.IntFlag{
			Name: "redis-port",
		},
	},
	Action: func(c *cli.Context) error {
		container := &globalflow.Container{
			Configuration: config.NewConfiguration(),
		}

		if c.String("node-name") != "" {
			container.Configuration.NodeID = c.String("node-name")
		}

		if c.Int("node-port") != 0 {
			container.Configuration.NodePort = c.Int("node-port")
		}

		if c.StringSlice("node-peers") != nil {
			container.Configuration.NodePeers = c.StringSlice("node-peers")
		}

		if c.String("node-zone") != "" {
			container.Configuration.NodeZone = c.String("node-zone")
		}

		if c.String("node-region") != "" {
			container.Configuration.NodeRegion = c.String("node-region")
		}

		if c.Int("redis-port") != 0 {
			container.Configuration.RedisPort = c.Int("redis-port")
		}

		server := globalflow.NewServer(container)

		sigs := make(chan os.Signal, 1)

		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigs
			server.Close()
		}()

		return server.Run(c.Context)
	},
}
