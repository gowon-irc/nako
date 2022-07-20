package main

import (
	"fmt"

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

func genPrivMsgHandler(g *gocui.Gui, channels []string, seed int) func(client mqtt.Client, msg mqtt.Message) {
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
		out := aurora.Index(id, fmt.Sprintf("%s: %s", m.Nick, m.Msg)).String()

		chatLogger(out, g)
	}
}

func genRawMsgHandler(g *gocui.Gui) func(client mqtt.Client, msg mqtt.Message) {
	return func(client mqtt.Client, msg mqtt.Message) {
		return
		// chatLogger(string(msg.Payload()), g)
	}
}

func createOnConnectHandler(g *gocui.Gui, topicRoot string, pmh, rmh mqtt.MessageHandler) func(mqtt.Client) {
	topic := topicRoot + "/input"
	rawTopic := topicRoot + "/raw/input"

	return func(client mqtt.Client) {
		chatLogger("connected to broker", g)

		client.Subscribe(topic, 0, pmh)
		chatLogger(fmt.Sprintf(fmt.Sprintf("Subscription to %s complete", topic)), g)

		client.Subscribe(rawTopic, 0, rmh)
		chatLogger(fmt.Sprintf(fmt.Sprintf("Subscription to %s complete", rawTopic)), g)
	}
}
