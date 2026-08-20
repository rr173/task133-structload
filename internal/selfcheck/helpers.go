package selfcheck

import "math"

// expectedWindPressureCenti independently recomputes the wind pressure (centi-kPa)
// for a windward/leeward/etc. surface so the smoke-test cross-checks loadcode.
// Parameters: heightMM, kztCenti, kdCenti, windSpeedDeci, exposure, surface, gustGCenti.
func expectedWindPressureCenti(heightMM, kztCenti, kdCenti, windSpeedDeci int64, exposure, surface string, gustGCenti int64) int64 {
	V := float64(windSpeedDeci) / 10.0
	Kz := expectedKz(exposure, float64(heightMM)/1000.0)
	Kzt := float64(kztCenti) / 100.0
	Kd := float64(kdCenti) / 100.0
	qzPa := 0.613 * Kz * Kzt * Kd * V * V
	qzCenti := roundHalfUp(qzPa / 10.0) // centi-kPa (0.01 kPa = 10 Pa)
	cpExt := cpExternal(surface)
	G := float64(gustGCenti) / 100.0
	pKPa := float64(qzCenti)/100.0 * (cpExt + 0.18) * G
	return roundHalfUp(pKPa * 100.0)
}

type expParam struct{ alpha, zg, zmin float64 }

var expParams = map[string]expParam{
	"B": {7.0, 365.4, 9.1},
	"C": {9.5, 274.3, 4.6},
	"D": {11.5, 213.4, 2.1},
}

func expectedKz(exposure string, zMeters float64) float64 {
	p, ok := expParams[exposure]
	if !ok {
		return 1.0
	}
	if zMeters < p.zmin {
		zMeters = p.zmin
	}
	return 2.01 * math.Pow(zMeters/p.zg, 2.0/p.alpha)
}

func cpExternal(surface string) float64 {
	switch surface {
	case "windward":
		return 0.8
	case "leeward":
		return -0.5
	case "side":
		return -0.7
	case "roof_uplift":
		return -1.0
	case "net_mwfrs":
		return 1.3
	}
	return 0
}

func roundHalfUp(f float64) int64 {
	// 1e-9 epsilon compensates float64 representation drift (e.g. 0.665*100).
	if f >= 0 {
		return int64(math.Floor(f + 0.5 + 1e-9))
	}
	return int64(math.Ceil(f - 0.5 - 1e-9))
}

// expectedReducedLiveLoad independently computes the reduced live load (centi-kPa).
func expectedReducedLiveLoad(kll int64, tributaryAreaCM2 int64, loCenti int64) int64 {
	atM2 := float64(tributaryAreaCM2) / 10000.0
	atFt2 := atM2 * 10.7639
	if atFt2 < 100.0 {
		return loCenti
	}
	ratio := 0.25 + 15.0/math.Sqrt(float64(kll)*atFt2)
	if ratio < 0.5 {
		ratio = 0.5
	}
	if ratio > 1.0 {
		ratio = 1.0
	}
	return roundHalfUp(float64(loCenti) * ratio)
}

// expectedSnowLoad independently computes the sloped roof snow load (centi-kPa).
func expectedSnowLoad(ceCenti, ctCenti int64, importanceFactor float64, pgCenti int64, roofSlopeDeciDeg int64) int64 {
	Ce := float64(ceCenti) / 100.0
	Ct := float64(ctCenti) / 100.0
	pg := float64(pgCenti) / 100.0
	pfKPa := 0.7 * Ce * Ct * importanceFactor * pg
	pfCenti := roundHalfUp(pfKPa * 100.0)
	Cs := expectedSnowSlope(roofSlopeDeciDeg)
	return roundHalfUp(float64(pfCenti) * Cs)
}

func expectedSnowSlope(roofSlopeDeciDeg int64) float64 {
	slopeDeg := float64(roofSlopeDeciDeg) / 10.0
	switch {
	case slopeDeg <= 30.0:
		return 1.0
	case slopeDeg >= 70.0:
		return 0.0
	default:
		return 1.0 - (slopeDeg-30.0)/40.0
	}
}
