package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"ransmart_notify/app/helper/response"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Notify struct {
	Id      int    `json:"id"`
	OrderId int    `json:"order_id"`
	Request string `json:"request"`
	IsSent  bool   `json:"is_sent"`
	gorm.Model
}

type ReqNotify struct {
	OrderId int         `json:"order_id"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
// 	fmt.Printf("%s", msg.Payload())
// }

// var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
// 	fmt.Println("Connected")
// }

// var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
// 	fmt.Printf("Connect lost: %v", err)
// }

func main() {

	// Load env
	err := godotenv.Load("params/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to database
	dsn := os.Getenv("DB_USERNAME") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME") + "?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Migrate the schema
	db.AutoMigrate(&Notify{})

	// Create a new MQTT Client
	// var broker = os.Getenv("BROKER")
	// var port = cast.ToInt(os.Getenv("MQTT_PORT"))
	// opts := mqtt.NewClientOptions()
	// opts.AddBroker(fmt.Sprintf("%s:%d", broker, port))
	// opts.SetClientID("go_mqtt_client")
	// opts.SetUsername("emqx")
	// opts.SetPassword("public")
	// opts.SetDefaultPublishHandler(messagePubHandler)
	// opts.OnConnect = connectHandler
	// opts.OnConnectionLost = connectLostHandler
	// client := mqtt.NewClient(opts)
	// if token := client.Connect(); token.Wait() && token.Error() != nil {
	// 	panic(token.Error())
	// }

	// // subcribe to topic
	// sub(client)

	// router
	r := chi.NewRouter()

	// check service
	r.Group(func(g chi.Router) {
		g.Get("/check", func(w http.ResponseWriter, r *http.Request) {
			response.Response(w, http.StatusOK, "Service is running", nil)
		})

		g.Post("/publis", func(w http.ResponseWriter, r *http.Request) {
			decoder := json.NewDecoder(r.Body)
			var datarequest ReqNotify
			if err := decoder.Decode(&datarequest); err != nil {
				response.Response(w, http.StatusBadRequest, "Invalid request payload", nil)
				return
			}

			// to json string
			newRequst, err := json.Marshal(datarequest)
			if err != nil {
				response.Response(w, http.StatusBadRequest, "Invalid request payload", nil)
				return
			}

			// Create
			db.Create(&Notify{OrderId: datarequest.OrderId, Request: string(newRequst)})

			// Publish
			// publish(client, string(newRequst))

			// update is_sent
			db.Model(&Notify{}).Where("order_id = ?", datarequest.OrderId).Update("is_sent", true)

			// Response
			response.Response(w, http.StatusOK, "Publish success", datarequest)
		})

		g.Get("/notifies", func(w http.ResponseWriter, r *http.Request) {
			var data []Notify
			db.Find(&data)
			response.Response(w, http.StatusOK, "Get notifies success", data)
		})
	})

	log.Println("Service running on " + os.Getenv("HOST") + ":" + os.Getenv("PORT"))

	portServer := os.Getenv("PORT")

	if portServer == "" {
		portServer = "8080"
	}

	if err := http.ListenAndServe(":"+portServer, r); err != nil {
		log.Println("Error Starting Service")
	}
}

// func publish(client mqtt.Client, msg string) {
// 	token := client.Publish(os.Getenv("TOPIC"), 0, false, msg)
// 	token.Wait()
// }

// func sub(client mqtt.Client) {
// 	topic := os.Getenv("TOPIC")
// 	token := client.Subscribe(topic, 1, nil)
// 	token.Wait()
// }
