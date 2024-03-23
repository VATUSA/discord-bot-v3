package bot

import (
	"github.com/VATUSA/discord-bot-v3/internal/api"
	"log"
	"strings"
	"time"
)

var FacilityDataByIdMap = make(map[string]api.FacilityData)
var LastFacilityDataLoad *time.Time = nil

const FacilityDataCacheDuration = 5 * time.Minute

func LoadFacilityData() {
	facilityData, err := api.GetFacilities()
	if err != nil {
		log.Printf("Error fetching facility data: %v", err)
		return
	}
	for _, facility := range facilityData {
		FacilityDataByIdMap[strings.ToUpper(facility.Id)] = facility
	}
}

func GetFacilityData(facilityId string) api.FacilityData {
	if LastFacilityDataLoad == nil || time.Now().After(LastFacilityDataLoad.Add(FacilityDataCacheDuration)) {
		LoadFacilityData()
		t := time.Now()
		LastFacilityDataLoad = &t
	}
	facility := FacilityDataByIdMap[facilityId]
	return facility
}
