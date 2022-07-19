package main

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
)

func genLayout(channel string) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()

		if v, err := g.SetView("chat", 0, 0, maxX, maxY-2, gocui.TOP); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}

			v.Autoscroll = true
			v.Wrap = true
			v.Frame = false
		}

		if v, err := g.SetView("channel", 0, maxY-2, len(channel)+2, maxY, gocui.TOP); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}

			v.Frame = false
			v.FgColor = gocui.ColorGreen

			fmt.Fprint(v, channel+":")
		}

		if v, err := g.SetView("entry", len(channel)+2, maxY-2, maxX, maxY, gocui.TOP); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}

			v.Frame = false
			v.Editable = true
			v.Wrap = true

			if _, err := g.SetCurrentView("entry"); err != nil {
				return err
			}

			g.Cursor = true
		}

		return nil
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func entrySwitch(g *gocui.Gui, v *gocui.View) error {
	if _, err := g.SetCurrentView("entry"); err != nil {
		return err
	}

	g.Cursor = true

	return nil
}

func chatSwitch(g *gocui.Gui, v *gocui.View) error {
	if _, err := g.SetCurrentView("chat"); err != nil {
		return err
	}

	g.Cursor = false

	return nil
}

func entryClear(g *gocui.Gui, v *gocui.View) error {
	v.Clear()
	return nil
}
