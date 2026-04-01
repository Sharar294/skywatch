package domain

// Flight represents the processed aviation telemetry.
type Flight struct {
	ICAO24         string  `json:"icao24"`
	Callsign       string  `json:"callsign"`
	OriginCountry  string  `json:"origin_country"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	BaroAltitude   float64 `json:"baro_altitude"`
	Velocity       float64 `json:"velocity"`
	OnGround       bool    `json:"on_ground"`
	// LastContact    int64   `json:"last_contact"`
}