package selfcheck

import "fmt"

// scenarioVelocityPressure: V=40 m/s, exposure C, height 10 m, Kzt=1, Kd=0.85.
// Computes qz and asserts the centi-kPa value via derived wind load on a windward
// component whose load magnitude encodes the wind pressure.
func scenarioVelocityPressure(c *httpClient) error {
	pid, err := c.makeProject("VP", 400, 0) // no snow, so only wind derived
	if err != nil {
		return err
	}
	// Add a level at 10 m so qz uses z=10 m.
	if _, err := c.post(fmt.Sprintf("/api/projects/%s/levels", pid), map[string]any{
		"name": "10m", "elevation": 10000, "gross_area": 1000000,
	}); err != nil {
		return err
	}
	// Create a windward beam on that level.
	lvlBody, err := c.get(fmt.Sprintf("/api/projects/%s/levels", pid))
	if err != nil {
		return err
	}
	levels := arr(lvlBody)
	if len(levels) == 0 {
		return assert(false, "expected one level")
	}
	levelID := str(field(idx(levels, 0), "id"))
	cmpBody, err := c.post(fmt.Sprintf("/api/levels/%s/components", levelID), map[string]any{
		"code": "WB1", "type": "beam", "span": 6000, "tributary_area": 400000,
		"tributary_width": 2500, "kll": 2, "nominal_moment": 100000,
		"nominal_axial": 0, "nominal_shear": 50000, "phi_b": 90, "phi_c": 0, "phi_v": 90,
		"wind_surface": "windward", "roof_slope": 0,
	})
	if err != nil {
		return err
	}
	componentID := str(field(cmpBody, "id"))
	// Derive wind.
	if _, err := c.post(fmt.Sprintf("/api/projects/%s/derive", pid), nil); err != nil {
		return err
	}
	// Fetch derived cases; find the W case and check its magnitude is the wind pressure.
	cases, err := c.get(fmt.Sprintf("/api/components/%s/loadcases", componentID))
	if err != nil {
		return err
	}
	var windMag int64
	for _, lc := range arr(cases) {
		m := objMap(lc)
		if str(field(m, "kind")) == "W" {
			windMag = num(field(m, "magnitude"))
		}
	}
	expected := expectedWindPressureCenti(10000, 100, 85, 400, "C", "windward", 85)
	return assert(windMag == expected, fmt.Sprintf("wind pressure %d != expected %d", windMag, expected))
}

func scenarioWindPressure(c *httpClient) error {
	// Windward (+0.8) vs leeward (-0.5) on the same project must differ in sign.
	pid, err := c.makeProject("WP", 400, 0)
	if err != nil {
		return err
	}
	if _, err := c.post(fmt.Sprintf("/api/projects/%s/levels", pid), map[string]any{"name": "L", "elevation": 10000, "gross_area": 1000000}); err != nil {
		return err
	}
	lvlBody, _ := c.get(fmt.Sprintf("/api/projects/%s/levels", pid))
	levelID := str(field(idx(arr(lvlBody), 0), "id"))

	ww, _ := c.post(fmt.Sprintf("/api/levels/%s/components", levelID), map[string]any{"code": "WW", "type": "beam", "span": 6000, "tributary_area": 400000, "tributary_width": 2500, "kll": 2, "nominal_moment": 100000, "nominal_shear": 50000, "phi_b": 90, "phi_c": 0, "phi_v": 90, "wind_surface": "windward", "roof_slope": 0})
	lw, _ := c.post(fmt.Sprintf("/api/levels/%s/components", levelID), map[string]any{"code": "LW", "type": "beam", "span": 6000, "tributary_area": 400000, "tributary_width": 2500, "kll": 2, "nominal_moment": 100000, "nominal_shear": 50000, "phi_b": 90, "phi_c": 0, "phi_v": 90, "wind_surface": "leeward", "roof_slope": 0})
	c.post(fmt.Sprintf("/api/projects/%s/derive", pid), nil)

	wwCases, _ := c.get(fmt.Sprintf("/api/components/%s/loadcases", str(field(ww, "id"))))
	lwCases, _ := c.get(fmt.Sprintf("/api/components/%s/loadcases", str(field(lw, "id"))))
	wwMag, lwMag := windMagnitude(wwCases), windMagnitude(lwCases)
	if err := assert(wwMag > 0, "windward wind must be > 0"); err != nil {
		return err
	}
	if err := assert(lwMag != 0 && sign(wwMag) != sign(lwMag), fmt.Sprintf("windward %d and leeward %d must have opposite signs", wwMag, lwMag)); err != nil {
		return err
	}
	return nil
}

func windMagnitude(cases any) int64 {
	for _, lc := range arr(cases) {
		m := objMap(lc)
		if str(field(m, "kind")) == "W" {
			return num(field(m, "magnitude"))
		}
	}
	return 0
}

func sign(x int64) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}
