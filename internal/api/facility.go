package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type FacilityData struct {
	Id                         string `json:"id"`
	Name                       string `json:"name"`
	AirTrafficManagerCID       uint64 `json:"atm"`
	DeputyAirTrafficManagerCID uint64 `json:"datm"`
	TrainingAdministratorCID   uint64 `json:"ta"`
	EventCoordinatorCID        uint64 `json:"ec"`
	FacilityEngineerCID        uint64 `json:"fe"`
	WebMasterCID               uint64 `json:"wm"`
}

type FacilityListWrapper struct {
	Data []FacilityData `json:"data"`
}

func GetFacilities() ([]FacilityData, error) {
	response, err := Get("/facility/")
	if err != nil {
		return nil, err
	}
	if response.StatusCode == 404 {
		return nil, nil
	}
	if response.StatusCode != 200 {
		return nil, errors.New("HTTP Error when fetching facility data")
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var wrapper FacilityListWrapper
	err = json.Unmarshal(responseData, &wrapper)

	if err != nil {
		return nil, err
	}
	out := make([]FacilityData, len(wrapper.Data))
	for _, value := range wrapper.Data {
		out = append(out, value)
	}
	return out, nil
}
