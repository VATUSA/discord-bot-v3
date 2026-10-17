package bot

import (
	"github.com/VATUSA/discord-bot-v3/internal/api"
	"github.com/bwmarrin/discordgo"
	"log"
)

func AddMemberHandlers(s *discordgo.Session) {
	s.AddHandler(ProcessGuildMembersChunk)
	s.AddHandler(ProcessGuildMemberAdd)
}

// ProcessGuildMemberAdd syncs a member as soon as they join, rather than
// waiting for the next hourly full pass.
func ProcessGuildMemberAdd(s *discordgo.Session, e *discordgo.GuildMemberAdd) {
	cfg := GetServerConfig(e.GuildID)
	if cfg == nil || !cfg.Active {
		return
	}
	err := ProcessMember(s, e.Member, cfg)
	if err != nil {
		log.Printf("[%s] Error in ProcessMember %s on join: %s", cfg.Name, e.User.ID, err.Error())
	}
}

func ProcessMember(s *discordgo.Session, m *discordgo.Member, cfg *ServerConfig) error {
	if m.User.Bot {
		return nil // Don't try to process bots
	}
	controller, err := api.GetControllerData(m.User.ID)
	if err != nil {
		return err
	}
	return syncMember(s, m, controller, cfg)
}

// syncMember applies already-fetched controller data to a member's nickname and roles.
func syncMember(s *discordgo.Session, m *discordgo.Member, controller *api.ControllerData, cfg *ServerConfig) error {
	err := SyncName(s, m, controller, cfg)
	if err != nil {
		return err
	}
	return SyncRoles(s, m, controller, cfg)
}

// ProcessGuildMembersChunk hands chunks to the running sweep, which dedupes
// users across guilds and paces the API lookups.
func ProcessGuildMembersChunk(s *discordgo.Session, mc *discordgo.GuildMembersChunk) {
	cfg := GetServerConfig(mc.GuildID)
	if cfg == nil {
		return
	}
	c := currentCollector()
	if c == nil || !c.add(mc, *cfg) {
		log.Printf("[%s] Ignoring member chunk %d/%d outside of a sweep", cfg.Name, mc.ChunkIndex+1, mc.ChunkCount)
		return
	}
	log.Printf("[%s] Collected member chunk %d/%d (%d members)", cfg.Name, mc.ChunkIndex+1, mc.ChunkCount, len(mc.Members))
}

func RequestGuildMembers(s *discordgo.Session, g *discordgo.Guild, cfg *ServerConfig) error {
	log.Printf("Fetching members for guild %s (%s)", g.ID, cfg.Name)
	err := s.RequestGuildMembers(g.ID, "", 0, "1", false)
	if err != nil {
		return err
	}
	return nil
}
