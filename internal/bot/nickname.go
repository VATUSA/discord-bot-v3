package bot

import (
	"errors"
	"fmt"
	"github.com/VATUSA/discord-bot-v3/internal/api"
	"github.com/VATUSA/discord-bot-v3/pkg/constants"
	"github.com/bwmarrin/discordgo"
	"log"
	"regexp"
	"strings"
)

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
	var prospect string
	if strings.HasSuffix(m.Nick, "| VATGOV") {
		prospect = fmt.Sprintf("%s | VATGOV", name)
	} else if title != "" {
		prospect = fmt.Sprintf("%s | %s", name, title)
	} else {
		prospect = name
	}
	if len(prospect) > 32 {
		oldProspect := prospect
		nameParts := strings.SplitN(name, " ", -1)
		prospect = fmt.Sprintf("%s %s | %s", nameParts[0], nameParts[len(nameParts)-1], title)
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
		return fmt.Sprintf("%s %s", c.FirstName, c.LastName[0]), nil
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
	} else if c.Facility == cfg.Facility {
		return c.RatingShort
	} else {
		return fmt.Sprintf("%s %s", c.Facility, c.RatingShort)
	}
}
