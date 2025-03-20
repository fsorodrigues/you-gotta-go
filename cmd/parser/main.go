package parser

import (
	"encoding/json"
	"log"
	trip "you-gotta-go/cmd/parser/tripping"
	utils "you-gotta-go/cmd/parser/utils"
)

func Unmarshal(dataIn []byte) utils.InputData {
	var data utils.InputData

	jsonErr := json.Unmarshal(dataIn, &data)
	if jsonErr != nil {
		log.Fatalln("Error parsing JSON input")
	}

	return data
}

func Parse(data utils.InputData, service string) *string {
	var trips []utils.Trip = utils.FilterByService(data.Trips, service)
	var message *string = new(string)

	if data.Closed || len(data.Trips) < 1 {
		*message = "No trips available"
	} else {
		var NextTrip utils.Trip = trip.GetNextTrip(trips)
		trip.ParseTrip(NextTrip, message)
	}

	return message
}
