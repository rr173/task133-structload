package loadcode

import (
	"task133-structload/internal/model"
)

// DemandEffect accumulates the three demand forces on a component for one combination.
// All centi-units: moment centi-kN·m, shear centi-kN, axial centi-kN.
type DemandEffect struct {
	Moment int64 // centi-kN·m
	Shear  int64 // centi-kN
	Axial  int64 // centi-kN
}

// EffectiveMagnitude is a resolved load magnitude (centi-kPa / centi-kN·m / centi-kN)
// together with its type and direction, ready to be turned into a demand effect.
type EffectiveMagnitude struct {
	Kind      model.LoadKind
	LoadType  model.LoadType
	Magnitude int64  // centi per LoadType
	Direction model.Direction
	CoefCenti int64  // combination coefficient (×100) for this kind
}

// AddComponentDemand folds one effective magnitude into the running demand effect for a component.
// Geometry in mm/cm² is converted internally; results rounded to centi-units via RoundHalfUp.
func (d *DemandEffect) AddComponentDemand(c model.Component, em EffectiveMagnitude) {
	if em.Magnitude == 0 || em.CoefCenti == 0 {
		return
	}
	coef := float64(em.CoefCenti) / 100.0
	switch em.LoadType {
	case model.LoadAreaPressure:
		qKPa := float64(em.Magnitude) / 100.0 * coef
		Lm := float64(c.Span) / 1000.0
		Bm := float64(c.TributaryWidth) / 1000.0
		Am2 := float64(c.TributaryArea) / 10000.0
		sign := areaSign(em.Direction)
		switch c.Type {
		case model.ComponentBeam:
			w := qKPa * Bm // kN/m
			d.Moment += RoundHalfUp(sign * w * Lm * Lm / 8.0 * 100.0)
			d.Shear += RoundHalfUp(sign * w * Lm / 2.0 * 100.0)
		case model.ComponentSlab:
			M := qKPa * Bm * Lm * Lm / 8.0
			d.Moment += RoundHalfUp(sign * M * 100.0)
		case model.ComponentColumn:
			switch em.Direction {
			case model.DirGravity, model.DirUplift:
				d.Axial += RoundHalfUp(sign * qKPa * Am2 * 100.0)
			case model.DirLateral:
				d.Shear += RoundHalfUp(sign * qKPa * Am2 * 100.0)
			}
		}
	case model.LoadLineLoad:
		wKNm := float64(em.Magnitude) / 100.0 * coef
		Lm := float64(c.Span) / 1000.0
		sign := areaSign(em.Direction)
		switch c.Type {
		case model.ComponentBeam, model.ComponentSlab:
			d.Moment += RoundHalfUp(sign * wKNm * Lm * Lm / 8.0 * 100.0)
			d.Shear += RoundHalfUp(sign * wKNm * Lm / 2.0 * 100.0)
		case model.ComponentColumn:
			d.Shear += RoundHalfUp(sign * wKNm * 100.0)
		}
	case model.LoadPointLoad:
		P := float64(em.Magnitude) / 100.0 * coef
		sign := areaSign(em.Direction)
		switch em.Direction {
		case model.DirGravity, model.DirUplift:
			d.Axial += RoundHalfUp(sign * P * 100.0)
		case model.DirLateral:
			d.Axial += RoundHalfUp(sign * P * 100.0)
		}
	}
}

// areaSign returns +1 for gravity/lateral (resisting demand) and -1 for uplift.
// For moment/shear the sign simply tracks direction so that gravity and uplift
// counteract when both present.
func areaSign(dir model.Direction) float64 {
	if dir == model.DirUplift {
		return -1
	}
	return 1
}

// Capacities returns the design strengths (phi * nominal) in centi-units for a component.
// capMoment = round(phiB/100 * Mn_centi/100 * 100) = round(phiB*Mn_centi/100).
func Capacities(c model.Component) (capMoment, capShear, capAxial int64) {
	capMoment = RoundHalfUp(float64(c.PhiB) * float64(c.NominalMoment) / 100.0)
	capShear = RoundHalfUp(float64(c.PhiV) * float64(c.NominalShear) / 100.0)
	capAxial = RoundHalfUp(float64(c.PhiC) * float64(c.NominalAxial) / 100.0)
	return
}
