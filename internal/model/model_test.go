package model

import "testing"

func TestExposureValid(t *testing.T) {
	for _, e := range []ExposureCategory{ExposureB, ExposureC, ExposureD} {
		if !e.Valid() {
			t.Errorf("%s should be valid", e)
		}
	}
	if ExposureCategory("Z").Valid() {
		t.Error("Z should be invalid")
	}
}

func TestImportanceFactorIs(t *testing.T) {
	cases := map[ImportanceCategory]float64{
		ImportanceI: 1.0, ImportanceII: 1.0, ImportanceIII: 1.1, ImportanceIV: 1.2,
	}
	for cat, want := range cases {
		if got := ImportanceFactorIs(cat); got != want {
			t.Errorf("Is(%s)=%v want %v", cat, got, want)
		}
	}
}

func TestWindSurfaceCpExternal(t *testing.T) {
	if WindWindward.CpExternal() != 0.8 {
		t.Error("windward Cp should be 0.8")
	}
	if WindLeeward.CpExternal() != -0.5 {
		t.Error("leeward Cp should be -0.5")
	}
	if WindNetMWFRS.CpExternal() != 1.3 {
		t.Error("net_mwfrs Cp should be 1.3")
	}
}

func TestWindSurfaceDirection(t *testing.T) {
	if WindRoofUplift.DirectionForWind() != DirUplift {
		t.Error("roof_uplift should map to uplift")
	}
	if WindWindward.DirectionForWind() != DirLateral {
		t.Error("windward should map to lateral")
	}
}

func TestCombinationHasAnyNonZero(t *testing.T) {
	c := LoadCombination{CoeffD: 0, CoeffL: 0}
	if c.HasAnyNonZero() {
		t.Error("all-zero combination should be invalid")
	}
	c.CoeffL = 160
	if !c.HasAnyNonZero() {
		t.Error("combination with L should be valid")
	}
}

func TestCombinationCoefficient(t *testing.T) {
	c := LoadCombination{CoeffD: 120, CoeffL: 160, CoeffS: 50}
	if c.Coefficient(LoadDead) != 120 {
		t.Error("D coefficient")
	}
	if c.Coefficient(LoadSnow) != 50 {
		t.Error("S coefficient")
	}
	if c.Coefficient(LoadWind) != 0 {
		t.Error("W should be 0")
	}
}

func TestHTTPCodeMapping(t *testing.T) {
	if HTTPCode(ErrNotFound) != 404 {
		t.Error("not found -> 404")
	}
	if HTTPCode(ErrInvalid) != 422 {
		t.Error("invalid -> 422")
	}
	if HTTPCode(ErrDuplicate) != 409 {
		t.Error("duplicate -> 409")
	}
	if HTTPCode(nil) != 200 {
		t.Error("nil -> 200")
	}
}
