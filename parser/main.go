package main

import (
	"fmt"
	trip "you-gotta-go/parser/tripping"
	utils "you-gotta-go/parser/utils"
)

func main() {
  var data utils.InputData = utils.Read()
  var trips []utils.Trip = utils.FilterByService(data.Trips, "23")

  if data.Closed || len(data.Trips) < 1 {
  }

}
