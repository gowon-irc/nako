package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/awesome-gocui/gocui"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jessevdk/go-flags"
)

const (
	mqttConnectRetryInternal = 5
	mqttDisconnectTimeout    = 1000
)

type Options struct {
	Broker     string   `short:"b" long:"broker" env:"NAKO_BROKER" default:"localhost:1883" description:"mqtt broker"`
	TopicRoot  string   `short:"t" long:"topic-root" env:"NAKO_TOPIC_ROOT" default:"/gowon" description:"mqtt topic root"`
	Module     string   `short:"m" long:"module" env:"NAKO_MODULE" default:"nako" description:"gowon module name"`
	Channels   []string `short:"c" long:"channels" env:"NAKO_CHANNELS" env-delim:"," description:"Channels to watch"`
	ShowJoins  bool     `short:"j" long:"show-joins" env:"NAKO_SHOW_JOINS" description:"Show join and part messages"`
	ColourSeed string   `short:"x" long:"color-seed" env:"NAKO_COLOUR_SEED" default:"0,7" description:"Colour seed,bound"`
}

func main() {
	opts := Options{}
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.SetManagerFunc(genLayout(opts.Channels))

	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(fmt.Sprintf("tcp://%s", opts.Broker))
	mqttOpts.SetClientID(opts.Module)
	mqttOpts.SetConnectRetry(true)
	mqttOpts.SetConnectRetryInterval(mqttConnectRetryInternal * time.Second)
	mqttOpts.SetAutoReconnect(true)

	mqttOpts.DefaultPublishHandler = genDefaultPublishHandler(g)
	mqttOpts.OnConnectionLost = genOnConnectionLostHandler(g)
	mqttOpts.OnReconnecting = genOnRecconnectingHandler(g)
	mqttOpts.OnConnect = createOnConnectHandler(opts.TopicRoot, opts.Channels, g)

	chatLogger("connecting to broker", g)

	c := mqtt.NewClient(mqttOpts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if len(opts.Channels) == 1 {
		if err := g.SetKeybinding("chat", gocui.KeyTab, gocui.ModNone, entrySwitch); err != nil {
			log.Panicln(err)
		}

		if err := g.SetKeybinding("entry", gocui.KeyTab, gocui.ModNone, chatSwitch); err != nil {
			log.Panicln(err)
		}

		if err := g.SetKeybinding("entry", gocui.KeyCtrlU, gocui.ModNone, entryClear); err != nil {
			log.Panicln(err)
		}

		sendMessage := genSendMessage(c, opts.Module, opts.TopicRoot+"/output", opts.Channels[0])
		if err := g.SetKeybinding("entry", gocui.KeyEnter, gocui.ModNone, sendMessage); err != nil {
			log.Panicln(err)
		}
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}
