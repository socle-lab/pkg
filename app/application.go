package app

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/socle-lab/core"
)

type Application struct {
	Core *core.Core
	wg   sync.WaitGroup
}

func (a *Application) shutdown() {
	a.wg.Wait()
}

func (app *Application) ListenForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit

	app.Core.Log.InfoLog.Println("Received signal", s.String())
	app.shutdown()

	os.Exit(0)
}

func (a *Application) Log(tag string, args ...any) {
	switch tag {
	case "info":
		a.Core.Log.InfoLog.Println(args...)
	case "error":
		a.Core.Log.ErrorLog.Println(args...)
	default:
		a.Core.Log.InfoLog.Println(args...)
	}
}
