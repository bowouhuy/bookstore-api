package cmd

import "github.com/urfave/cli/v2"

// NewCLI is a constructor
func NewCLI() *cli.App {
	c := cli.NewApp()
	c.Name = "bookstore-api"
	c.Usage = "Bookstore REST API"

	return c
}
