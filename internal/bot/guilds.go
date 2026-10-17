package bot

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

func ProcessMemberInGuilds(s *discordgo.Session, id string) {
	for _, guild := range s.State.Guilds {
		cfg := GetServerConfig(guild.ID)
		if cfg != nil && cfg.Active {
			member, err := s.GuildMember(guild.ID, id)
			if err != nil {
				continue
			}
			err = ProcessMember(s, member, cfg)
			if err != nil {
				log.Printf("[%s] Error in ProcessMember %s: %s", cfg.Name, member.User.ID, err.Error())
			}
		}
	}
}
