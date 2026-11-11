package cmd

import (
	"github.com/urfave/cli/v2"

	"github.com/bowouhuy/bookstore-api/config"
	"github.com/bowouhuy/bookstore-api/internal/consts"
	"github.com/bowouhuy/bookstore-api/internal/server"
	"github.com/bowouhuy/bookstore-api/pkg/logger"
	"github.com/bowouhuy/bookstore-api/pkg/msg"
)

// httpStart is a method to start http server
func (cmd *Command) httpStart() *cli.Command {
	return &cli.Command{
		Name:  "http:start",
		Usage: "A command to start http server",
		Action: func(c *cli.Context) error {
			logger.SetJSONFormatter()
			logger.Setup(logger.Config{
				Environment: cmd.App.Env,
				Level:       cmd.App.LogLevel,
				ServiceName: cmd.App.Name,
			})

			if err := msg.Load(config.Messages); err != nil {
				return err
			}

			httpServer := server.NewHTTPServer()

			logger.Info(logger.MessageFormat("starting %s on port %d", cmd.App.Name, cmd.App.Port), logger.EventName(consts.LogEventNameServiceStarting))

			return httpServer.Run()
		},
	}
}
