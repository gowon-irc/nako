package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gowon-irc/go-gowon"
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

func defaultPublishHandler(c mqtt.Client, msg mqtt.Message) {
	log.Printf("unexpected message:  %s\n", msg)
}

func onConnectionLostHandler(c mqtt.Client, err error) {
	log.Println("connection to broker lost")
}

func onRecconnectingHandler(c mqtt.Client, opts *mqtt.ClientOptions) {
	log.Println("attempting to reconnect to broker")
}

func privMsgHandler(client mqtt.Client, msg mqtt.Message) {
	m, err := gowon.CreateMessageStruct(msg.Payload())
	if err != nil {
		log.Print(err)

		return
	}

	fmt.Println(m.Msg)
}

func rawMsgHandler(client mqtt.Client, msg mqtt.Message) {
	fmt.Println(string(msg.Payload()))
}

func createOnConnectHandler(topicRoot string) func(mqtt.Client) {
	topic := topicRoot + "/input"
	rawTopic := topicRoot + "/raw/input"

	return func(client mqtt.Client) {
		log.Println("connected to broker")

		client.Subscribe(topic, 0, privMsgHandler)
		log.Printf(fmt.Sprintf("Subscription to %s complete", topic))

		client.Subscribe(rawTopic, 0, rawMsgHandler)
		log.Printf(fmt.Sprintf("Subscription to %s complete", rawTopic))
	}
}

func main() {
	opts := Options{}
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(fmt.Sprintf("tcp://%s", opts.Broker))
	mqttOpts.SetClientID(opts.Module)
	mqttOpts.SetConnectRetry(true)
	mqttOpts.SetConnectRetryInterval(mqttConnectRetryInternal * time.Second)
	mqttOpts.SetAutoReconnect(true)

	mqttOpts.DefaultPublishHandler = defaultPublishHandler
	mqttOpts.OnConnectionLost = onConnectionLostHandler
	mqttOpts.OnReconnecting = onRecconnectingHandler
	mqttOpts.OnConnect = createOnConnectHandler(opts.TopicRoot)

	log.Print("connecting to broker")

	c := mqtt.NewClient(mqttOpts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	fmt.Println()
	log.Println("signal caught, exiting")
	c.Disconnect(mqttDisconnectTimeout)
	log.Println("shutdown complete")
}
