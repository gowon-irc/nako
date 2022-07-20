package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/awesome-gocui/gocui"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gowon-irc/go-gowon"
	"github.com/logrusorgru/aurora"
)

func genLayout(channels []string) func(g *gocui.Gui) error {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()

		chatMaxY := maxY
		initialView := "chat"

		if len(channels) == 1 {
			chatMaxY = maxY - 2
			initialView = "entry"

			if v, err := g.SetView("channel", 0, chatMaxY, len(channels[0])+2, maxY, gocui.TOP); err != nil {
				if !errors.Is(err, gocui.ErrUnknownView) {
					return err
				}

				v.Frame = false
				v.FgColor = gocui.ColorGreen

				fmt.Fprint(v, channels[0]+":")
			}

			if v, err := g.SetView("entry", len(channels[0])+2, chatMaxY, maxX, maxY, gocui.TOP); err != nil {
				if !errors.Is(err, gocui.ErrUnknownView) {
					return err
				}

				v.Frame = false
				v.Editable = true
				v.Wrap = true

				g.Cursor = true
			}
		}

		if v, err := g.SetView("chat", 0, -1, maxX, chatMaxY, gocui.TOP); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}

			v.Autoscroll = true
			v.Wrap = true
			v.Frame = false

			if _, err := g.SetCurrentView(initialView); err != nil {
				return err
			}
		}

		return nil
	}
}

func getTime() string {
	t := time.Now()
	ft := t.Format("15:04")

	return aurora.Bold(ft).String()
}

func chatLogger(s string, g *gocui.Gui, time ...string) {
	var t string

	if len(time) == 0 {
		t = getTime()
	} else {
		t = time[0]
	}

	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("chat")
		if err != nil {
			return err
		}

		fmt.Fprintln(v, t, s)

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
		chatLogger("you: "+v.Buffer(), g)
		v.Clear()

		return nil
	}
}
