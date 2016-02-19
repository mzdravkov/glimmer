package main

import (
	"github.com/tucnak/climax"
)

func createProgram() *climax.Application {
	program := climax.New("glimmer")

	program.Brief = "Glimmer is a tool that visualises the communication between goroutines"
	program.Version = "0.0.1"

	on := climax.Command{
		Name:  "on",
		Brief: "set the path to the project, onto which you want to run glimmer",
		Usage: "/path/to/some/project",
		Help:  "set the path to the project, onto which you want to run glimmer",

		Flags: []climax.Flag{
			{
				Name:     "port",
				Short:    "p",
				Usage:    "--port=4242",
				Variable: true,
			},
		},

		Examples: []climax.Example{
			{Usecase: "."},
		},

		Handle: func(ctx climax.Context) int {
			if len(ctx.Args) != 1 {
				ctx.Log("on command needs exactly one arugment (it's possible that unknown flag was passed)",
					"'glimmer help on' for more information")
				return 1
			}

			run(ctx.Args[0])
			return 0
		},
	}

	program.AddCommand(on)

	return program
}
