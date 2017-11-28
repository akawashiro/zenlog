package zenlog

import (
	"github.com/kr/pty"
	"github.com/omakoto/zenlog-go/zenlog/config"
	"github.com/omakoto/zenlog-go/zenlog/envs"
	"github.com/omakoto/zenlog-go/zenlog/logger"
	"github.com/omakoto/zenlog-go/zenlog/util"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

func StartZenlog(args []string) bool {
	config := config.InitConfigiForLogger()
	util.Dump("config=", config)

	logger := logger.NewLogger(config)
	defer logger.CleanUp()
	util.Dump("Logger=", logger)

	// Set up signal handler.
	sigch := make(chan os.Signal)
	signal.Notify(sigch, syscall.SIGCHLD, syscall.SIGWINCH)

	// Set up environmental variables.
	logger.ExportEnviron()

	// Create a pty and start the child command.
	util.Debugf("Executing: %s", config.StartCommand)
	c := exec.Command("/bin/sh", "-c", envs.ZENLOG_TTY+"=\"$(tty)\" "+config.StartCommand)
	m, err := pty.Start(c)
	util.Check(err, "Unable to create pty or execute /bin/sh")

	util.PropagateTerminalSize(os.Stdin, m)

	// WG to wait for child exit.
	var wg sync.WaitGroup
	wg.Add(1)

	var childStatus int = -1

	// Signal handler.
	go func() {
		for s := range sigch {
			switch s {
			case syscall.SIGWINCH:
				util.Debugf("Caught SIGWINCH")
				util.PropagateTerminalSize(os.Stdin, m)
			case syscall.SIGCHLD:
				util.Debugf("Caught SIGCHLD")
				ps, err := c.Process.Wait()
				if err != nil {
					util.Fatalf("Wait failed: %s", err)
				} else {
					childStatus = ps.Sys().(syscall.WaitStatus).ExitStatus()
				}
				os.Stdin.Close()
				os.Stdout.Close()
				m.Close()
				wg.Done()
			default:
				util.Debugf("Caught unexpected signal: %+v", s)
			}
		}
	}()

	// Forward the input from stdin to the logger.
	go func() {
		io.Copy(m, os.Stdin)
	}()

	// Read the output, and write to the STDOUT, and also to the pipe.
	go func() {
		buf := make([]byte, 32*1024)

		for {
			nr, er := m.Read(buf)
			if nr > 0 {
				// First, write to stdout.
				nw, ew := os.Stdout.Write(buf[0:nr])
				if ew != nil {
					err = ew
					break
				}
				if nr != nw {
					err = io.ErrShortWrite
					break
				}
				// Then, write to logger.
				nw, ew = logger.ForwardPipe.Write(buf[0:nr])
				if ew != nil {
					err = ew
					break
				}
				if nr != nw {
					err = io.ErrShortWrite
					break
				}
			}
			if er != nil {
				if er != io.EOF {
					err = er
				}
				break
			}
		}
	}()
	// Logger.
	go func() {
		logger.DoLogger()
	}()

	wg.Wait()

	util.Debugf("Child exited with=%d", childStatus)
	return true
}