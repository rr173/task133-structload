package loadcode

import "math"

// RoundHalfUp rounds f to the nearest integer, halves rounded away from zero.
// A tiny epsilon (1e-9) compensates for float64 representation drift so that
// mathematically-exact half values (e.g. 0.665*100 == 66.4999... in float) still
// round up as intended. Used for every fixed-point storage value so selfcheck
// can assert exact ints.
func RoundHalfUp(f float64) int64 {
	if f >= 0 {
		return int64(math.Floor(f + 0.5 + 1e-9))
	}
	return int64(math.Ceil(f - 0.5 - 1e-9))
}

// Exposure constants per ASCE 7: alpha (power-law exponent), zg (gradient height m), zmin (m).
type exposureParam struct {
	alpha float64
	zg    float64
	zmin  float64
}

var exposureParams = map[string]exposureParam{
	"B": {alpha: 7.0, zg: 365.4, zmin: 9.1},
	"C": {alpha: 9.5, zg: 274.3, zmin: 4.6},
	"D": {alpha: 11.5, zg: 213.4, zmin: 2.1},
}

// Kz computes the velocity-pressure exposure coefficient for a height z (m) and exposure category.
// Kz = 2.01 * (z_eff/zg)^(2/alpha), z_eff = max(z, zmin). Returns 1.0 if exposure unknown.
func Kz(exposure string, zMeters float64) float64 {
	p, ok := exposureParams[exposure]
	if !ok {
		return 1.0
	}
	zEff := zMeters
	if zEff < p.zmin {
		zEff = p.zmin
	}
	return 2.01 * math.Pow(zEff/p.zg, 2.0/p.alpha)
}

// VelocityPressure computes qz = 0.613 * Kz * Kzt * Kd * V^2 in Pascals.
// V is in m/s; Kz/Kzt/Kd are dimensionless factors.
func VelocityPressure(Kz, Kzt, Kd, Vmps float64) float64 {
	return 0.613 * Kz * Kzt * Kd * Vmps * Vmps
}

// VelocityPressureCenti stores qz in centi-kPa (0.01 kPa = 10 Pa) as a fixed int.
// qz_centi = round_half_up(qz_Pa / 10).
func VelocityPressureCenti(exposure string, heightMM int64, kztCenti, kdCenti, windSpeedDeci int64) int64 {
	Vmps := float64(windSpeedDeci) / 10.0
	Kz := Kz(exposure, float64(heightMM)/1000.0)
	Kzt := float64(kztCenti) / 100.0
	Kd := float64(kdCenti) / 100.0
	qzPa := VelocityPressure(Kz, Kzt, Kd, Vmps)
	// 1 centi-kPa = 10 Pa
	return RoundHalfUp(qzPa / 10.0)
}

// InternalPressureCoefficient (ASCE 7 enclosed building, +Gcpi).
const CpInternal = 0.18

// WindPressureCenti computes design wind pressure p (centi-kPa) on a component's wind surface.
// p_kPa = qz_kPa * (Cp_ext + Cp_int) * G.
// With qz in centi-kPa: p_centi = round_half_up(qz_centi * (Cp_ext + Cp_int) * G_centi / 100).
func WindPressureCenti(qzCenti int64, cpExternal float64, gustGCenti int64) int64 {
	qzKPa := float64(qzCenti) / 100.0
	G := float64(gustGCenti) / 100.0
	pKPa := qzKPa * (cpExternal + CpInternal) * G
	return RoundHalfUp(-pKPa * 100.0)
}
