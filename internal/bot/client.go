package bot

import (
	"context"
	"github.com/VATUSA/discord-bot-v3/internal/commands"
	"github.com/VATUSA/discord-bot-v3/internal/config"
	"github.com/bwmarrin/discordgo"
	"log"
	"sync"
	"sync/atomic"
)

var ready int32

// IsReady reports whether the bot has logged in to Discord.
func IsReady() bool {
	return atomic.LoadInt32(&ready) == 1
}

func Session() (*discordgo.Session, error) {
	discord, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		return nil, err
	}
	return discord, nil
}

// Run connects to Discord and processes commands until ctx is cancelled.
func Run(ctx context.Context, cmds <-chan commands.Command) error {
	log.Print("Starting discord-bot-v3")
	session, err := Session()
	if err != nil {
		return err
	}
	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	AddMemberHandlers(session)

	// Ready fires again on every full reconnect, so only start the background workers once.
	var startWorkers sync.Once
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		startWorkers.Do(func() {
			atomic.StoreInt32(&ready, 1)
			go IntervalRefreshAll(s)
			go IntervalReloadConfigs()
			go QueueListen(ctx, s, cmds)
		})
	})

	// TODO: Add hook for GuildMemberAdd to automatically trigger roles for that member.

	err = session.Open()
	if err != nil {
		return err
	}
	defer session.Close()

	<-ctx.Done()
	log.Print("Shutting down discord-bot-v3")
	return nil
}
