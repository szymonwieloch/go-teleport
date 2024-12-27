// Definitions of structures that map to the command line arguments
package main

import "github.com/alexflint/go-arg"

type JobID string

type listCmd struct {
}

type startCmd struct {
	Command []string `arg:"positional,required" help:"Command to run"`
}

type stopCmd struct {
	JobID JobID `arg:"positional,required" help:"Job ID to stop"`
}

type logCmd struct {
	JobID JobID `arg:"positional,required" help:"Job ID to show logs"`
}

type statusCmd struct {
	JobID JobID `arg:"positional,required" help:"Job ID to show status"`
}

type args struct {
	Address string     `arg:"env,required" help:"Address of the server"`
	Start   *startCmd  `arg:"subcommand:start" help:"Starts a new remote job"`
	Stop    *stopCmd   `arg:"subcommand:stop" help:"Stops a remote job"`
	List    *listCmd   `arg:"subcommand:list" help:"Lists all remote job"`
	Log     *logCmd    `arg:"subcommand:log" help:"Shows logs of the remote job"`
	Status  *statusCmd `arg:"subcommand:status" help:"Prints status of the remote job"`
}

// Parses command line arguments
func parseArgs() args {
	var result args
	p := arg.MustParse(&result)
	if result.Start == nil && result.Stop == nil && result.List == nil && result.Log == nil && result.Status == nil {
		p.Fail("Please choose subcommand")
	}
	return result
}
