package service

import (
	"skywatch/internal/domain"
)

// OpenSkyResponse matches the top-level JSON structure from the API.
type OpenSkyResponse struct {
	Time   int64           `json:"time"`
	States [][]interface{} `json:"states"`
}

// MapToFlights converts the raw 2D 'any' array into a slice of Flight structs.
func MapToFlights(states [][]interface{}) []domain.Flight {
	flights := make([]domain.Flight, 0, len(states))

	for _, s := range states {
		// OpenSky indices: 0:icao24, 1:callsign, 2:origin, 5:long, 6:lat, 7:baro_alt, 8:on_ground, 9:velocity
		f := domain.Flight{
			ICAO24:        toString(s[0]),
			Callsign:      toString(s[1]),
			OriginCountry: toString(s[2]),
			Longitude:     toFloat(s[5]),
			Latitude:      toFloat(s[6]),
			BaroAltitude:  toFloat(s[7]),
			OnGround:      toBool(s[8]),
			Velocity:      toFloat(s[9]),
		}
		flights = append(flights, f)
	}
	return flights
}

// Helper assertions to prevent panics on nil or unexpected types
func toString(v interface{}) string {
	if v == nil { return "" }
	s, _ := v.(string)
	return s
}

func toFloat(v interface{}) float64 {
	if v == nil { return 0.0 }
	f, _ := v.(float64)
	return f
}

func toBool(v interface{}) bool {
	if v == nil { return false }
	b, _ := v.(bool)
	return b
}