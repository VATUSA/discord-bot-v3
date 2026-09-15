package bot

import (
	"context"
	"github.com/VATUSA/discord-bot-v3/internal/commands"
	"github.com/bwmarrin/discordgo"
	"log"
)

func QueueListen(ctx context.Context, s *discordgo.Session, cmds <-chan commands.Command) {
	log.Print("Waiting for commands...")

	for {
		select {
		case <-ctx.Done():
			log.Print("Stopped processing commands.")
			return
		case cmd := <-cmds:
			switch c := cmd.(type) {
			case commands.SyncMember:
				log.Printf("Received sync for member: %s", c.UserID)
				ProcessMemberInGuilds(s, c.UserID)
			}
		}
	}
}
