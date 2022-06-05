package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"ransmart_notify/app/helper/response"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("%s", msg.Payload())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {

	err := godotenv.Load("params/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var broker = os.Getenv("BROKER")
	var port = cast.ToInt(os.Getenv("PORT"))
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername("emqx")
	opts.SetPassword("public")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	sub(client)

	// router
	r := chi.NewRouter()

	// check service
	r.Group(func(g chi.Router) {
		g.Post("/publis", func(w http.ResponseWriter, r *http.Request) {
			decoder := json.NewDecoder(r.Body)
			var datarequest map[string]interface{}
			if err := decoder.Decode(&datarequest); err != nil {
				response.Response(w, http.StatusBadRequest, "Invalid request payload", nil)
				return
			}

			newRequst, err := json.Marshal(datarequest)
			if err != nil {
				response.Response(w, http.StatusBadRequest, "Invalid request payload", nil)
				return
			}

			publish(client, string(newRequst))
			response.Response(w, http.StatusOK, "Publish success", datarequest)
		})
	})

	log.Println("Service running on " + os.Getenv("APP_LOCAL_HOST") + ":" + os.Getenv("APP_LOCAL_PORT"))
	if err := http.ListenAndServe(":"+os.Getenv("APP_LOCAL_PORT"), r); err != nil {
		log.Println("Error Starting Service")
	}
}

func publish(client mqtt.Client, msg string) {
	token := client.Publish(os.Getenv("TOPIC"), 0, false, msg)
	token.Wait()
}

func sub(client mqtt.Client) {
	topic := os.Getenv("TOPIC")
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
}
