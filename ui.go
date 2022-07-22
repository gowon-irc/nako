package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
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

func chatLogger(s string, g *gocui.Gui, tt ...string) {
	var t string

	if len(tt) == 0 {
		now := time.Now()
		t = now.Format("15:04")
	} else {
		t = tt[0]
	}

	ft := aurora.Bold(t).String()

	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("chat")
		if err != nil {
			return err
		}

		fmt.Fprintln(v, ft, s)

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

func genColourAllocator(seed int) func(s string) uint8 {
	m := make(map[string]uint8)

	return func(s string) uint8 {
		v, p := m[s]
		if p {
			return v
		}

		sum := 0
		for _, c := range s {
			sum += int(c)
		}

		rand.Seed(int64(seed + sum))
		id := uint8(rand.Intn(6) + 1)
		m[s] = id
		return id
	}
}

func ircToAnsiColours(s string) string {
	out := s

	m := map[string]string{
		"\u000301": "\x1b[30m", // red
		"\u000302": "\x1b[34m", // navy blue -> blue
		"\u000303": "\x1b[32m", // green
		"\u000304": "\x1b[31m", // red
		"\u000305": "\x1b[33m", // brown -> yellow
		"\u000306": "\x1b[35m", // purple -> magenta
		"\u000307": "\x1b[32m", // olive -> green
		"\u000308": "\x1b[33m", // yellow
		"\u000309": "\x1b[32m", // lime green -> green
		"\u000310": "\x1b[36m", // teal -> cyan
		"\u000311": "\x1b[36m", // aqua blue -> cyan
		"\u000312": "\x1b[34m", // royal blue -> blue
		"\u000313": "\x1b[31m", // hot pink -> red
		"\u000314": "\x1b[30m", // dark grey -> black
		"\u000315": "\x1b[30m", // light grey -> black
		"\u000316": "\x1b[37m", // white
		"\u000399": "\x1b[0m",  // reset
	}

	for k, v := range m {
		out = strings.Replace(out, k, v, -1)
	}

	return out
}

func getCommand(s string) (command string, args []string) {
	if !strings.HasPrefix(s, "/") {
		return "", []string{}
	}

	fields := strings.Fields(s)
	command = strings.TrimPrefix(fields[0], "/")

	return command, fields[1:]
}

func stringIsNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func genSendMessage(c mqtt.Client, module, topicRoot, channel string) func(g *gocui.Gui, v *gocui.View) error {
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

		if strings.HasPrefix(b, "/") {
			if !strings.HasPrefix(b, "//") {
				chatLogger("command not recognised", g)
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
			chatLogger(err.Error(), g)
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
