package main

import (
	"github.com/omakoto/go-common/src/utils"
	"github.com/akawashiro/zenlog/zenlog"
	"github.com/akawashiro/zenlog/zenlog/builtins"
	"github.com/akawashiro/zenlog/zenlog/config"
	"github.com/akawashiro/zenlog/zenlog/util"
	"runtime"
)

func restart() {
	util.Say("Restarting zenlog...")
	util.MustExec(utils.StringSlice(util.FindZenlogBin()))
}

func realMain() int {
	command, args := util.GetSubcommand()

	if command == "" {
		config.SetIsLogger(true)
		builtins.FailIfInZenlog()
		status, resurrect := zenlog.StartZenlog(args)
		if resurrect {
			restart()
		}
		return status
	}
	config.SetIsLogger(false)
	builtins.MaybeRunBuiltin(command, args)
	MaybeRunExternalCommand(command, args)

	util.Fatalf("Unknown subcommand: '%s'", command)
	return 0
}

func main() {
	runtime.GOMAXPROCS(1)
	util.RunAndExit(realMain)
}
