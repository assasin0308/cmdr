/*
 * Copyright © 2019 Hedzr Yeh.
 */

package demo

import (
	"github.com/hedzr/cmdr"
	"github.com/hedzr/cmdr/examples/demo/svr"
	"github.com/hedzr/cmdr/plugin/daemon"
	"github.com/hedzr/log"
)

// Entry is app main entry
func Entry() {

	// logrus.SetLevel(logrus.DebugLevel)
	// logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	if err := cmdr.Exec(rootCmd,
		// To disable internal commands and flags, uncomment the following codes
		// cmdr.WithBuiltinCommands(false, false, false, false, false),
		daemon.WithDaemon(svr.NewDaemon(), nil, nil, nil),

		cmdr.WithLogx(log.NewStdLoggerWith(log.DebugLevel)),

		cmdr.WithHelpTabStop(40),
	); err != nil {
		log.Fatalf("error: %+v", err)
	}

}
