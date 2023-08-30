package bot

import (
	"fmt"
	"github.com/VATUSA/discord-bot-v3/internal/config"
	"github.com/VATUSA/discord-bot-v3/internal/integration/api2"
	"github.com/VATUSA/discord-bot-v3/pkg/constants"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slices"
	"log"
	"regexp"
	"strings"
)

func SyncRoles(s *discordgo.Session, m *discordgo.Member, c *api2.ControllerData, cfg *config.ServerConfig) error {
	for _, role := range cfg.Roles {
		assigned := false
		roleDisplay := role.ID
		if role.Name != "" {
			roleDisplay = role.Name
		}
		for _, criteria := range role.Criteria {
			if checkCriteria(c, &criteria) {
				if !slices.Contains(m.Roles, role.ID) {
					log.Printf("Add role %s to member %s %s", roleDisplay, m.Nick, m.User.ID)
					err := s.GuildMemberRoleAdd(m.GuildID, m.User.ID, role.ID)
					if err != nil {
						return err
					}
				}
				assigned = true
				break
			}
		}
		if !assigned {
			if slices.Contains(m.Roles, role.ID) {
				log.Printf("Remove role %s from member %s %s", roleDisplay, m.Nick, m.User.ID)
				err := s.GuildMemberRoleRemove(m.GuildID, m.User.ID, role.ID)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkCriteria(c *api2.ControllerData, criteria *config.CriteriaConfig) bool {
	if c == nil {
		return false
	}
	if c.Rating < 1 {
		return false
	}
	for _, cond := range criteria.Conditions {
		if !checkCondition(c, cond.Type, cond.Value) {
			return false
		}
	}
	return true
}

func checkCondition(c *api2.ControllerData, condType constants.ConditionType, value *string) bool {
	switch condType {
	case constants.Condition_All:
		return true
	case constants.Condition_InDivision:
		return *value == "true" == c.FlagHomeController
	case constants.Condition_HomeFacility:
		return c.Facility == *value
	case constants.Condition_VisitFacility:
		for _, v := range c.VisitingFacilities {
			if v.Facility == *value {
				return true
			}
		}
		return false
	case constants.Condition_HomeOrVisit:
		if c.Facility == *value {
			return true
		}
		for _, v := range c.VisitingFacilities {
			if v.Facility == *value {
				return true
			}
		}
		return false
	case constants.Condition_Rating:
		return c.RatingShort == *value
	case constants.Condition_DivisionStaff:
		for _, r := range c.Roles {
			if strings.HasPrefix(r.Role, "US") {
				re := regexp.MustCompile("[0-9]+")
				match := re.FindString(r.Role)
				if match != "" {
					return true
				}
			}
		}
		return false
	case constants.Condition_FacilityStaff:
		for _, r := range c.Roles {
			re := regexp.MustCompile("ATM|DATM|TA|FE|EC|WM")
			if re.MatchString(r.Role) && r.Facility == *value {
				return true
			}
		}
		return false
	case constants.Condition_Role:
		for _, r := range c.Roles {
			if r.Role == *value {
				return true
			}
		}
		return false
	case constants.Condition_FacilityRole:
		for _, r := range c.Roles {
			if *value == fmt.Sprintf("%s:%s", r.Facility, r.Role) {
				return true
			}
		}
		return false
	default:
		log.Printf("Invalid RoleConditionCriteriaType %d", condType)
		return false

	}
}
