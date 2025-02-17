package builtins

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/akawashiro/zenlog/zenlog/builtins/commandline"
	"github.com/akawashiro/zenlog/zenlog/builtins/history"
	"github.com/akawashiro/zenlog/zenlog/config"
	"github.com/akawashiro/zenlog/zenlog/envs"
	"github.com/akawashiro/zenlog/zenlog/logger"
	"github.com/akawashiro/zenlog/zenlog/util"
)

// InZenlog true if the current process is running in a zenlog session.
func InZenlog() bool {
	sig := util.Tty() + ":" + logger.Signature()
	util.Debugf("Signature=%s", sig)
	return sig == os.Getenv(envs.ZenlogSignature)
}

// FailIfInZenlog quites the current process with an error code with an error message if it's running in a zenlog session.
func FailIfInZenlog() {
	if InZenlog() {
		util.Fatalf("Already in zenlog.")
	}
}

// FailUnlessInZenlog quites the current process with an error code with an error message unless it's running in a zenlog session.
func FailUnlessInZenlog() {
	if !InZenlog() {
		util.Fatalf("Not in zenlog.")
	}
}

func copyStdinToFile(file string) {
	out, err := os.OpenFile(file, os.O_WRONLY, 0)
	util.Check(err, "Unable to open "+file)
	io.Copy(out, os.Stdin)
}

// WriteToLogger read from STDIN and writes to the current logger. Implies FailUnlessInZenlog().
func WriteToLogger() {
	FailUnlessInZenlog()
	copyStdinToFile(os.Getenv(envs.ZenlogLoggerIn))
}

// WriteToOuter read from STDIN and writes to the console, without logging. Implies FailUnlessInZenlog().
func WriteToOuter() {
	FailUnlessInZenlog()
	file := os.Getenv(envs.ZenlogOuterTty)
	out, err := os.OpenFile(file, os.O_WRONLY, 0)
	util.Check(err, "Unable to open "+file)

	in := bufio.NewReader(os.Stdin)

	crlf := make([]byte, 2)
	crlf[0] = '\r'
	crlf[1] = '\n'

	for {
		line, err := in.ReadBytes('\n')
		if line != nil {
			line = bytes.TrimRight(line, "\r\n")
			out.Write(line)
			out.Write(crlf)
		}
		if err != nil {
			break
		}
	}
}

// OuterTty prints the outer TTY device filename. Implies FailUnlessInZenlog().
func OuterTty() {
	FailUnlessInZenlog()
	fmt.Println(os.Getenv(envs.ZenlogOuterTty))
}

// LoggerPipe prints named pile filename to the logger. Implies FailUnlessInZenlog().
func LoggerPipe() {
	FailUnlessInZenlog()
	fmt.Println(os.Getenv(envs.ZenlogLoggerIn))
}

func checkBinUpdate() {
	if strconv.FormatInt(util.ZenlogBinCtime().Unix(), 10) == os.Getenv(envs.ZenlogBinCtime) {
		util.ExitSuccess()
	}
	util.Say("Zenlog binary updated. Run \"zenlog_restart\" (or \"exit 13\") to restart a zenlog session.")
	util.ExitFailure()
}

// MaybeRunBuiltin runs a builtin command if a given command is a builtin subcommand.
func MaybeRunBuiltin(command string, args []string) {
	switch strings.Replace(command, "_", "-", -1) {
	case "in-zenlog":
		util.Exit(InZenlog())

	case "zenlog-bin":
		fmt.Println(util.FindZenlogBin())
		util.ExitSuccess()

	case "zenlog-src-top":
		fmt.Println(config.ZenlogSrcTopDir())
		util.ExitSuccess()

	case "temp-dir":
		fmt.Println(config.InitConfigForCommands().TempDir)
		util.ExitSuccess()

	case "fail-if-in-zenlog":
		FailIfInZenlog()

	case "fail-unless-in-zenlog":
		FailUnlessInZenlog()

	case "write-to-logger":
		FailUnlessInZenlog()
		WriteToLogger()

	case "write-to-outer":
		FailUnlessInZenlog()
		WriteToOuter()

	case "outer-tty":
		FailUnlessInZenlog()
		OuterTty()

	case "logger-pipe":
		FailUnlessInZenlog()
		LoggerPipe()

	case "history":
		FailUnlessInZenlog()
		history.AllHistoryCommand(args)

	case "current-log":
		FailUnlessInZenlog()
		history.CurrentLogCommand(args)

	case "last-log":
		FailUnlessInZenlog()
		history.LastLogCommand(args)

	case "insert-log-bash":
		FailUnlessInZenlog()
		commandline.InsertLogBash(args)

	case "insert-log-zsh":
		FailUnlessInZenlog()
		commandline.InsertLogZsh(args)

	case "all-commands":
		AllCommandsAndLogCommand(args)

	case "check-bin-update":
		FailUnlessInZenlog()
		checkBinUpdate()

	case "start-command":
		FailUnlessInZenlog()
		startCommand(args)

	case "stop-log", "end-command":
		FailUnlessInZenlog()
		endCommand(args)

	case "list-logs":
		listLogsCommand(args)

	default:
		return
	}
	util.ExitSuccess()
}
