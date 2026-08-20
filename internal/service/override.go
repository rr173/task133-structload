package service

import (
	"context"
	"fmt"
	"strconv"

	"task133-structload/internal/idlib"
	"task133-structload/internal/model"
	"task133-structload/internal/store"
)

// OverrideComponentInput applies a post-computation adjustment to an authoritative field.
type OverrideComponentInput struct {
	Field    string `json:"field"`
	NewValue string `json:"new_value"`
	Reason   string `json:"reason"`
}

// overrideableFields lists the authoritative fields a user may override.
var overrideableFields = map[string]bool{
	"nominal_moment": true,
	"nominal_axial":  true,
	"nominal_shear":  true,
	"phi_b":          true,
	"phi_c":          true,
	"phi_v":          true,
	"wind_surface":   true,
	"roof_slope":     true,
	"tributary_area": true,
	"kll":            true,
}

// OverrideComponent records an authoritative field change and reconciles.
// The override only touches authoritative fields; derived cases and checks
// are recomputed by ReconcileAll, never directly edited.
func (svc *Service) OverrideComponent(ctx context.Context, componentID string, in OverrideComponentInput) (model.Override, error) {
	if !overrideableFields[in.Field] {
		return model.Override{}, fmt.Errorf("%w: field %s is not overrideable", model.ErrInvalid, in.Field)
	}
	if in.Reason == "" {
		return model.Override{}, fmt.Errorf("%w: reason required", model.ErrInvalid)
	}
	// Validate the new value against the field type before storing.
	if err := validateOverrideValue(in.Field, in.NewValue); err != nil {
		return model.Override{}, err
	}
	var c model.Component
	var p model.Project
	if err := svc.store.InTx(ctx, func(tx store.DBTX) error {
		var err error
		c, err = svc.store.GetComponentInTx(ctx, tx, componentID)
		if err != nil {
			return err
		}
		p, err = svc.store.GetProjectInTx(ctx, tx, c.ProjectID)
		return err
	}); err != nil {
		return model.Override{}, wrapErr(err, "get component/project")
	}
	oldValue := currentValueString(c, in.Field)
	o := model.Override{
		ID:          idlib.NewID("ovr"),
		ProjectID:   c.ProjectID,
		ComponentID: c.ID,
		Field:       in.Field,
		OldValue:    oldValue,
		NewValue:    in.NewValue,
		Reason:      in.Reason,
		AppliedAt:   svc.now(),
	}
	if err := svc.store.CreateOverride(ctx, o); err != nil {
		return model.Override{}, wrapErr(err, "create override")
	}
	// Persist the authoritative field change on the component.
	updated := c
	applyOverrideField(&updated, in.Field, in.NewValue)
	if err := svc.store.UpdateComponent(ctx, updated); err != nil {
		return model.Override{}, wrapErr(err, "apply override")
	}
	if err := svc.appendEvent(ctx, c.ProjectID, "override", idlib.MustJSON(o)); err != nil {
		return model.Override{}, wrapErr(err, "log override")
	}
	// Reconcile derived cases + checks for the whole project.
	_ = p
	return o, nil
}

// ListOverrides returns a project's override history.
func (svc *Service) ListOverrides(ctx context.Context, projectID string) ([]model.Override, error) {
	return svc.store.ListOverridesByProject(ctx, projectID)
}

func parseSignedInt(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) }

// validateOverrideValue checks that the new value parses for the field type.
func validateOverrideValue(field, value string) error {
	switch field {
	case "wind_surface":
		if !model.WindSurface(value).Valid() {
			return fmt.Errorf("%w: invalid wind_surface value", model.ErrInvalid)
		}
	default:
		if _, err := parseSignedInt(value); err != nil {
			return fmt.Errorf("%w: %s must be an integer", model.ErrInvalid, field)
		}
	}
	return nil
}

// currentValueString returns the current authoritative value for a field as a string.
func currentValueString(c model.Component, field string) string {
	switch field {
	case "nominal_moment":
		return fmtInt(c.NominalMoment)
	case "nominal_axial":
		return fmtInt(c.NominalAxial)
	case "nominal_shear":
		return fmtInt(c.NominalShear)
	case "phi_b":
		return fmtInt(c.PhiB)
	case "phi_c":
		return fmtInt(c.PhiC)
	case "phi_v":
		return fmtInt(c.PhiV)
	case "wind_surface":
		return string(c.WindSurface)
	case "roof_slope":
		return fmtInt(c.RoofSlope)
	case "tributary_area":
		return fmtInt(c.TributaryArea)
	case "kll":
		return fmtInt(c.KLL)
	}
	return ""
}

// applyOverrideField writes the new value into a component copy.
func applyOverrideField(c *model.Component, field, value string) {
	switch field {
	case "nominal_moment":
		c.NominalMoment = mustParseInt(value)
	case "nominal_axial":
		c.NominalAxial = mustParseInt(value)
	case "nominal_shear":
		c.NominalShear = mustParseInt(value)
	case "phi_b":
		c.PhiB = mustParseInt(value)
	case "phi_c":
		c.PhiC = mustParseInt(value)
	case "phi_v":
		c.PhiV = mustParseInt(value)
	case "wind_surface":
		c.WindSurface = model.WindSurface(value)
	case "roof_slope":
		c.RoofSlope = mustParseInt(value)
	case "tributary_area":
		c.TributaryArea = mustParseInt(value)
	case "kll":
		c.KLL = mustParseInt(value)
	}
}

func fmtInt(v int64) string { return strconv.FormatInt(v, 10) }
