package store

import (
	"context"
	"fmt"

	"task133-structload/internal/model"
)

func (s *Store) loadProjects(ctx context.Context, snap *Snapshot) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id,code,name,site,wind_zone,snow_zone,exposure,importance,
		wind_speed,kzt,kd,gust_g,ground_snow,snow_exposure_factor,thermal_factor,units,created_at
		FROM projects ORDER BY created_at`)
	if err != nil {
		return fmt.Errorf("load projects: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var p model.Project
		var ms int64
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Site, &p.WindZone, &p.SnowZone,
			&p.Exposure, &p.Importance, &p.WindSpeed, &p.Kzt, &p.Kd, &p.GustG, &p.GroundSnow,
			&p.SnowExposureFactor, &p.ThermalFactor, &p.Units, &ms); err != nil {
			return err
		}
		p.CreatedAt = fromMS(ms)
		snap.Projects[p.ID] = p
	}
	return rows.Err()
}

func (s *Store) loadLevels(ctx context.Context, snap *Snapshot) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id,project_id,name,elevation,gross_area,created_at FROM levels ORDER BY created_at`)
	if err != nil {
		return fmt.Errorf("load levels: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var l model.Level
		var ms int64
		if err := rows.Scan(&l.ID, &l.ProjectID, &l.Name, &l.Elevation, &l.GrossArea, &ms); err != nil {
			return err
		}
		l.CreatedAt = fromMS(ms)
		snap.Levels[l.ID] = l
	}
	return rows.Err()
}

func (s *Store) loadComponents(ctx context.Context, snap *Snapshot) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id,project_id,level_id,code,type,span,tributary_area,tributary_width,
		kll,nominal_moment,nominal_axial,nominal_shear,phi_b,phi_c,phi_v,wind_surface,roof_slope,section_label,created_at
		FROM components ORDER BY created_at`)
	if err != nil {
		return fmt.Errorf("load components: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var c model.Component
		var ms int64
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.LevelID, &c.Code, &c.Type, &c.Span, &c.TributaryArea,
			&c.TributaryWidth, &c.KLL, &c.NominalMoment, &c.NominalAxial, &c.NominalShear,
			&c.PhiB, &c.PhiC, &c.PhiV, &c.WindSurface, &c.RoofSlope, &c.SectionLabel, &ms); err != nil {
			return err
		}
		c.CreatedAt = fromMS(ms)
		snap.Components[c.ID] = c
	}
	return rows.Err()
}

func (s *Store) loadLoadCases(ctx context.Context, snap *Snapshot) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id,project_id,component_id,kind,load_type,magnitude,direction,origin,source_event_id,note,created_at
		FROM load_cases ORDER BY created_at`)
	if err != nil {
		return fmt.Errorf("load load_cases: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var lc model.LoadCase
		var ms int64
		if err := rows.Scan(&lc.ID, &lc.ProjectID, &lc.ComponentID, &lc.Kind, &lc.LoadType,
			&lc.Magnitude, &lc.Direction, &lc.Origin, &lc.SourceEventID, &lc.Note, &ms); err != nil {
			return err
		}
		lc.CreatedAt = fromMS(ms)
		if lc.Origin == model.OriginManual {
			snap.ManualCases = append(snap.ManualCases, lc)
		} else {
			snap.DerivedCases = append(snap.DerivedCases, lc)
		}
	}
	return rows.Err()
}

func (s *Store) loadCombos(ctx context.Context, snap *Snapshot) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id,project_id,name,kind,coeff_d,coeff_l,coeff_lr,coeff_s,coeff_w,coeff_r,coeff_e,created_at
		FROM load_combinations ORDER BY created_at`)
	if err != nil {
		return fmt.Errorf("load load_combinations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var c model.LoadCombination
		var ms int64
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Name, &c.Kind, &c.CoeffD, &c.CoeffL, &c.CoeffLr,
			&c.CoeffS, &c.CoeffW, &c.CoeffR, &c.CoeffE, &ms); err != nil {
			return err
		}
		c.CreatedAt = fromMS(ms)
		snap.Combos[c.ID] = c
	}
	return rows.Err()
}

func (s *Store) loadOverrides(ctx context.Context, snap *Snapshot) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id,project_id,component_id,field,old_value,new_value,reason,applied_at FROM overrides ORDER BY applied_at`)
	if err != nil {
		return fmt.Errorf("load overrides: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var o model.Override
		var ms int64
		if err := rows.Scan(&o.ID, &o.ProjectID, &o.ComponentID, &o.Field, &o.OldValue, &o.NewValue, &o.Reason, &ms); err != nil {
			return err
		}
		o.AppliedAt = fromMS(ms)
		snap.Overrides = append(snap.Overrides, o)
	}
	return rows.Err()
}

func (s *Store) loadEvents(ctx context.Context, snap *Snapshot) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id,project_id,seq,event_type,payload,created_at FROM event_log ORDER BY seq`)
	if err != nil {
		return fmt.Errorf("load event_log: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var e model.EventLog
		var ms int64
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Seq, &e.EventType, &e.Payload, &ms); err != nil {
			return err
		}
		e.CreatedAt = fromMS(ms)
		snap.Events = append(snap.Events, e)
	}
	return rows.Err()
}
