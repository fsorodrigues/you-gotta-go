package tripping

import (
	"fmt"
	"sort"
	"time"
  "strconv"
	utils "you-gotta-go/parser/utils"
)

func GetNextTrip(data []utils.Trip) utils.Trip {
  d := data

  sort.Slice(d, func(i, j int) bool { 
    ii := d[i].Arrival.Aimed.Time
    jj := d[j].Arrival.Aimed.Time

    return ii.Before(jj)
  })
  
  return d[0]
}

func formatTime(d time.Duration) string {
  return strconv.FormatFloat(d.Seconds() / 60, 'f', 0, 64)
}

func decideScenario(trip utils.Trip) string {
  if !trip.Monitored { return "not-tracking" }
  if trip.Status == "cancelled" { return "cancelled" }
  if trip.Status == "delayed" { return "delayed" }
  if trip.Status == "early" { return "early" }
  return "on-time"
}

func CreateMessageFromTrip(scn string, trip utils.Trip) string {
  var diff time.Duration
  now := time.Now()

  switch scn {
  case "not-tracking":
    diff = trip.Arrival.Aimed.Time.Sub(now)
    return fmt.Sprintf("%s: %sm nt", trip.ServiceID, formatTime(diff))
  case "cancelled":
   return fmt.Sprintf("%s: CANC", trip.ServiceID)
  case "delayed":
    diff = trip.Arrival.Expected.Time.Sub(now)
    return fmt.Sprintf("%s: %s del", trip.ServiceID, formatTime(diff))
  case "early":
    diff = trip.Arrival.Expected.Time.Sub(now)
    return fmt.Sprintf("%s: %sm ear", trip.ServiceID, formatTime(diff))
  case "on-time":
    diff = trip.Arrival.Expected.Time.Sub(now)
    return fmt.Sprintf("%s: %sm", trip.ServiceID, formatTime(diff))
  default:
    return "ERROR. Scenario not recognized"
  }
 
}

func ParseTrip(trip utils.Trip, msg *string) {
  scenario := decideScenario(trip)
  *msg = CreateMessageFromTrip(scenario, trip) 
}
