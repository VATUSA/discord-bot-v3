package bot

import (
	"github.com/VATUSA/discord-bot-v3/internal/queue"
	"github.com/bwmarrin/discordgo"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func QueueListen(s *discordgo.Session) {
	conn, err := amqp.Dial(queue.ConnectionString())
	if err != nil {
		log.Println("Failed to connect to RabbitMQ")
		return
	}
	notify := make(chan *amqp.Error)
	conn.NotifyClose(notify)
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Println("Failed to open a channel")
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"discord_sync", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Println("Failed to declare a queue")
		return
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Println("Failed to register a consumer")
		return
	}

	log.Print("Connected to RabbitMQ. Waiting for messages...")

loop:
	for {
		select {
		case err = <-notify:
			break loop
		case d := <-msgs:
			log.Printf("Received a message: %s", d.Body)
			ProcessMemberInGuilds(s, string(d.Body))
			d.Ack(false)
		}
	}
	log.Print("Stopped processing RabbitMQ.")
}
