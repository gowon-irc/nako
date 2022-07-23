package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gowon-irc/go-gowon"
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

func genSendMessage(c mqtt.Client, module, topicRoot, channel string, l *logger) func(g *gocui.Gui, v *gocui.View) error {
	inputTopic := topicRoot + "/input"
	outputTopic := topicRoot + "/output"
	rawOutputTopic := topicRoot + "/raw/output"

	return func(g *gocui.Gui, v *gocui.View) error {
		b := v.Buffer()

		if b == "" {
			return nil
		}

		v.Clear()

		command, args := getCommand(b)

		if command == "ch" || command == "chatlog" {
			hl := "10"

			if len(args) > 0 && stringIsNumber(args[0]) {
				hl = args[0]
			}

			c.Publish(rawOutputTopic, 0, false, fmt.Sprintf("CHATHISTORY LATEST %s * %s", channel, hl))
			return nil
		}

		if command == "t" || command == "topic" {
			c.Publish(rawOutputTopic, 0, false, fmt.Sprintf("TOPIC %s", channel))
			return nil
		}

		if command == "n" || command == "names" {
			c.Publish(rawOutputTopic, 0, false, fmt.Sprintf("NAMES %s", channel))
			return nil
		}

		if command == "c" || command == "clear" {
			g.Update(func(g *gocui.Gui) error {
				vc, err := g.View("chat")
				if err != nil {
					return err
				}

				vc.Clear()

				return nil
			})

			return nil
		}

		if strings.HasPrefix(b, "/") {
			if !strings.HasPrefix(b, "//") {
				l.Log("command not recognised")
				return nil
			}
			b = strings.TrimPrefix(b, "/")
		}

		m := &gowon.Message{
			Module: module,
			Nick:   "you",
			Dest:   channel,
			Msg:    b,
		}

		mj, err := json.Marshal(m)
		if err != nil {
			l.Log(err.Error())
			return err
		}

		c.Publish(inputTopic, 0, false, mj)
		c.Publish(outputTopic, 0, false, mj)
		v.Clear()

		return nil
	}
}

func genScrollX(y int) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		ox, oy := v.Origin()
		ny := oy + y

		// if we have reached the bottom, don't scroll down
		if ny < 0 {
			return nil
		}

		_, h := v.Size()
		lh := v.LinesHeight()

		// if we have less lines than the view holds, don't scroll
		if lh < h {
			return nil
		}

		// if the last line is visible, don't scroll up
		if (h + ny) >= lh {
			v.Autoscroll = true
			return nil
		}

		v.Autoscroll = false
		if err := v.SetOrigin(ox, oy+y); err != nil {
			return err
		}

		return nil
	}
}
