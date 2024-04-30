package tripping

import (
	"sort"
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
