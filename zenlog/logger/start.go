package logger

import (
	"github.com/akawashiro/go-common/src/utils"
	"github.com/akawashiro/zenlog/zenlog/config"
	"github.com/akawashiro/zenlog/zenlog/logfiles"
	"github.com/akawashiro/zenlog/zenlog/util"
	"strings"
	"time"
)

type StartRequest struct {
	Command   logfiles.Command
	LogFiles  logfiles.LogFiles
	StartTime time.Time
}

func StartCommand(envs string, commandLineArray []string, clock utils.Clock) {
	config := config.InitConfigForCommands()

	commandLine := strings.Join(commandLineArray, " ")
	command := logfiles.ParseCommandLine(config, commandLine)

	// Open the log file.
	now := util.GetInjectedNow(clock)
	logFiles := logfiles.CreateAndOpenLogFiles(config, now, command)
	defer logFiles.Close()

	logFiles.WriteEnv(command, envs, now)
	logFiles.Close() // We want to close it before Dump().

	// Send the start request to the logger.
	req := StartRequest{*command, logFiles, now}
	util.Dump("startRequest=", req)

	MustSendToLogger(config, utils.StringSlice(CommandStartCommand, util.MustMarshal(req)))
}
