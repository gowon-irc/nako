package main

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gowon-irc/go-gowon"
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

func genPrivMsgHandler(g *gocui.Gui) func(client mqtt.Client, msg mqtt.Message) {
	return func(client mqtt.Client, msg mqtt.Message) {
		m, err := gowon.CreateMessageStruct(msg.Payload())
		if err != nil {
			chatLogger(err.Error(), g)

			return
		}

		chatLogger(m.Msg, g)
	}
}

func genRawMsgHandler(g *gocui.Gui) func(client mqtt.Client, msg mqtt.Message) {
	return func(client mqtt.Client, msg mqtt.Message) {
		m, _ := gowon.CreateMessageStruct(msg.Payload())

		chatLogger(m.Raw, g)
	}
}

func createOnConnectHandler(topicRoot string, g *gocui.Gui) func(mqtt.Client) {
	topic := topicRoot + "/input"
	rawTopic := topicRoot + "/raw/input"

	return func(client mqtt.Client) {
		chatLogger("connected to broker", g)

		client.Subscribe(topic, 0, genPrivMsgHandler(g))
		chatLogger(fmt.Sprintf(fmt.Sprintf("Subscription to %s complete", topic)), g)

		client.Subscribe(rawTopic, 0, genRawMsgHandler(g))
		chatLogger(fmt.Sprintf(fmt.Sprintf("Subscription to %s complete", rawTopic)), g)
	}
}
