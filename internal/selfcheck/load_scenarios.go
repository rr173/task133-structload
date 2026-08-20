package selfcheck

import "fmt"

// scenarioLiveReduction: beam, tributary 40 m², KLL=2, Lo=240 (2.4 kPa).
// Expected reduced L = 183 centi-kPa (see design doc §10.3).
func scenarioLiveReduction(c *httpClient) error {
	pid, err := c.makeProject("LR", 400, 0)
	if err != nil {
		return err
	}
	if _, err := c.post(fmt.Sprintf("/api/projects/%s/levels", pid), map[string]any{"name": "L", "elevation": 10000, "gross_area": 1000000}); err != nil {
		return err
	}
	lvlBody, _ := c.get(fmt.Sprintf("/api/projects/%s/levels", pid))
	levelID := str(field(idx(arr(lvlBody), 0), "id"))
	cmp, err := c.post(fmt.Sprintf("/api/levels/%s/components", levelID), map[string]any{
		"code": "LB", "type": "beam", "span": 6000, "tributary_area": 400000,
		"tributary_width": 2500, "kll": 2, "nominal_moment": 100000,
		"nominal_shear": 50000, "phi_b": 90, "phi_c": 0, "phi_v": 90,
		"wind_surface": "none", "roof_slope": 0,
	})
	if err != nil {
		return err
	}
	cid := str(field(cmp, "id"))
	// Manual live load Lo=240.
	if _, err := c.post(fmt.Sprintf("/api/components/%s/loadcases", cid), map[string]any{
		"kind": "L", "load_type": "area_pressure", "magnitude": 240, "direction": "gravity",
	}); err != nil {
		return err
	}
	c.post(fmt.Sprintf("/api/projects/%s/derive", pid), nil)
	cases, _ := c.get(fmt.Sprintf("/api/components/%s/loadcases", cid))
	// Two L cases: manual (240) and derived (reduced). The reduced must equal expected.
	var derivedL int64 = -1
	var manualL int64
	for _, lc := range arr(cases) {
		m := objMap(lc)
		if str(field(m, "kind")) == "L" {
			if str(field(m, "origin")) == "derived" {
				derivedL = num(field(m, "magnitude"))
			} else {
				manualL = num(field(m, "magnitude"))
			}
		}
	}
	if err := assert(manualL == 240, "manual Lo should be 240"); err != nil {
		return err
	}
	expected := expectedReducedLiveLoad(2, 400000, 240)
	return assert(derivedL == expected, fmt.Sprintf("reduced L %d != expected %d", derivedL, expected))
}

// scenarioSmallAreaNoReduction: tributary 5 m² (< 9.29) -> no reduction, L = Lo.
func scenarioSmallAreaNoReduction(c *httpClient) error {
	pid, err := c.makeProject("SA", 400, 0)
	if err != nil {
		return err
	}
	if _, err := c.post(fmt.Sprintf("/api/projects/%s/levels", pid), map[string]any{"name": "L", "elevation": 10000, "gross_area": 1000000}); err != nil {
		return err
	}
	lvlBody, _ := c.get(fmt.Sprintf("/api/projects/%s/levels", pid))
	levelID := str(field(idx(arr(lvlBody), 0), "id"))
	cmp, err := c.post(fmt.Sprintf("/api/levels/%s/components", levelID), map[string]any{
		"code": "SB", "type": "beam", "span": 3000, "tributary_area": 50000, // 5 m²
		"tributary_width": 1000, "kll": 2, "nominal_moment": 100000,
		"nominal_shear": 50000, "phi_b": 90, "phi_c": 0, "phi_v": 90,
		"wind_surface": "none", "roof_slope": 0,
	})
	if err != nil {
		return err
	}
	cid := str(field(cmp, "id"))
	c.post(fmt.Sprintf("/api/components/%s/loadcases", cid), map[string]any{"kind": "L", "load_type": "area_pressure", "magnitude": 240, "direction": "gravity"})
	c.post(fmt.Sprintf("/api/projects/%s/derive", pid), nil)
	cases, _ := c.get(fmt.Sprintf("/api/components/%s/loadcases", cid))
	var derivedL int64 = -1
	for _, lc := range arr(cases) {
		m := objMap(lc)
		if str(field(m, "kind")) == "L" && str(field(m, "origin")) == "derived" {
			derivedL = num(field(m, "magnitude"))
		}
	}
	if err := assert(derivedL != -1, "expected a derived L case"); err != nil {
		return err
	}
	return assert(derivedL == 240, fmt.Sprintf("small-area L should equal Lo=240, got %d", derivedL))
}

// scenarioSnowSlope: pg=0.95 kPa (95), Ce=1, Ct=1, Is=1 (II). Slope 45° -> Cs=0.625.
func scenarioSnowSlope(c *httpClient) error {
	pid, err := c.makeProject("SN", 400, 95)
	if err != nil {
		return err
	}
	if _, _, err := c.do("PATCH", fmt.Sprintf("/api/projects/%s", pid), map[string]any{"wind_speed": 400}); err != nil {
		// not fatal; wind unused here
		_ = err
	}
	if _, err := c.post(fmt.Sprintf("/api/projects/%s/levels", pid), map[string]any{"name": "Roof", "elevation": 30000, "gross_area": 1000000}); err != nil {
		return err
	}
	lvlBody, _ := c.get(fmt.Sprintf("/api/projects/%s/levels", pid))
	levelID := str(field(idx(arr(lvlBody), 0), "id"))
	cmp, err := c.post(fmt.Sprintf("/api/levels/%s/components", levelID), map[string]any{
		"code": "RF", "type": "slab", "span": 6000, "tributary_area": 400000,
		"tributary_width": 2500, "kll": 1, "nominal_moment": 100000,
		"nominal_shear": 0, "phi_b": 90, "phi_c": 0, "phi_v": 0,
		"wind_surface": "none", "roof_slope": 450, // 45.0°
	})
	if err != nil {
		return err
	}
	cid := str(field(cmp, "id"))
	c.post(fmt.Sprintf("/api/projects/%s/derive", pid), nil)
	cases, _ := c.get(fmt.Sprintf("/api/components/%s/loadcases", cid))
	var snow int64 = -1
	for _, lc := range arr(cases) {
		m := objMap(lc)
		if str(field(m, "kind")) == "S" {
			snow = num(field(m, "magnitude"))
		}
	}
	if err := assert(snow != -1, "expected a derived S case"); err != nil {
		return err
	}
	expected := expectedSnowLoad(100, 100, 1.0, 95, 450)
	return assert(snow == expected, fmt.Sprintf("snow %d != expected %d", snow, expected))
}
