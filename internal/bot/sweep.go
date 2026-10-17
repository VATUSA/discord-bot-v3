package bot

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/VATUSA/discord-bot-v3/internal/api"
	"github.com/bwmarrin/discordgo"
)

const (
	// refreshInterval is how often a full sweep of every guild starts.
	refreshInterval = 1 * time.Hour
	// sweepBudget is how long a sweep's API lookups are spread over. It is kept
	// below refreshInterval so a sweep finishes before the next one is due.
	sweepBudget = 45 * time.Minute
	// minLookupInterval caps the lookup rate (20/s) however many users there are.
	minLookupInterval = 50 * time.Millisecond
	// maxLookupInterval keeps small sweeps from crawling.
	maxLookupInterval = 1 * time.Second
	// chunkWaitTimeout bounds how long to wait for Discord to deliver member chunks.
	chunkWaitTimeout = 2 * time.Minute
	// progressEvery is how many users are processed between progress log lines.
	progressEvery = 500
)

type guildMember struct {
	member *discordgo.Member
	cfg    ServerConfig
}

// sweepCollector gathers member chunks from every guild requested in a sweep.
type sweepCollector struct {
	mu       sync.Mutex
	members  map[string][]guildMember // user ID -> memberships in active guilds
	order    []string                 // user IDs in the order first seen
	received map[string]int           // guild ID -> chunks received
	pending  map[string]bool          // guild IDs still awaiting chunks
	done     chan struct{}
}

var (
	collectorMu sync.Mutex
	collector   *sweepCollector
)

func currentCollector() *sweepCollector {
	collectorMu.Lock()
	defer collectorMu.Unlock()
	return collector
}

func setCollector(c *sweepCollector) {
	collectorMu.Lock()
	collector = c
	collectorMu.Unlock()
}

// add records a chunk's members. It reports false once the sweep is no longer collecting.
func (c *sweepCollector) add(mc *discordgo.GuildMembersChunk, cfg ServerConfig) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.pending[mc.GuildID] {
		return false
	}
	for _, m := range mc.Members {
		if m.User == nil || m.User.Bot {
			continue
		}
		if _, seen := c.members[m.User.ID]; !seen {
			c.order = append(c.order, m.User.ID)
		}
		c.members[m.User.ID] = append(c.members[m.User.ID], guildMember{member: m, cfg: cfg})
	}
	c.received[mc.GuildID]++
	if c.received[mc.GuildID] >= mc.ChunkCount {
		c.finishGuild(mc.GuildID)
	}
	return true
}

// finishGuild must be called with c.mu held.
func (c *sweepCollector) finishGuild(guildID string) {
	delete(c.pending, guildID)
	if len(c.pending) == 0 {
		close(c.done)
	}
}

// lookupInterval spreads n lookups across sweepBudget, within the rate bounds.
func lookupInterval(n int) time.Duration {
	if n <= 0 {
		return maxLookupInterval
	}
	d := sweepBudget / time.Duration(n)
	if d < minLookupInterval {
		return minLookupInterval
	}
	if d > maxLookupInterval {
		return maxLookupInterval
	}
	return d
}

// IntervalRefreshAll runs a sweep every refreshInterval. The next sweep is
// scheduled from when the previous one started, and never overlaps it.
func IntervalRefreshAll(ctx context.Context, s *discordgo.Session) {
	for {
		start := time.Now()
		SweepAllGuilds(ctx, s)
		elapsed := time.Since(start)
		wait := refreshInterval - elapsed
		if wait < 0 {
			log.Printf("Sweep took %s, longer than the %s interval; starting the next one now", elapsed.Round(time.Second), refreshInterval)
			wait = 0
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}

// SweepAllGuilds fetches every active guild's members, then looks up each
// unique user once, paced to fit inside sweepBudget, and syncs them in each guild.
func SweepAllGuilds(ctx context.Context, s *discordgo.Session) {
	log.Printf("Fetching all guilds")
	c := &sweepCollector{
		members:  map[string][]guildMember{},
		received: map[string]int{},
		pending:  map[string]bool{},
		done:     make(chan struct{}),
	}
	var guilds []*discordgo.Guild
	for _, guild := range s.State.Guilds {
		cfg := GetServerConfig(guild.ID)
		if cfg != nil && cfg.Active {
			guilds = append(guilds, guild)
			c.pending[guild.ID] = true
		}
	}
	if len(guilds) == 0 {
		return
	}
	setCollector(c)
	for _, guild := range guilds {
		cfg := GetServerConfig(guild.ID)
		if cfg == nil {
			continue
		}
		if err := RequestGuildMembers(s, guild, cfg); err != nil {
			log.Printf("Error while requesting guild members for guild %s (%s): %s\n", guild.ID, cfg.Name, err.Error())
			c.mu.Lock()
			if c.pending[guild.ID] {
				c.finishGuild(guild.ID)
			}
			c.mu.Unlock()
		}
	}

	select {
	case <-c.done:
	case <-time.After(chunkWaitTimeout):
		c.mu.Lock()
		log.Printf("Timed out waiting for member chunks from %d guild(s); continuing with what arrived", len(c.pending))
		c.pending = map[string]bool{}
		c.mu.Unlock()
	case <-ctx.Done():
	}
	setCollector(nil)
	if ctx.Err() != nil {
		return
	}

	c.mu.Lock()
	order, members := c.order, c.members
	c.mu.Unlock()

	interval := lookupInterval(len(order))
	log.Printf("Sweep collected %d unique users across %d guild(s); one lookup every %s (~%s)",
		len(order), len(guilds), interval, (interval * time.Duration(len(order))).Round(time.Second))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for i, userID := range order {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		controller, err := api.GetControllerData(userID)
		if err != nil {
			log.Printf("Error fetching controller data for %s: %s", userID, err.Error())
			continue
		}
		for _, gm := range members[userID] {
			if err := syncMember(s, gm.member, controller, &gm.cfg); err != nil {
				log.Printf("[%s] Error syncing member %s: %s", gm.cfg.Name, userID, err.Error())
			}
		}
		if (i+1)%progressEvery == 0 {
			log.Printf("Sweep progress: %d/%d users", i+1, len(order))
		}
	}
	log.Printf("Sweep finished: %d users", len(order))
}
