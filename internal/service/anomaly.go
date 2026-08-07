package service

import (
	"fmt"
	"math"

	"skywatch/internal/domain"
)

// Detector holds the tunable thresholds for the baseline anomaly detector.
//
// The baseline combines two complementary strategies:
//
//  1. Physical rule checks — measurements that are impossible or nonsensical for
//     real air traffic (e.g. an airliner reporting Mach-3 ground speed, or a
//     flight marked "on ground" while moving at 300 kt). These do not depend on
//     the rest of the batch.
//
//  2. Statistical outlier detection — within a single batch we compute the mean
//     and standard deviation of velocity and altitude for airborne flights, then
//     flag anything beyond ZThreshold standard deviations. This adapts to
//     whatever traffic mix happens to be in the sky at that moment.
//
// This is deliberately simple and explainable — a foundation to iterate on
// before reaching for a trained model.
type Detector struct {
	ZThreshold      float64 // std-devs from the mean to flag a statistical outlier
	MaxVelocity     float64 // m/s, above this is physically implausible
	MaxAltitude     float64 // m, above this is above any service ceiling
	MinAltitude     float64 // m, below this (negative) is implausible
	GroundMaxVel    float64 // m/s, faster than this while "on ground" is contradictory
	MinAirborneStat int     // need at least this many airborne flights to trust z-scores
}

// NewDetector returns a Detector pre-loaded with sensible defaults for
// commercial + general aviation traffic from the OpenSky feed.
func NewDetector() *Detector {
	return &Detector{
		ZThreshold:      3.0,
		MaxVelocity:     400.0,  // ~778 kt; airliners cruise ~250 m/s, so 400 leaves margin
		MaxAltitude:     16000.0, // ~52,500 ft; above business-jet ceilings
		MinAltitude:     -500.0,  // allow for pressure/geoid noise near sea level
		GroundMaxVel:    120.0,   // ~230 kt; no aircraft taxis this fast
		MinAirborneStat: 8,       // z-scores are meaningless on tiny samples
	}
}

// Detect runs the baseline over one batch of flights and returns the anomalies
// found plus summary statistics for the batch.
func (d *Detector) Detect(flights []domain.Flight) ([]domain.Anomaly, domain.BatchStats) {
	anomalies := make([]domain.Anomaly, 0)

	// Collect airborne flights with a usable position; these drive the stats.
	airborne := make([]domain.Flight, 0, len(flights))
	for _, f := range flights {
		if !f.OnGround && (f.Longitude != 0 || f.Latitude != 0) {
			airborne = append(airborne, f)
		}
	}

	meanVel, stdVel := meanStdDev(airborne, func(f domain.Flight) float64 { return f.Velocity })
	meanAlt, stdAlt := meanStdDev(airborne, func(f domain.Flight) float64 { return f.BaroAltitude })

	stats := domain.BatchStats{
		TotalFlights:   len(flights),
		AirborneCount:  len(airborne),
		MeanVelocity:   meanVel,
		StdDevVelocity: stdVel,
		MeanAltitude:   meanAlt,
		StdDevAltitude: stdAlt,
	}

	useStats := len(airborne) >= d.MinAirborneStat

	for _, f := range flights {
		// --- Physical rule checks (batch-independent) ---
		if f.Velocity > d.MaxVelocity || f.Velocity < 0 {
			anomalies = append(anomalies, newAnomaly(f, "impossible_velocity",
				fmt.Sprintf("velocity %.0f m/s outside plausible range", f.Velocity), f.Velocity, 0))
			continue
		}
		if f.BaroAltitude > d.MaxAltitude || f.BaroAltitude < d.MinAltitude {
			anomalies = append(anomalies, newAnomaly(f, "impossible_altitude",
				fmt.Sprintf("altitude %.0f m outside plausible range", f.BaroAltitude), f.BaroAltitude, 0))
			continue
		}
		if f.OnGround && f.Velocity > d.GroundMaxVel {
			anomalies = append(anomalies, newAnomaly(f, "onground_but_fast",
				fmt.Sprintf("reported on ground while moving at %.0f m/s", f.Velocity), f.Velocity, 0))
			continue
		}

		// --- Statistical outlier checks (airborne only) ---
		if !useStats || f.OnGround {
			continue
		}
		if stdVel > 0 {
			z := (f.Velocity - meanVel) / stdVel
			if math.Abs(z) > d.ZThreshold {
				anomalies = append(anomalies, newAnomaly(f, "velocity_outlier",
					fmt.Sprintf("velocity %.0f m/s is %.1fσ from batch mean %.0f", f.Velocity, z, meanVel), f.Velocity, z))
				continue
			}
		}
		if stdAlt > 0 {
			z := (f.BaroAltitude - meanAlt) / stdAlt
			if math.Abs(z) > d.ZThreshold {
				anomalies = append(anomalies, newAnomaly(f, "altitude_outlier",
					fmt.Sprintf("altitude %.0f m is %.1fσ from batch mean %.0f", f.BaroAltitude, z, meanAlt), f.BaroAltitude, z))
				continue
			}
		}
	}

	stats.AnomalyCount = len(anomalies)
	return anomalies, stats
}

func newAnomaly(f domain.Flight, reason, detail string, value, z float64) domain.Anomaly {
	return domain.Anomaly{
		ICAO24:        f.ICAO24,
		Callsign:      f.Callsign,
		OriginCountry: f.OriginCountry,
		Reason:        reason,
		Detail:        detail,
		Value:         value,
		ZScore:        z,
	}
}

// meanStdDev computes the population mean and standard deviation of a field
// extracted from the given flights. Returns (0, 0) for an empty slice.
func meanStdDev(flights []domain.Flight, field func(domain.Flight) float64) (float64, float64) {
	n := len(flights)
	if n == 0 {
		return 0, 0
	}
	var sum float64
	for _, f := range flights {
		sum += field(f)
	}
	mean := sum / float64(n)

	var variance float64
	for _, f := range flights {
		diff := field(f) - mean
		variance += diff * diff
	}
	variance /= float64(n)
	return mean, math.Sqrt(variance)
}
