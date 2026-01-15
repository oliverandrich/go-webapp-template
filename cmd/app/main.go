// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package main

import (
	"context"
	"log"
	"os"

	"github.com/oliverandrich/go-webapp-template/internal/config"
	"github.com/oliverandrich/go-webapp-template/internal/server"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:   "app",
		Usage:  "Start the web application",
		Flags:  config.Flags(),
		Action: server.Run,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
