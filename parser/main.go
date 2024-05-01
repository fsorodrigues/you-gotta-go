package main

import (
	"fmt"
	trip "you-gotta-go/parser/tripping"
	utils "you-gotta-go/parser/utils"
)

func main() {
  var data utils.InputData = utils.Read()
  var trips []utils.Trip = utils.FilterByService(data.Trips, "23")
  var message *string = new(string)

  if data.Closed || len(data.Trips) < 1 {
    *message = "No trips available"
  } else {
    var NextTrip utils.Trip = trip.GetNextTrip(trips)
    trip.ParseTrip(NextTrip, message)
  }

  fmt.Println(*message)
}
