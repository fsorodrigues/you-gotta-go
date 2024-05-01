package utils

func FilterByService(trips []Trip, service string) []Trip {
  arr_map := make(map[string][]Trip)
  for _, trip := range trips {
    arr_map[trip.ServiceID] = append(arr_map[trip.ServiceID], trip)
  }

  return arr_map[service]
} 
