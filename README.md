# gomeqtt
Simple to use MQTT abstraction lib

Based on the paho.mqtt.golang I build this library to quickly include MQTT messages and commands in the house automation code.
I didn't always wanted to setup everything from scratch.

# Todos
- qos settings
- consume messages


# Examples

## Publish example
```
package main

import (
	"fmt"
	"log"

	mqtt "github.com/kuli21/gomeqtt/eventbus"
)

func main() {
	fmt.Println("test mqtt lib")
	c := mqtt.MqttConfig{
		Host:     "blubb.de",
		Port:     8883,
		ClientId: "test-go-12345",
		Username: "user1",
		Password: "pass1",
		UseTls:   true,
		CaFile:   "./config/certs/blubb.de.ca.pem",
		CrtFile:  "./config/certs/blubb.de.crt.pem",
		KeyFile:  "./config/certs/blubb.de.key.pem",
	}
	mc, err := c.NewMqttConnection()
	if err != nil {
		log.Panic(err)
	}
	defer mc.Close()

	mc.Publish("test/bla", "moinmoin")

    var f mqtt.SubscribeCallback = func(topic string, payload string) {
		fmt.Println("callback: recieved message from:", topic, ":", payload)
	}
	mc.Subscribe("test/bla", f)

	for i := 1; i < 20; i++ {
		time.Sleep(time.Second * 1)
		o := fmt.Sprintf("%d", i)
		mc.Publish("test/bla", "loop-no: " + o)
	}
}

```