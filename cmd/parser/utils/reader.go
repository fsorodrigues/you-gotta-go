package utils 

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"os"
	"time"
  "strings"
)

type NullTime sql.NullTime

func (d *NullTime) UnmarshalJSON(data []byte) error {
  s := string(data)
  s = strings.ReplaceAll(s, "\"", "")

  x, err := time.Parse(time.RFC3339, s)
  if err != nil {
    d.Valid = false
    return nil
  }

  d.Time = x
  d.Valid = true
  return nil
}

type Stop struct {
  Stop_id string `json:"stop_id"`
  Name string `json:"name"`
}

type Prediction struct {
  Aimed NullTime `json:"aimed"`
  Expected NullTime `json:"expected"`
}

type Trip struct {
  StopID string `json:"stop_id"`
  ServiceID string `json:"service_id"`
  Direction string `json:"direction"`
  Operator string `json:"operator"`
  Origin Stop `json:"origin"`
  Destination Stop `json:"destination"`
  Delay string `json:"delay"`
  VehicleID string `json:"vehicle_id"`
  Name string `json:"name"`
  Arrival Prediction `json:"arrival"`
  Departure Prediction `json:"departure"`
  Status string `json:"status"`
  Monitored bool `json:"monitored"`
  Wheelchair_accessible bool `json:"wheelchair_accessible"`
  TripID string `json:"trip_id"`
}

type InputData struct {
  Farezone string `json:"farezone"`
  Closed bool `json:"closed"`
  Trips []Trip `json:"departures"`
}

func Read() InputData {
  stdin, err := io.ReadAll(os.Stdin)
  if err != nil {
    log.Fatalln(err)
  }

  var d InputData

  jsonErr := json.Unmarshal(stdin, &d)
  if jsonErr != nil {
    log.Fatalln("Error parsing JSON input")
  }
  return d
}
