package eventbus

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttConfig struct {
	Host     string
	Port     int
	ClientId string
	Username string
	Password string
	UseTls   bool
	CaFile   string
	CrtFile  string
	KeyFile  string
}

type MqttClient struct {
	client     mqtt.Client
	recChannel chan string
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v", err)
}

type SubscribeCallback func(t string, p string)

func (m *MqttConfig) NewMqttConnection() (*MqttClient, error) {
	c, err := initMqttClient(m)
	mc := MqttClient{client: c}
	if err != nil {
		return nil, err
	} else {
		return &mc, nil
	}
}

func initMqttClient(m *MqttConfig) (mqtt.Client, error) {
	var broker = m.Host
	var port = m.Port
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("ssl://%s:%d", broker, port))
	opts.SetClientID(m.ClientId)
	opts.SetUsername(m.Username)
	opts.SetPassword(m.Password)
	if m.UseTls {
		tlsConfig := newTlsConfig(m)
		opts.SetTLSConfig(tlsConfig)
	}
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}

func (c MqttClient) Publish(topic string, msg string) {
	token := c.client.Publish(topic, 0, false, msg)
	token.Wait()
	time.Sleep(time.Millisecond * 200)
}

func (c MqttClient) Subscribe(topic string, extCallback SubscribeCallback) {
	var callback mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		v := msg.Payload()
		p := string(v)
		t := msg.Topic()

		extCallback(t, p)
	}

	go subscribeAndListen(c, topic, callback)
}

func subscribeAndListen(c MqttClient, topic string, callback mqtt.MessageHandler) {
	token := c.client.Subscribe(topic, 1, callback)
	for {
		token.Wait()
		time.Sleep(time.Second * 1)
	}
}

func newTlsConfig(m *MqttConfig) *tls.Config {
	certpool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(m.CaFile)
	if err != nil {
		log.Fatalln(err.Error())
	}
	certpool.AppendCertsFromPEM(ca)
	clientKeyPair, err := tls.LoadX509KeyPair(m.CrtFile, m.KeyFile)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{clientKeyPair},
	}
}
