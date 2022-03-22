package main

import (
	cli "github.com/urfave/cli/v2"
)

var getShardCmd = &cli.Command{
	Name: "get",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "key",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		//key := cctx.String("key")

		return nil
	},
}
