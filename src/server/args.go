// Definitions of structures that map to the command line arguments
package main

import "github.com/alexflint/go-arg"

type args struct {
	Address string `arg:"env,required" help:"Address of the server"`
}

// Parses command line arguments
func parseArgs() args {
	var result args
	arg.MustParse(&result)

	return result
}
