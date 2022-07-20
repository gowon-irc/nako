package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
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
	ShowJoins   bool     `short:"j" long:"show-joins" env:"NAKO_SHOW_JOINS" description:"Show join and part messages"`
	ColourSeed  int      `short:"s" long:"color-seed" env:"NAKO_COLOUR_SEED" default:"0" description:"Colour seed"`
	ColourBound int      `short:"B" long:"color-bound" env:"NAKO_COLOUR_BOUND" default:"7" description:"Color bound (0-n)"`
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

	rand.Seed(time.Now().UnixNano())
	clientId := "nako_" + fmt.Sprint(rand.Int())

	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(fmt.Sprintf("tcp://%s", opts.Broker))
	mqttOpts.SetClientID(clientId)
	mqttOpts.SetConnectRetry(true)
	mqttOpts.SetConnectRetryInterval(mqttConnectRetryInternal * time.Second)
	mqttOpts.SetAutoReconnect(true)

	mqttOpts.DefaultPublishHandler = genDefaultPublishHandler(g)
	mqttOpts.OnConnectionLost = genOnConnectionLostHandler(g)
	mqttOpts.OnReconnecting = genOnRecconnectingHandler(g)

	privMsgHandler := genPrivMsgHandler(g, opts.Channels, opts.ColourSeed)
	rawMsgHandler := genRawMsgHandler(g)
	mqttOpts.OnConnect = createOnConnectHandler(g, opts.TopicRoot, opts.Channels, privMsgHandler, rawMsgHandler)

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

		sendMessage := genSendMessage(c, clientId, opts.TopicRoot+"/output", opts.Channels[0], opts.ColourSeed)
		if err := g.SetKeybinding("entry", gocui.KeyEnter, gocui.ModNone, sendMessage); err != nil {
			log.Panicln(err)
		}
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}
