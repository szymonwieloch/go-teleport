// Definitions of structures that map to the command line arguments
package main

import "github.com/alexflint/go-arg"

type Args struct {
	Address  string `arg:"env,required" help:"Address of the server"`
	AuthKey  string `arg:"env" help:"Path to a authentication key, if desired"`
	AuthCert string `arg:"env" help:"Path to a authentication certificate, if desired"`
	Secret   string `arg:"env" help:"A secret for authentication, if desired"`
	Limits   bool   `help:"Enable cgroup limits"`
}

// Either all are empty or all are set
func definedTogether(a ...string) bool {
	if len(a) < 1 {
		return true
	}
	first := a[0] == ""
	for _, v := range a {
		if first != (v == "") {
			return false
		}
	}
	return true
}

// Parses command line arguments
func parseArgs() Args {
	var result Args
	parser := arg.MustParse(&result)
	if !definedTogether(result.Secret, result.AuthCert, result.AuthKey) {
		parser.Fail("Authentication key, certificate and secret need to be provided together")
	}

	return result
}
