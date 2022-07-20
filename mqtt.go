package main

import (
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gowon-irc/go-gowon"
	"github.com/logrusorgru/aurora"
)

func genDefaultPublishHandler(g *gocui.Gui) func(c mqtt.Client, msg mqtt.Message) {
	return func(c mqtt.Client, msg mqtt.Message) {
		chatLogger(fmt.Sprintf("unexpected message:  %s\n", msg), g)
	}
}

func genOnConnectionLostHandler(g *gocui.Gui) func(c mqtt.Client, err error) {
	return func(c mqtt.Client, err error) {
		chatLogger("connection to broker lost", g)
	}
}

func genOnRecconnectingHandler(g *gocui.Gui) func(c mqtt.Client, opts *mqtt.ClientOptions) {
	return func(c mqtt.Client, opts *mqtt.ClientOptions) {
		chatLogger("attempting to reconnect to broker", g)
	}
}

func containsString(ss []string, s string) bool {
	for _, i := range ss {
		if i == s {
			return true
		}
	}
	return false
}

func genPrivMsgHandler(g *gocui.Gui, channels, highlights []string, seed int) func(client mqtt.Client, msg mqtt.Message) {
	colorAllocator := genColorAllocator(seed)

	return func(client mqtt.Client, msg mqtt.Message) {
		m, err := gowon.CreateMessageStruct(msg.Payload())

		if err != nil {
			chatLogger(err.Error(), g)
			return
		}

		if len(channels) > 0 && !containsString(channels, m.Dest) {
			return
		}

		id := colorAllocator(m.Nick)
		out := aurora.Index(id, fmt.Sprintf("%s: %s", m.Nick, m.Msg))

		for _, h := range highlights {
			if strings.Contains(m.Msg, h) {
				out = out.Black().BgIndex(id)
			}
		}

		output := ircToAnsiColours(out.String())

		chatLogger(output, g)
	}
}

func genRawMsgHandler(g *gocui.Gui, channels []string, seed int) func(client mqtt.Client, msg mqtt.Message) {
	colorAllocator := genColorAllocator(seed)

	return func(client mqtt.Client, msg mqtt.Message) {
		m, err := gowon.CreateMessageStruct(msg.Payload())

		if err != nil && err.Error() != gowon.ErrorMessageNoBodyMsg {
			chatLogger(err.Error(), g)
			return
		}

		id := colorAllocator(m.Nick)

		if m.Code == "JOIN" {
			if len(channels) > 0 && !containsString(channels, m.Arguments[0]) {
				return
			}

			out := aurora.Index(id, fmt.Sprintf("-> %s joined %s", m.Nick, m.Arguments[0])).String()
			chatLogger(out, g)
			return
		}

		if m.Code == "332" {
			if len(channels) > 0 && !containsString(channels, m.Arguments[1]) {
				return
			}

			out := fmt.Sprintf("topic for %s is: \"%s\"", m.Arguments[1], m.Arguments[2])
			chatLogger(out, g)
		}

		if m.Code == "353" {
			if len(channels) > 0 && !containsString(channels, m.Arguments[2]) {
				return
			}

			out := fmt.Sprintf("In %s are: %s", m.Arguments[2], strings.Join(m.Arguments[3:], " "))
			chatLogger(out, g)
		}
	}
}

func createOnConnectHandler(g *gocui.Gui, topicRoot string, channels []string, pmh, rmh mqtt.MessageHandler) func(mqtt.Client) {
	inputTopic := topicRoot + "/input"
	rawInputTopic := topicRoot + "/raw/input"
	rawOutputTopic := topicRoot + "/raw/output"

	return func(client mqtt.Client) {
		chatLogger("connected to broker", g)

		client.Subscribe(inputTopic, 0, pmh)
		chatLogger(fmt.Sprintf(fmt.Sprintf("Subscription to %s complete", inputTopic)), g)

		client.Subscribe(rawInputTopic, 0, rmh)
		chatLogger(fmt.Sprintf(fmt.Sprintf("Subscription to %s complete", rawInputTopic)), g)

		client.Publish(rawOutputTopic, 0, false, fmt.Sprintf("JOIN %s", strings.Join(channels, ",")))

		for _, c := range channels {
			client.Publish(rawOutputTopic, 0, false, fmt.Sprintf("TOPIC %s", c))
			client.Publish(rawOutputTopic, 0, false, fmt.Sprintf("NAMES %s", c))
		}
	}
}
