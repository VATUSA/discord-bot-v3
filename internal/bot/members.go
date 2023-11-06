package bot

import (
	"github.com/VATUSA/discord-bot-v3/internal/api"
	"github.com/bwmarrin/discordgo"
	"log"
)

func AddMemberHandlers(s *discordgo.Session) {
	s.AddHandler(ProcessGuildMembersChunk)
}

func ProcessMember(s *discordgo.Session, m *discordgo.Member, cfg *ServerConfig) error {
	if m.User.Bot {
		return nil // Don't try to process bots
	}
	controller, err := api.GetControllerData(m.User.ID)
	if err != nil {
		return err
	}
	err = SyncName(s, m, controller, cfg)
	if err != nil {
		return err
	}
	err = SyncRoles(s, m, controller, cfg)
	if err != nil {
		return err
	}
	return nil
}

func ProcessGuildMembersChunk(s *discordgo.Session, mc *discordgo.GuildMembersChunk) {
	cfg := GetServerConfig(mc.GuildID)
	if cfg == nil {
		return
	}
	for _, member := range mc.Members {
		err := ProcessMember(s, member, cfg)
		if err != nil {
			log.Printf("Error in ProcessMember %s: %s", member.User.ID, err.Error())
		}
	}
}

func RequestGuildMembers(s *discordgo.Session, g *discordgo.Guild, cfg *ServerConfig) error {
	log.Printf("Fetching members for guild %s (%s)", g.ID, cfg.Name)
	err := s.RequestGuildMembers(g.ID, "", 0, "1", true)
	if err != nil {
		return err
	}
	return nil
}
