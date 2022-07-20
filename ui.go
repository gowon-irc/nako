package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gowon-irc/go-gowon"
)

func genLayout(channels []string) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()

		chatY := maxY

		if len(channels) == 1 {
			chatY = maxY - 2

			if v, err := g.SetView("entry", len(channels[0])+2, maxY-2, maxX, maxY, gocui.TOP); err != nil {
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

			if v, err := g.SetView("channel", 0, maxY-2, len(channels[0])+2, maxY, gocui.TOP); err != nil {
				if !errors.Is(err, gocui.ErrUnknownView) {
					return err
				}

				v.Frame = false
				v.FgColor = gocui.ColorGreen

				fmt.Fprint(v, channels[0]+":")
			}
		}

		if v, err := g.SetView("chat", 0, 0, maxX, chatY, gocui.TOP); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}

			v.Autoscroll = true
			v.Wrap = true
			v.Frame = false
		}

		return nil
	}
}

func chatLogger(s string, g *gocui.Gui, date ...string) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("chat")
		if err != nil {
			return err
		}

		fmt.Fprintln(v, s)

		return nil
	})
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

func genSendMessage(c mqtt.Client, module, topic, channel string) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if v.Buffer() == "" {
			return nil
		}

		m := &gowon.Message{
			Module: module,
			Dest:   channel,
			Msg:    v.Buffer() + " ",
		}

		mj, err := json.Marshal(m)
		if err != nil {
			chatLogger(err.Error(), g)
			return err
		}

		c.Publish(topic, 0, false, mj)
		chatLogger(v.Buffer(), g)
		v.Clear()

		return nil
	}
}
