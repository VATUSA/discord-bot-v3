package bot

import (
	"github.com/VATUSA/discord-bot-v3/internal/config"
	"github.com/VATUSA/discord-bot-v3/internal/integration/api2"
	"github.com/bwmarrin/discordgo"
	"log"
)

func AddMemberHandlers(s *discordgo.Session) {
	s.AddHandler(ProcessGuildMembersChunk)
}

func ProcessMember(s *discordgo.Session, m *discordgo.Member, cfg *config.ServerConfig) error {
	controller, err := api2.GetControllerData(m.User.ID)
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
	cfg := config.GetServerConfig(mc.GuildID)
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

func RequestGuildMembers(s *discordgo.Session, g *discordgo.Guild, cfg *config.ServerConfig) error {
	log.Printf("Fetching members for guild %s (%s)", g.ID, cfg.Name)
	err := s.RequestGuildMembers(g.ID, "", 0, "1", true)
	if err != nil {
		return err
	}
	return nil
}
