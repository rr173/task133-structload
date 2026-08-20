package selfcheck

import (
	"context"
	"fmt"

	"task133-structload/internal/service"
)

// scenarioFullPipeline exercises the complete closed loop end-to-end.
func scenarioFullPipeline(c *httpClient, svc *service.Service, ctx context.Context) error {
	pid, err := c.makeProject("FP", 400, 95)
	if err != nil {
		return err
	}
	if _, err := c.post(fmt.Sprintf("/api/projects/%s/levels", pid), map[string]any{"name": "1F", "elevation": 10000, "gross_area": 1000000}); err != nil {
		return err
	}
	lvlBody, _ := c.get(fmt.Sprintf("/api/projects/%s/levels", pid))
	levelID := str(field(idx(arr(lvlBody), 0), "id"))

	cmp, err := c.post(fmt.Sprintf("/api/levels/%s/components", levelID), map[string]any{
		"code": "B1", "type": "beam", "span": 6000, "tributary_area": 400000,
		"tributary_width": 2500, "kll": 2, "nominal_moment": 100000,
		"nominal_shear": 50000, "phi_b": 90, "phi_c": 0, "phi_v": 90,
		"wind_surface": "windward", "roof_slope": 0,
	})
	if err != nil {
		return err
	}
	cid := str(field(cmp, "id"))
	// Manual loads: D=150, Lo=240.
	c.post(fmt.Sprintf("/api/components/%s/loadcases", cid), map[string]any{"kind": "D", "load_type": "area_pressure", "magnitude": 150, "direction": "gravity"})
	c.post(fmt.Sprintf("/api/components/%s/loadcases", cid), map[string]any{"kind": "L", "load_type": "area_pressure", "magnitude": 240, "direction": "gravity"})
	// Combination: 1.2D + 1.6L + 0.5S.
	c.post(fmt.Sprintf("/api/projects/%s/combinations", pid), map[string]any{
		"name": "1.2D+1.6L+0.5S", "kind": "strength",
		"coeff_d": 120, "coeff_l": 160, "coeff_s": 50,
		"coeff_lr": 0, "coeff_w": 0, "coeff_r": 0, "coeff_e": 0,
	})
	// Run checks.
	if _, err := c.post(fmt.Sprintf("/api/projects/%s/checks/run", pid), nil); err != nil {
		return err
	}
	checks, err := c.get(fmt.Sprintf("/api/components/%s/checks", cid))
	if err != nil {
		return err
	}
	chs := arr(checks)
	if err := assert(len(chs) == 1, fmt.Sprintf("expected 1 check, got %d", len(chs))); err != nil {
		return err
	}
	ch := objMap(idx(chs, 0))
	ur := num(field(ch, "ur"))
	status := str(field(ch, "status"))
	if err := assert(ur > 0, "UR must be > 0"); err != nil {
		return err
	}
	if err := assert(status == "pass" || status == "marginal" || status == "fail", "status must be classified"); err != nil {
		return err
	}
	// Summary must report counts.
	sum, err := c.get(fmt.Sprintf("/api/projects/%s/summary", pid))
	if err != nil {
		return err
	}
	if err := assert(num(field(sum, "total_checks")) >= 1, "summary total_checks >= 1"); err != nil {
		return err
	}
	return nil
}

// scenarioRestartConsistency: snapshot checks, ReconcileAll, assert identical.
func scenarioRestartConsistency(c *httpClient, svc *service.Service, ctx context.Context) error {
	ps, _ := c.get("/api/projects")
	var pid string
	for _, p := range arr(ps) {
		if str(field(objMap(p), "code")) == "FP" {
			pid = str(field(objMap(p), "id"))
			break
		}
	}
	if pid == "" {
		return assert(false, "FP project missing for restart test")
	}
	before, _ := c.get(fmt.Sprintf("/api/projects/%s/checks", pid))
	beforeChecks := arr(before)
	if err := svc.ReconcileAll(ctx, pid); err != nil {
		return err
	}
	after, _ := c.get(fmt.Sprintf("/api/projects/%s/checks", pid))
	afterChecks := arr(after)
	if err := assert(len(beforeChecks) == len(afterChecks), "check count must match after reconcile"); err != nil {
		return err
	}
	for i := range beforeChecks {
		b := objMap(beforeChecks[i])
		a := objMap(afterChecks[i])
		if err := assert(num(field(b, "ur")) == num(field(a, "ur")) && str(field(b, "status")) == str(field(a, "status")) &&
			num(field(b, "demand_moment")) == num(field(a, "demand_moment")) &&
			num(field(b, "capacity_moment")) == num(field(a, "capacity_moment")),
			"check row must be byte-identical after reconcile"); err != nil {
			return err
		}
	}
	// Audit must report consistent.
	audit, _ := c.get(fmt.Sprintf("/api/projects/%s/audit", pid))
	if err := assert(field(audit, "consistent") == true, "audit must report consistent"); err != nil {
		return err
	}
	return nil
}

// scenarioOverride: adjust Mn and assert UR drops after reconcile.
func scenarioOverride(c *httpClient) error {
	ps, _ := c.get("/api/projects")
	var pid, cid string
	for _, p := range arr(ps) {
		if str(field(objMap(p), "code")) == "FP" {
			pid = str(field(objMap(p), "id"))
		}
	}
	comps, _ := c.get(fmt.Sprintf("/api/projects/%s/components", pid))
	for _, comp := range arr(comps) {
		if str(field(objMap(comp), "code")) == "B1" {
			cid = str(field(objMap(comp), "id"))
		}
	}
	if cid == "" {
		return assert(false, "FP B1 component missing for override test")
	}
	beforeChecks, _ := c.get(fmt.Sprintf("/api/components/%s/checks", cid))
	beforeUR := num(field(objMap(idx(arr(beforeChecks), 0)), "ur"))
	// The governing ratio is shear; override nominal_shear upward so UR drops.
	if _, err := c.post(fmt.Sprintf("/api/components/%s/override", cid), map[string]any{
		"field": "nominal_shear", "new_value": "100000", "reason": "increase shear capacity",
	}); err != nil {
		return err
	}
	afterChecks, _ := c.get(fmt.Sprintf("/api/components/%s/checks", cid))
	afterUR := num(field(objMap(idx(arr(afterChecks), 0)), "ur"))
	return assert(afterUR < beforeUR, fmt.Sprintf("UR should drop after shear capacity increase: %d -> %d", beforeUR, afterUR))
}
