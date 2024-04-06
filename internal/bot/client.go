package bot

import (
	"github.com/VATUSA/discord-bot-v3/internal/config"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"time"
)

func Session() (*discordgo.Session, error) {
	discord, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		return nil, err
	}
	return discord, nil
}

func Run() {
	log.Print("Starting discord-bot-v3")
	session, err := Session()
	if err != nil {
		println(err.Error())
	}
	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	AddMemberHandlers(session)

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		go IntervalRefreshAll(s)
		go IntervalReloadConfigs()
		go func() {
			for {
				QueueListen(s)
				log.Println("MQ Connection lost. Attempting to reconnect.")
				time.Sleep(5 * time.Second)
			}
		}()
	})

	// TODO: Add hook for GuildMemberAdd to automatically trigger roles for that member.

	err = session.Open()
	if err != nil {
		println(err.Error())
	}
	defer session.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}
