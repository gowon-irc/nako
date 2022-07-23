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
	Broker      string   `short:"b" long:"broker" env:"NAKO_BROKER" default:"localhost:1883" description:"mqtt broker"`
	TopicRoot   string   `short:"t" long:"topic-root" env:"NAKO_TOPIC_ROOT" default:"/gowon" description:"mqtt topic root"`
	Channels    []string `short:"c" long:"channels" env:"NAKO_CHANNELS" env-delim:"," description:"Channels to watch"`
	Highlights  []string `short:"H" long:"highlights" env:"NAKO_HIGHLIGHTS" env-delim:"," description:"Words to highlight"`
	ShowJoins   bool     `short:"j" long:"show-joins" env:"NAKO_SHOW_JOINS" description:"Show join and part messages"`
	ColourSeed  int      `short:"s" long:"color-seed" env:"NAKO_COLOUR_SEED" default:"0" description:"Colour seed"`
	ColourBound int      `short:"B" long:"color-bound" env:"NAKO_COLOUR_BOUND" default:"7" description:"Color bound (0-n)"`
}

func main() {
	// Parse options

	opts := Options{}
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	// Create gui

	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.SetManagerFunc(genLayout(opts.Channels))

	// Setup application logger

	loggerFunc := genChatViewLoggerFunc(g)
	appLogger := createLogger(loggerFunc)

	// Setup mqtt client

	clientId := "nako_" + fmt.Sprint(os.Getpid())
	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(fmt.Sprintf("tcp://%s", opts.Broker))
	mqttOpts.SetClientID(clientId)
	mqttOpts.SetConnectRetry(true)
	mqttOpts.SetConnectRetryInterval(mqttConnectRetryInternal * time.Second)
	mqttOpts.SetAutoReconnect(true)

	// Setup mqtt handlers

	mqttOpts.DefaultPublishHandler = genDefaultPublishHandler(appLogger)
	mqttOpts.OnConnectionLost = genOnConnectionLostHandler(appLogger)
	mqttOpts.OnReconnecting = genOnRecconnectingHandler(appLogger)

	colourAllocator := createColourAllocator(opts.ColourSeed)
	privMsgHandler := genPrivMsgHandler(opts.Channels, opts.Highlights, colourAllocator, appLogger)
	rawMsgHandler := genRawMsgHandler(opts.Channels, colourAllocator, appLogger)
	mqttOpts.OnConnect = createOnConnectHandler(opts.TopicRoot, opts.Channels, privMsgHandler, rawMsgHandler, appLogger)

	// Connect to mqtt broker

	appLogger.Log("connecting to broker")

	c := mqtt.NewClient(mqttOpts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Setup gui keybindings

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

		sendMessage := genSendMessage(c, clientId, opts.TopicRoot, opts.Channels[0], appLogger)
		if err := g.SetKeybinding("entry", gocui.KeyEnter, gocui.ModNone, sendMessage); err != nil {
			log.Panicln(err)
		}
	}

	if err := g.SetKeybinding("chat", 'j', gocui.ModNone, genScrollX(1)); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("chat", 'k', gocui.ModNone, genScrollX(-1)); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("chat", 'J', gocui.ModNone, genScrollX(10)); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("chat", 'K', gocui.ModNone, genScrollX(-10)); err != nil {
		log.Panicln(err)
	}

	// Start gui

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}
