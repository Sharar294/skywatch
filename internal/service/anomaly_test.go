package service

import "testing"

import "skywatch/internal/domain"

// normalFlight returns a plausible airborne flight used to pad batches so the
// statistical detector has enough samples to work with.
func normalFlight(icao string) domain.Flight {
	return domain.Flight{
		ICAO24:       icao,
		Callsign:     "TEST",
		Longitude:    10.0,
		Latitude:     50.0,
		BaroAltitude: 10000,
		Velocity:     230,
		OnGround:     false,
	}
}

func padBatch(n int) []domain.Flight {
	flights := make([]domain.Flight, 0, n)
	for i := 0; i < n; i++ {
		flights = append(flights, normalFlight("normal"))
	}
	return flights
}

func hasReason(anomalies []domain.Anomaly, reason string) bool {
	for _, a := range anomalies {
		if a.Reason == reason {
			return true
		}
	}
	return false
}

func TestImpossibleVelocity(t *testing.T) {
	d := NewDetector()
	batch := padBatch(10)
	bad := normalFlight("fast")
	bad.Velocity = 900 // Mach-3-ish, impossible
	batch = append(batch, bad)

	anomalies, stats := d.Detect(batch)
	if !hasReason(anomalies, "impossible_velocity") {
		t.Fatalf("expected impossible_velocity anomaly, got %+v", anomalies)
	}
	if stats.AnomalyCount != len(anomalies) {
		t.Fatalf("stats.AnomalyCount %d != len(anomalies) %d", stats.AnomalyCount, len(anomalies))
	}
}

func TestImpossibleAltitude(t *testing.T) {
	d := NewDetector()
	batch := padBatch(10)
	bad := normalFlight("high")
	bad.BaroAltitude = 20000 // above any service ceiling
	batch = append(batch, bad)

	anomalies, _ := d.Detect(batch)
	if !hasReason(anomalies, "impossible_altitude") {
		t.Fatalf("expected impossible_altitude anomaly, got %+v", anomalies)
	}
}

func TestOnGroundButFast(t *testing.T) {
	d := NewDetector()
	batch := padBatch(10)
	bad := normalFlight("taxi")
	bad.OnGround = true
	bad.Velocity = 200
	batch = append(batch, bad)

	anomalies, _ := d.Detect(batch)
	if !hasReason(anomalies, "onground_but_fast") {
		t.Fatalf("expected onground_but_fast anomaly, got %+v", anomalies)
	}
}

func TestVelocityOutlier(t *testing.T) {
	d := NewDetector()
	// A tight cluster around 230 m/s, then one plausible-but-far outlier.
	batch := padBatch(20)
	bad := normalFlight("outlier")
	bad.Velocity = 380 // within physical limits but far from the mean
	batch = append(batch, bad)

	anomalies, _ := d.Detect(batch)
	if !hasReason(anomalies, "velocity_outlier") {
		t.Fatalf("expected velocity_outlier anomaly, got %+v", anomalies)
	}
}

func TestNoFalsePositivesOnCleanBatch(t *testing.T) {
	d := NewDetector()
	anomalies, stats := d.Detect(padBatch(30))
	if len(anomalies) != 0 {
		t.Fatalf("expected no anomalies on a uniform clean batch, got %+v", anomalies)
	}
	if stats.AirborneCount != 30 {
		t.Fatalf("expected 30 airborne, got %d", stats.AirborneCount)
	}
}

func TestSmallBatchSkipsStatistics(t *testing.T) {
	d := NewDetector()
	// Below MinAirborneStat: a velocity outlier should NOT be flagged
	// statistically (too few samples to trust), only physical rules apply.
	batch := padBatch(3)
	bad := normalFlight("outlier")
	bad.Velocity = 380
	batch = append(batch, bad)

	anomalies, _ := d.Detect(batch)
	if hasReason(anomalies, "velocity_outlier") {
		t.Fatalf("did not expect statistical outlier on tiny batch, got %+v", anomalies)
	}
}
