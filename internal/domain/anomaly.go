package domain

// Anomaly represents a single flagged flight vector along with the reason it
// was considered anomalous. It is the output of the baseline detector that runs
// inside the Analyzer service.
type Anomaly struct {
	ICAO24        string  `json:"icao24"`
	Callsign      string  `json:"callsign"`
	OriginCountry string  `json:"origin_country"`
	Reason        string  `json:"reason"`  // machine-readable code, e.g. "velocity_outlier"
	Detail        string  `json:"detail"`  // human-readable explanation
	Value         float64 `json:"value"`   // the offending measurement
	ZScore        float64 `json:"z_score"` // z-score when the reason is statistical (else 0)
}

// BatchStats captures the summary statistics computed for one batch of flight
// vectors. It is useful both for the statistical detector and for observability.
type BatchStats struct {
	TotalFlights   int     `json:"total_flights"`
	AirborneCount  int     `json:"airborne_count"`
	AnomalyCount   int     `json:"anomaly_count"`
	MeanVelocity   float64 `json:"mean_velocity"`
	StdDevVelocity float64 `json:"stddev_velocity"`
	MeanAltitude   float64 `json:"mean_altitude"`
	StdDevAltitude float64 `json:"stddev_altitude"`
}
