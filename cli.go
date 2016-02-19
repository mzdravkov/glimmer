package main

import (
	"github.com/tucnak/climax"
)

var flags = make(map[string]string)

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
			{
				Name:     "examine-all",
				Short:    "a",
				Variable: false,
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

			if value, ok := ctx.Get("port"); ok {
				flags["port"] = value
			} else {
				flags["port"] = "9610"
			}

			if value, ok := ctx.Get("delay"); ok {
				flags["delay"] = value
			} else {
				flags["delay"] = "1000"
			}

			if _, ok := ctx.Get("examine-all"); ok {
				flags["examine-all"] = "true"
			}

			run(ctx.Args[0], flags)
			return 0
		},
	}

	program.AddCommand(on)

	return program
}
