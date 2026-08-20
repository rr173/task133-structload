package loadcode

import "math"

// SnowLoadCenti computes the flat/gabled roof snow load in centi-kPa.
// pf = 0.7 * Ce * Ct * Is * pg (flat roof); ps = pf * Cs (sloped).
// Storage formula: pf_centi = round_half_up(0.7 * Ce * Ct * Is * pg_centi)
//                  ps_centi = round_half_up(pf_centi * Cs).
// Ce/Ct are centi factors; Is is the importance factor (float); pg is centi-kPa.
func SnowLoadCenti(ceCenti, ctCenti int64, importanceFactor float64, pgCenti int64, roofSlopeDeciDeg int64) (pfCenti, psCenti int64) {
	Ce := float64(ceCenti) / 100.0
	Ct := float64(ctCenti) / 100.0
	pg := float64(pgCenti) / 100.0
	pfKPa := 0.7 * Ce * Ct * importanceFactor * pg
	pfCenti = RoundHalfUp(pfKPa * 100.0)
	Cs := SnowSlopeFactor(roofSlopeDeciDeg)
	psCenti = RoundHalfUp(float64(pfCenti) * Cs)
	return pfCenti, psCenti
}

// SnowSlopeFactor returns Cs for a warm roof:
//   slope <= 30°  -> 1.0
//   30° < slope <= 70° -> linear 1.0 -> 0.0 over [30,70]
//   slope > 70°  -> 0.0
func SnowSlopeFactor(roofSlopeDeciDeg int64) float64 {
	slopeDeg := float64(roofSlopeDeciDeg) / 10.0
	switch {
	case slopeDeg <= 30.0:
		return 1.0
	case slopeDeg >= 70.0:
		return 0.0
	default:
		// linear from 1.0 at 30° to 0.0 at 70°
		return 1.0 - (slopeDeg-30.0)/40.0
	}
}

// LiveReductionRatio returns the ASCE 7 live-load reduction multiplier, clamped to [0.5, 1.0].
// formula: ratio = 0.25 + 15/sqrt(KLL * At_ft2), At_ft2 = At_m2 * 10.7639.
// Returns 1.0 when reduction does not apply (At_ft2 < 100, i.e. At_m2 < 9.29).
func LiveReductionRatio(kll int64, tributaryAreaCM2 int64) float64 {
	atM2 := float64(tributaryAreaCM2) / 10000.0
	atFt2 := atM2 * 10.7639
	if atFt2 < 100.0 {
		return 1.0 // no reduction below 100 ft²
	}
	k := float64(kll)
	ratio := 0.25 + 15.0/sqrt(k*atFt2)
	if ratio < 0.5 {
		return 0.5
	}
	if ratio > 1.0 {
		return 1.0
	}
	return ratio
}

// ReducedLiveLoadCenti applies the reduction to an unreduced Lo (centi-kPa).
func ReducedLiveLoadCenti(kll int64, tributaryAreaCM2 int64, loCenti int64) int64 {
	ratio := LiveReductionRatio(kll, tributaryAreaCM2)
	return RoundHalfUp(float64(loCenti) * ratio)
}

// sqrt is a thin wrapper over math.Sqrt for readability.
func sqrt(x float64) float64 {
	return math.Sqrt(x)
}
