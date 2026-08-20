package loadcode

import (
	"testing"

	"task133-structload/internal/model"
)

func TestBug13_RoofUpliftKeepsDirectionAndUtilization(t *testing.T) {
	direction := model.WindRoofUplift.DirectionForWind()
	if direction != model.DirUplift { t.Fatalf("roof uplift direction = %q", direction) }
	component := model.Component{Type: model.ComponentBeam, Span: 6000, TributaryWidth: 2000, TributaryArea: 120000, NominalMoment: 10000, NominalShear: 10000, PhiB: 100, PhiV: 100}
	var demand DemandEffect
	demand.AddComponentDemand(component, EffectiveMagnitude{Kind: model.LoadWind, LoadType: model.LoadAreaPressure, Magnitude: 100, Direction: direction, CoefCenti: 100})
	if demand.Moment >= 0 { t.Fatalf("uplift moment = %d, want negative", demand.Moment) }
	capM, capV, capP := Capacities(component)
	ur := ComputeUR(component, demand, capM, capV, capP)
	if ur.UR == 0 || ur.GoverningKind != model.GovM { t.Fatalf("uplift utilization = %#v", ur) }
}
