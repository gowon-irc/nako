package main

import (
	"fmt"
	"io"
	"time"

	"github.com/awesome-gocui/gocui"
	"github.com/logrusorgru/aurora"
)

func genChatViewLoggerFunc(g *gocui.Gui) func(s string) {
	return func(s string) {
		g.Update(func(g *gocui.Gui) error {
			v, err := g.View("chat")
			if err != nil {
				return err
			}

			fmt.Fprintln(v, s)
			return nil
		})
	}
}

func genWriterLoggerFunc(w io.Writer) func(s string) {
	return func(s string) {
		fmt.Fprintln(w, s)
	}
}

type logger struct {
	loggerFunc func(s string)
}

func (c *logger) Log(s string, tt ...string) {
	var t string

	if len(tt) == 0 {
		now := time.Now()
		t = now.Format("15:04")
	} else {
		t = tt[0]
	}

	ft := aurora.Bold(t).String()
	out := fmt.Sprintf("%s %s", ft, s)

	c.loggerFunc(out)
}

func createLogger(f func(s string)) *logger {
	return &logger{
		loggerFunc: f,
	}
}
