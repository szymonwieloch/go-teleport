package main

import (
	"fmt"
	"os"
	"time"
)

// Colors for printing to console

const colorReset = "\033[0m"
const colorRed = "\033[31m"
const colorGreen = "\033[32m"
const colorCyan = "\033[36m"

// Prints one line of log
func printLog(log string, timestamp time.Time, stderr bool) {
	timeStr := timestamp.Local().Format(time.DateTime)
	output := os.Stdout
	color := colorGreen
	if stderr {
		output = os.Stderr
		color = colorRed
	}

	fmt.Fprintf(output, "%s%s: %s%s%s\n", colorCyan, timeStr, color, log, colorReset)
}

// Prints error and exits
func fatalError(err error, msg string) {
	fmt.Fprintf(os.Stderr, msg+": %v\n", err)
	os.Exit(1)
}
