package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/urfave/cli/v2"
)

// healthcheck is used as the container HEALTHCHECK, since the
// distroless image has no shell or curl.
func (cmd *Command) healthcheck() *cli.Command {
	return &cli.Command{
		Name:  "healthcheck",
		Usage: "Exit with status 0 if the local HTTP server reports healthy",
		Action: func(c *cli.Context) error {
			return checkLiveness(fmt.Sprintf("http://127.0.0.1:%d/liveness", cmd.App.Port))
		},
	}
}

func checkLiveness(url string) error {
	client := http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("liveness request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("liveness returned status %d", resp.StatusCode)
	}

	return nil
}
