package loadcode

import (
	"math"
	"testing"

	"task133-structload/internal/model"
)

func TestRoundHalfUp(t *testing.T) {
	cases := []struct{ in float64; want int64 }{
		{0, 0}, {0.4, 0}, {0.5, 1}, {0.6, 1}, {1.5, 2}, {2.5, 3}, {-0.5, -1}, {-1.5, -2}, {66.5, 67},
	}
	for _, c := range cases {
		if got := RoundHalfUp(c.in); got != c.want {
			t.Errorf("RoundHalfUp(%v)=%d want %d", c.in, got, c.want)
		}
	}
}

func TestKzExposure(t *testing.T) {
	// Exposure C, z = 10 m: Kz = 2.01 * (10/274.3)^(2/9.5)
	want := 2.01 * math.Pow(10.0/274.3, 2.0/9.5)
	if got := Kz("C", 10.0); math.Abs(got-want) > 1e-9 {
		t.Errorf("Kz C@10m = %v want %v", got, want)
	}
	// Below zmin (C zmin=4.6m) clamps to zmin.
	low := Kz("C", 1.0)
	atMin := Kz("C", 4.6)
	if math.Abs(low-atMin) > 1e-9 {
		t.Errorf("Kz below zmin should clamp: %v vs %v", low, atMin)
	}
	// Unknown exposure returns 1.0.
	if Kz("Z", 10) != 1.0 {
		t.Error("unknown exposure should be 1.0")
	}
}

func TestVelocityPressureCenti(t *testing.T) {
	// V=40 m/s, exposure C, z=10m, Kzt=1, Kd=0.85.
	qz := VelocityPressureCenti("C", 10000, 100, 85, 400)
	// qz_Pa = 0.613 * Kz(10m,C) * 1 * 0.85 * 40^2 ; centi-kPa = round(Pa/10)
	Kz := 2.01 * math.Pow(10.0/274.3, 2.0/9.5)
	wantPa := 0.613 * Kz * 1 * 0.85 * 1600
	want := RoundHalfUp(wantPa / 10.0)
	if qz != want {
		t.Errorf("qz_centi = %d want %d", qz, want)
	}
	if qz <= 0 {
		t.Error("qz should be positive")
	}
}

func TestWindPressureSign(t *testing.T) {
	qz := VelocityPressureCenti("C", 10000, 100, 85, 400)
	ww := WindPressureCenti(qz, 0.8, 85)   // windward, net positive
	lw := WindPressureCenti(qz, -0.5, 85)  // leeward, net negative (suction)
	if ww <= 0 {
		t.Errorf("windward wind pressure should be > 0, got %d", ww)
	}
	if lw >= 0 {
		t.Errorf("leeward wind pressure should be < 0 (suction), got %d", lw)
	}
}

func TestLiveReductionRatio(t *testing.T) {
	// 40 m², KLL=2 -> At_ft² = 40*10.7639 = 430.56; ratio = 0.25+15/sqrt(2*430.56)
	want := 0.25 + 15.0/math.Sqrt(2.0*40.0*10.7639)
	if got := LiveReductionRatio(2, 400000); math.Abs(got-want) > 1e-9 {
		t.Errorf("reduction ratio = %v want %v", got, want)
	}
	// Below threshold (5 m² < 9.29) -> 1.0 (no reduction).
	if got := LiveReductionRatio(2, 50000); got != 1.0 {
		t.Errorf("small area should not reduce: %v", got)
	}
	// Clamped to >= 0.5 for huge areas.
	r := LiveReductionRatio(1, 100000000)
	if r < 0.5 {
		t.Errorf("ratio clamped to >= 0.5: %v", r)
	}
}

func TestReducedLiveLoadCentiValue(t *testing.T) {
	// Lo=240 (2.4 kPa), 40 m², KLL=2 -> reduced ~183.
	got := ReducedLiveLoadCenti(2, 400000, 240)
	want := RoundHalfUp(240.0 * (0.25 + 15.0/math.Sqrt(2.0*40.0*10.7639)))
	if got != want {
		t.Errorf("reduced L = %d want %d", got, want)
	}
	if got <= 0 || got >= 240 {
		t.Errorf("reduced L should be in (0, Lo=240): %d", got)
	}
}

func TestSnowSlopeFactor(t *testing.T) {
	if SnowSlopeFactor(0) != 1.0 {
		t.Error("flat roof Cs=1.0")
	}
	if SnowSlopeFactor(300) != 1.0 { // 30°
		t.Error("30° Cs=1.0")
	}
	if SnowSlopeFactor(700) != 0.0 { // 70°
		t.Error("70° Cs=0")
	}
	// 45° -> 1 - (45-30)/40 = 0.625
	if got := SnowSlopeFactor(450); math.Abs(got-0.625) > 1e-9 {
		t.Errorf("45° Cs = %v want 0.625", got)
	}
}

func TestSnowLoadCenti(t *testing.T) {
	// pg=0.95 kPa (95), Ce=1, Ct=1, Is=1, flat roof (slope 0).
	// pf_kPa = 0.7 * 1 * 1 * 1 * 0.95 = 0.665; pf_centi = round_half_up(0.665*100) = round(66.5) = 67.
	pf, ps := SnowLoadCenti(100, 100, 1.0, 95, 0)
	want := RoundHalfUp(0.7 * 1.0 * 1.0 * 1.0 * 0.95 * 100.0)
	if pf != want {
		t.Errorf("pf = %d want %d", pf, want)
	}
	if ps != pf {
		t.Errorf("flat roof ps should equal pf: %d vs %d", ps, pf)
	}
}

func TestComputeURBeam(t *testing.T) {
	c := testComponent()
	// demand moment = 5690 centi-kN·m, cap = round(90*10000/100)=9000.
	d := DemandEffect{Moment: 5690, Shear: 3794}
	capM, capV, _ := Capacities(c)
	if capM != 9000 {
		t.Errorf("capM = %d want 9000", capM)
	}
	// UR = max(5690/9000, 3794/4500) = max(0.632, 0.843) = 0.843 -> 84
	r := ComputeUR(c, d, capM, capV, 0)
	if r.UR <= 0 {
		t.Errorf("UR should be positive, got %d", r.UR)
	}
	if r.GoverningKind == "" {
		t.Error("governing kind should be set")
	}
}

func TestStatusForUR(t *testing.T) {
	if StatusForUR(80) != "pass" {
		t.Error("80 -> pass")
	}
	if StatusForUR(100) != "pass" {
		t.Error("100 -> pass")
	}
	if StatusForUR(103) != "marginal" {
		t.Error("103 -> marginal")
	}
	if StatusForUR(120) != "fail" {
		t.Error("120 -> fail")
	}
}

// helper: a beam with Mn=10000, Vn=5000, phi 90.
func testComponent() model.Component {
	return model.Component{
		Type: model.ComponentBeam, Span: 6000, TributaryWidth: 2500, TributaryArea: 400000,
		KLL: 2, NominalMoment: 10000, NominalShear: 5000, PhiB: 90, PhiV: 90, PhiC: 90,
	}
}
