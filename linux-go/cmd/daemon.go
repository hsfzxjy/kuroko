package cmd

import (
	"context"
	"kuroko-linux/internal"
	"kuroko-linux/internal/logging"
	"kuroko-linux/services"
	"kuroko-linux/util"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/sevlyar/go-daemon"
)

type startCommand struct{}
type stopCommand struct{}
type restartCommand struct{}

func getDaemonContext() (ctx daemon.Context) {
	appDir := internal.APP_DIR
	logDir := path.Join(appDir, "log")
	os.MkdirAll(logDir, 0755)
	ctx = daemon.Context{
		PidFileName: path.Join(appDir, "daemon.pid"),
		PidFilePerm: 0644,
		LogFileName: path.Join(logDir, "daemon.log"),
		LogFilePerm: 0644,
		Umask:       027,
	}
	return
}

func (x *startCommand) Execute(args []string) error {
	ctx := getDaemonContext()

	if !daemon.WasReborn() {
		logger.Info("Starting daemon...")

	}
	d, err := ctx.Reborn()
	if err != nil {
		logger.Errorf(err, "Unable to start daemon")
		os.Exit(1)
	}
	if d != nil {
		return nil
	}
	defer ctx.Release()

	logging.SetupLog(ctx.LogFileName)

	logger.Info("------ DAEMON STARTED ------")
	services.Manager.ServeAll()
	logger.Info("------ DAEMON STOPPED ------")

	return nil
}

func (x *stopCommand) Execute(args []string) (err error) {
	ctx := getDaemonContext()

	logger.Info("Stopping daemon...")

	d, err := ctx.Search()

	if errors.Is(err, os.ErrNotExist) {
		logger.Info("Daemon not started")
		goto CLEAR_RET
	} else if err != nil {

		logger.Errorf(err, "Unable send signal to the daemon")
		goto CLEAR_RET
	}

	if err = d.Signal(syscall.Signal(0)); errors.Is(err, os.ErrProcessDone) {
		logger.Info("Daemon has gone, removing PID file...")
		if err = os.Remove(ctx.PidFileName); err != nil {
			logger.Errorf(err, "Error on removing PID file")
			os.Exit(1)
		}
		goto CLEAR_RET
	}

	if err = daemon.SendCommands(d); err != nil {
		logger.Errorf(err, "Error on signaling daemon")
		os.Exit(1)
	}

	{
		c, stop := context.WithTimeout(context.Background(), 3*time.Second)
		defer stop()
		ok := util.EnsureCondition(c, 10*time.Millisecond, func() bool {
			_, err := os.Stat(ctx.PidFileName)
			return errors.Is(err, os.ErrNotExist)
		})
		if !ok {
			logger.Warn("Daemon takes too long to stop")
			os.Exit(1)
		}
	}

	return

CLEAR_RET:
	return nil
}

func (*restartCommand) Execute(args []string) (err error) {
	if !daemon.WasReborn() {
		(&stopCommand{}).Execute(args)
	}
	return (&startCommand{}).Execute(args)
}

func init() {
	var stopFlag bool = true
	daemon.AddCommand(daemon.BoolFlag(&stopFlag), syscall.SIGTERM, nil)

	CLIParser.AddCommand("start", "Start daemon", "", &startCommand{})
	CLIParser.AddCommand("stop", "Stop daemon", "", &stopCommand{})
	CLIParser.AddCommand("restart", "Restart daemon", "", &restartCommand{})
}
