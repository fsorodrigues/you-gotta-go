package main 

import (
  "fmt"
  "time"
  "os"
  utils "you-gotta-go/parser/utils"
  trip "you-gotta-go/parser/tripping"
)

func main() {
  var data utils.InputData = utils.Read()
  var trips []utils.Trip = utils.FilterByService(data.Trips, "23")
  var message string

  if data.Closed || len(data.Trips) < 1 {
    message = "No trips available"
    fmt.Println(message)
    os.Exit(0)
  }

  var NextTrip utils.Trip = trip.GetNextTrip(trips)
  now := time.Now()

  if !NextTrip.Monitored {
    message = "Bus not tracking. Gotta use the aimed time"
    fmt.Println(message)
    os.Exit(0)
  }
  
  message = "%s: %s"
  fmt.Println(fmt.Sprintf(
    message, 
    NextTrip.ServiceID,
    NextTrip.Arrival.Expected.Time.Sub(now),
  ))
}
