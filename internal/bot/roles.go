package bot

import (
	"github.com/VATUSA/discord-bot-v3/internal/api"
	"github.com/VATUSA/discord-bot-v3/pkg/constants"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slices"
	"log"
	"regexp"
	"strings"
)

func SyncRoles(s *discordgo.Session, m *discordgo.Member, c *api.ControllerData, cfg *ServerConfig) error {
	for _, role := range cfg.Roles {
		assigned := false
		roleDisplay := role.ID
		if role.Name != "" {
			roleDisplay = role.Name
		}
		for _, criteria := range role.Criteria {
			if checkCriteria(c, &criteria) {
				if !slices.Contains(m.Roles, role.ID) {
					log.Printf("[%s] Add role %s to member %s %s", cfg.Name, roleDisplay, m.Nick, m.User.ID)
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
				log.Printf("[%s] Remove role %s from member %s %s", cfg.Name, roleDisplay, m.Nick, m.User.ID)
				err := s.GuildMemberRoleRemove(m.GuildID, m.User.ID, role.ID)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkCriteria(c *api.ControllerData, criteria *CriteriaConfig) bool {
	if c == nil {
		return false
	}
	if c.Rating < 1 {
		return false
	}
	for _, cond := range criteria.Conditions {
		if !checkConditionWithInvert(c, &cond) {
			return false
		}
	}
	return true
}

func checkConditionWithInvert(c *api.ControllerData, cond *ConditionConfig) bool {
	ret := checkCondition(c, cond.Type, cond.Value)
	if cond.Invert {
		return !ret
	} else {
		return ret
	}
}

func checkCondition(c *api.ControllerData, condType constants.ConditionType, value *string) bool {
	switch condType {
	case constants.Condition_All:
		return true
	case constants.Condition_InDivision:
		return *value == "true" == c.FlagHomeController
	case constants.Condition_DivisionVisitor:
		return *value == "true" == (len(c.VisitingFacilities) > 0)
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
		if value == nil {
			log.Print("Invalid facility_role condition: missing value")
			return false
		}
		facility, position, scope, ok := parseFacilityRoleValue(*value)
		if !ok {
			log.Printf("Invalid facility_role condition value %q", *value)
			return false
		}
		if scope == facilityRoleScopeAny {
			return holdsFacilityRole(c, facility, position)
		}
		poc, known := FacilityPOC(facility, position)
		if !known {
			log.Printf("Invalid facility_role condition value %q: %s is not a staff position", *value, position)
			return false
		}
		if scope == facilityRoleScopePOC {
			return poc != 0 && poc == c.CID
		}
		return holdsFacilityRole(c, facility, position) && c.CID != poc
	default:
		log.Printf("Invalid RoleConditionCriteriaType %q", condType)
		return false

	}
}

type facilityRoleScope int

const (
	facilityRoleScopeAny  facilityRoleScope = iota // "ZZZ:WM" - anyone holding the role
	facilityRoleScopePOC                           // "ZZZ:WM:POC" - the facility's point of contact
	facilityRoleScopeTeam                          // "ZZZ:WM:TEAM" - role holders other than the POC
)

func parseFacilityRoleValue(value string) (facility string, position string, scope facilityRoleScope, ok bool) {
	parts := strings.Split(value, ":")
	switch len(parts) {
	case 2:
		return parts[0], parts[1], facilityRoleScopeAny, true
	case 3:
		switch strings.ToUpper(parts[2]) {
		case "POC":
			return parts[0], parts[1], facilityRoleScopePOC, true
		case "TEAM":
			return parts[0], parts[1], facilityRoleScopeTeam, true
		}
	}
	return "", "", facilityRoleScopeAny, false
}

func holdsFacilityRole(c *api.ControllerData, facility string, position string) bool {
	for _, r := range c.Roles {
		if r.Facility == facility && r.Role == position {
			return true
		}
	}
	return false
}
