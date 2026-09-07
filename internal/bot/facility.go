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
	facility := FacilityDataByIdMap[strings.ToUpper(facilityId)]
	return facility
}

var facilityPOCByPosition = map[string]func(api.FacilityData) uint64{
	"ATM":  func(f api.FacilityData) uint64 { return f.AirTrafficManagerCID },
	"DATM": func(f api.FacilityData) uint64 { return f.DeputyAirTrafficManagerCID },
	"TA":   func(f api.FacilityData) uint64 { return f.TrainingAdministratorCID },
	"EC":   func(f api.FacilityData) uint64 { return f.EventCoordinatorCID },
	"FE":   func(f api.FacilityData) uint64 { return f.FacilityEngineerCID },
	"WM":   func(f api.FacilityData) uint64 { return f.WebMasterCID },
}

func FacilityPOC(facilityId string, position string) (uint64, bool) {
	get, ok := facilityPOCByPosition[strings.ToUpper(position)]
	if !ok {
		return 0, false
	}
	return get(GetFacilityData(facilityId)), true
}
