package main

import (
	"github.com/rmrfslashbin/goawsloc/subcmds/loc"

	"github.com/sirupsen/logrus"
)

func main() {
	// Catch errors
	var err error
	defer func() {
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("main crashed")
		}
	}()
	Execute()
}

// Execute the root command
func Execute() error {
	return loc.RootCmd.Execute()
}
