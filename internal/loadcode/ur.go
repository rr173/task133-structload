package loadcode

import (
	"task133-structload/internal/model"
)

// URResult is the utilization ratio with the governing demand ratio.
type URResult struct {
	UR            int64               // centi (×100)
	GoverningKind model.GoverningKind
}

// ComputeUR computes the utilization ratio (centi) for a demand vs capacity.
// beam/slab: max(M/capM, V/capV); column: max(P/capP, M/capM); a zero capacity term is skipped.
func ComputeUR(c model.Component, d DemandEffect, capMoment, capShear, capAxial int64) URResult {
	type ratio struct {
		v    float64
		kind model.GoverningKind
	}
	var ratios []ratio
	addRatio := func(num float64, cap int64, kind model.GoverningKind) {
		if cap > 0 {
			ratios = append(ratios, ratio{absFloat(num) / float64(cap), kind})
		}
	}
	switch c.Type {
	case model.ComponentBeam:
		addRatio(float64(d.Moment), capMoment, model.GovM)
		addRatio(float64(d.Shear), capShear, model.GovV)
	case model.ComponentSlab:
		addRatio(float64(d.Moment), capMoment, model.GovM)
	case model.ComponentColumn:
		addRatio(float64(d.Axial), capAxial, model.GovP)
		addRatio(float64(d.Moment), capMoment, model.GovM)
	}
	if len(ratios) == 0 {
		return URResult{UR: 0, GoverningKind: model.GovNone}
	}
	best := ratios[0]
	for _, r := range ratios[1:] {
		if r.v > best.v {
			best = r
		}
	}
	return URResult{UR: RoundHalfUp(best.v * 100.0), GoverningKind: best.kind}
}

// StatusForUR classifies a utilization ratio (centi) into pass/marginal/fail.
func StatusForUR(urCenti int64) model.CheckStatus {
	switch {
	case urCenti <= 100:
		return model.StatusPass
	case urCenti <= 105:
		return model.StatusMarginal
	default:
		return model.StatusFail
	}
}

func absFloat(x float64) float64 {
	return x
}
