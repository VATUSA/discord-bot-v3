package bot

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/VATUSA/discord-bot-v3/internal/api"
	"github.com/VATUSA/discord-bot-v3/pkg/constants"
	"github.com/bwmarrin/discordgo"
)

// maxNicknameLength is Discord's nickname limit, counted in characters.
const maxNicknameLength = 32

// shortenNickname fits name+suffix into Discord's nickname limit by trying
// progressively shorter forms of the name: first and last, first and last
// initial, first name only, then a hard truncation of the first name.
func shortenNickname(name, suffix string) string {
	fits := func(n string) bool {
		return utf8.RuneCountInString(n+suffix) <= maxNicknameLength
	}
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return name + suffix
	}
	first, last := parts[0], parts[len(parts)-1]
	candidates := []string{first}
	if len(parts) > 1 {
		initial, _ := utf8.DecodeRuneInString(last)
		candidates = []string{
			first + " " + last,
			first + " " + string(initial),
			first,
		}
	}
	for _, c := range candidates {
		if fits(c) {
			return c + suffix
		}
	}
	// Even the first name doesn't fit; truncate it, dropping the suffix if it
	// alone leaves no room.
	room := maxNicknameLength - utf8.RuneCountInString(suffix)
	if room < 1 {
		room, suffix = maxNicknameLength, ""
	}
	r := []rune(first)
	if len(r) > room {
		r = r[:room]
	}
	return string(r) + suffix
}

func SyncName(s *discordgo.Session, m *discordgo.Member, c *api.ControllerData, cfg *ServerConfig) error {
	if c == nil {
		if m.Nick != "" {
			log.Printf("[%s] Nickname Removed %s for ID %s", cfg.Name, m.Nick, m.User.ID)
			err := s.GuildMemberNickname(m.GuildID, m.User.ID, "")
			if err != nil {
				return err
			}
		}
		return nil
	}
	name, err := CalculateName(c, cfg)
	if err != nil {
		return err
	}
	if name == "" {
		return nil
	}
	title, err := CalculateTitle(c, cfg)
	if err != nil {
		return nil
	}
	var suffix string
	if strings.HasSuffix(m.Nick, "| VATGOV") {
		suffix = " | VATGOV"
	} else if title != "" {
		suffix = " | " + title
	}
	prospect := name + suffix
	if utf8.RuneCountInString(prospect) > maxNicknameLength {
		oldProspect := prospect
		prospect = shortenNickname(name, suffix)
		log.Printf("[%s] Prospective nickname too long %s - Shortened to %s", cfg.Name, oldProspect, prospect)
	}
	if prospect != m.Nick {
		log.Printf("[%s] Nickname Change %s -> %s for ID %s", cfg.Name, m.Nick, prospect, m.User.ID)
		err := s.GuildMemberNickname(m.GuildID, m.User.ID, prospect)
		if err != nil {
			return err
		}
	}
	return nil
}

func CalculateName(c *api.ControllerData, cfg *ServerConfig) (string, error) {
	switch cfg.NameFormatType {
	case constants.NameFormat_None:
		return "", nil
	case constants.NameFormat_FirstLast:
		if c.FlagNamePrivacy {
			return fmt.Sprintf("%s %d", c.FirstName, c.CID), nil
		}
		return fmt.Sprintf("%s %s", c.FirstName, c.LastName), nil

	case constants.NameFormat_FirstL:
		if c.FlagNamePrivacy {
			return fmt.Sprintf("%s", c.FirstName), nil
		}
		return fmt.Sprintf("%s %c", c.FirstName, c.LastName[0]), nil
	case constants.NameFormat_CertificateID:
		return fmt.Sprintf("%d", c.CID), nil
	default:
		return "", errors.New("invalid NameFormat")
	}
}

func CalculateTitle(c *api.ControllerData, cfg *ServerConfig) (string, error) {
	switch cfg.TitleType {
	case constants.Title_Division:
		return CalculateDivisionTitle(c, cfg), nil
	case constants.Title_Local:
		return CalculateLocalTitle(c, cfg), nil
	case constants.Title_None:
		return "", nil
	case constants.Title_Rating:
		return c.RatingShort, nil
	default:
		return "", errors.New("invalid TitleFormat")
	}
}

func facilityStaffTitle(c *api.ControllerData) (string, bool) {
	facility := GetFacilityData(c.Facility)
	if facility.AirTrafficManagerCID == c.CID {
		return "ATM", true
	}
	if facility.DeputyAirTrafficManagerCID == c.CID {
		return "DATM", true
	}
	if facility.TrainingAdministratorCID == c.CID {
		return "TA", true
	}
	if facility.EventCoordinatorCID == c.CID {
		return "EC", true
	}
	if facility.FacilityEngineerCID == c.CID {
		return "FE", true
	}
	if facility.WebMasterCID == c.CID {
		return "WM", true
	}

	return "", false
}

func CalculateDivisionTitle(c *api.ControllerData, cfg *ServerConfig) string {
	for _, r := range c.Roles {
		if strings.HasPrefix(r.Role, "US") {
			re := regexp.MustCompile("[0-9]+")
			match := re.FindString(r.Role)
			if match == "0" {
				return "VATUSA"
			}
			if match != "" {
				return fmt.Sprintf("VATUSA%s", match)
			}
		}
	}
	roleTitle, hasRoleTitle := facilityStaffTitle(c)
	if hasRoleTitle {
		return fmt.Sprintf("%s %s", c.Facility, roleTitle)
	}
	if c.Facility == "ZZN" {
		return fmt.Sprintf("%s", c.RatingShort)
	} else if c.Facility == "ZAE" {
		return "ZAE"
	} else if c.Rating < 1 {
		return ""
	} else {
		return fmt.Sprintf("%s %s", c.Facility, c.RatingShort)
	}
}

func CalculateLocalTitle(c *api.ControllerData, cfg *ServerConfig) string {
	for _, r := range c.Roles {
		if strings.HasPrefix(r.Role, "US") {
			re := regexp.MustCompile("[0-9]+")
			match := re.FindString(r.Role)
			if match == "0" {
				return "VATUSA"
			}
			if match != "" {
				return fmt.Sprintf("VATUSA%s", match)
			}
		}
	}
	roleTitle, hasRoleTitle := facilityStaffTitle(c)
	if hasRoleTitle && c.Facility == cfg.Facility {
		return fmt.Sprintf("%s", roleTitle)
	} else if hasRoleTitle {
		return fmt.Sprintf("%s %s", c.Facility, roleTitle)
	}
	if c.Facility == "ZZN" {
		return fmt.Sprintf("%s", c.RatingShort)
	} else if c.Facility == "ZAE" {
		return "ZAE"
	} else if c.Rating < 1 {
		return ""
	} else if c.Facility == cfg.Facility {
		return c.RatingShort
	} else {
		return fmt.Sprintf("%s %s", c.Facility, c.RatingShort)
	}
}
