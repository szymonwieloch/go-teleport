package main

import "github.com/alexflint/go-arg"

type JobID string

type args struct {
	Address string `arg:"env,required" help:"Address of the server"`
}

func parseArgs() args {
	var result args
	arg.MustParse(&result)

	return result
}
