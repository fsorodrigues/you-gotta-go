package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	trip "you-gotta-go/cmd/parser/trip"
	utils "you-gotta-go/cmd/parser/utils"
)

type ParserError struct {
	Err error
}

func (d ParserError) Error() string {
	return fmt.Sprintf("Error parsing message %v", d.Err.Error())
}

func Unmarshal(dataIn []byte) (utils.InputData, error) {
	var data utils.InputData

	jsonErr := json.Unmarshal(dataIn, &data)
	if jsonErr != nil {
		errMsg := fmt.Sprintf("Error parsing JSON input. %v\n", jsonErr)
		return utils.InputData{}, ParserError{
			Err: errors.New(errMsg),
		}
	}

	return data, nil
}

func Parse(data utils.InputData, service string) (*string, error) {
	var trips []utils.Trip = utils.FilterByService(data.Trips, service)
	var message *string = new(string)

	if data.Closed {
		*message = "Stop closed"
	} else if len(data.Trips) < 1 {
		*message = "No trips available"
	} else {
		NextTrip, err := trip.GetNextTrip(trips)
		if err != nil {
			errMsg := fmt.Sprintf("Can't get next trip. %v\n", err)
			return nil, ParserError{
				Err: errors.New(errMsg),
			}
		}
		trip.ParseTrip(NextTrip, message)
	}

	return message, nil
}
