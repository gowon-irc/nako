package main

import (
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gowon-irc/go-gowon"
	"github.com/logrusorgru/aurora"
)

func genDefaultPublishHandler(l *logger) func(c mqtt.Client, msg mqtt.Message) {
	return func(c mqtt.Client, msg mqtt.Message) {
		l.Log(fmt.Sprintf("unexpected message:  %s\n", msg))
	}
}

func genOnConnectionLostHandler(l *logger) func(c mqtt.Client, err error) {
	return func(c mqtt.Client, err error) {
		l.Log("connection to broker lost")
	}
}

func genOnRecconnectingHandler(l *logger) func(c mqtt.Client, opts *mqtt.ClientOptions) {
	return func(c mqtt.Client, opts *mqtt.ClientOptions) {
		l.Log("attempting to reconnect to broker")
	}
}

func genPrivMsgHandler(channels, highlights []string, ca *colourAllocator, l *logger) func(client mqtt.Client, msg mqtt.Message) {
	return func(client mqtt.Client, msg mqtt.Message) {
		m, err := gowon.CreateMessageStruct(msg.Payload())

		if err != nil {
			l.Log(err.Error())
			return
		}

		if len(channels) > 0 && !containsString(channels, m.Dest) {
			return
		}

		id := ca.Allocate(m.Nick)
		out := aurora.Index(id, fmt.Sprintf("%s: %s", m.Nick, m.Msg))

		for _, h := range highlights {
			if strings.Contains(m.Msg, h) {
				out = out.Black().BgIndex(id)
			}
		}

		output := ircToAnsiColours(out.String())

		serverTime := m.Tags["time"]

		if serverTime == "" {
			l.Log(output)
			return
		}

		t, err := time.Parse("2006-01-02T15:04:05.000Z", serverTime)
		if err != nil {
			l.Log(output)
			return
		}

		l.Log(output, t.Format("15:04"))
	}
}

func genRawMsgHandler(channels []string, ca *colourAllocator, l *logger) func(client mqtt.Client, msg mqtt.Message) {
	return func(client mqtt.Client, msg mqtt.Message) {
		m, err := gowon.CreateMessageStruct(msg.Payload())

		if err != nil && err.Error() != gowon.ErrorMessageNoBodyMsg {
			l.Log(err.Error())
			return
		}

		id := ca.Allocate(m.Nick)

		if m.Code == "JOIN" {
			if len(channels) > 0 && !containsString(channels, m.Arguments[0]) {
				return
			}

			out := aurora.Index(id, fmt.Sprintf("-> %s joined %s", m.Nick, m.Arguments[0])).String()
			l.Log(out)
			return
		}

		if m.Code == "332" {
			if len(channels) > 0 && !containsString(channels, m.Arguments[1]) {
				return
			}

			out := fmt.Sprintf("topic for %s is: \"%s\"", m.Arguments[1], m.Arguments[2])
			l.Log(out)
		}

		if m.Code == "353" {
			if len(channels) > 0 && !containsString(channels, m.Arguments[2]) {
				return
			}

			out := fmt.Sprintf("In %s are: %s", m.Arguments[2], colourNamesList(m.Arguments[3], ca))
			l.Log(out)
		}
	}
}

func createOnConnectHandler(topicRoot string, channels []string, pmh, rmh mqtt.MessageHandler, l *logger) func(mqtt.Client) {
	inputTopic := topicRoot + "/input"
	rawInputTopic := topicRoot + "/raw/input"
	rawOutputTopic := topicRoot + "/raw/output"

	return func(client mqtt.Client) {
		l.Log("connected to broker")

		client.Subscribe(inputTopic, 0, pmh)
		l.Log(fmt.Sprintf(fmt.Sprintf("Subscription to %s complete", inputTopic)))

		client.Subscribe(rawInputTopic, 0, rmh)
		l.Log(fmt.Sprintf(fmt.Sprintf("Subscription to %s complete", rawInputTopic)))

		client.Publish(rawOutputTopic, 0, false, fmt.Sprintf("JOIN %s", strings.Join(channels, ",")))

		for _, c := range channels {
			client.Publish(rawOutputTopic, 0, false, fmt.Sprintf("TOPIC %s", c))
			client.Publish(rawOutputTopic, 0, false, fmt.Sprintf("NAMES %s", c))
		}
	}
}
