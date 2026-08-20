package model

import "time"

// Fixed-point unit conventions (all external-facing numeric fields are int64):
//   pressure / area load  -> centi-kPa   (0.01 kPa)      [1 kPa = 100]
//   line load             -> centi-kN/m  (0.01 kN/m)
//   point load            -> centi-kN    (0.01 kN)
//   moment                -> centi-kN·m  (0.01 kN·m)
//   length / elevation    -> mm
//   area                  -> cm²         [1 m² = 10000]
//   wind speed            -> deci-m/s    (0.1 m/s)
//   roof slope            -> deci-deg    (0.1°)
//   dimensionless factors  -> centi      (×100, e.g. 0.85 -> 85)
//   UR (utilization)      -> centi       (×100, e.g. 0.6329 -> 63)

// Project is the top-level building entity carrying wind/snow design parameters.
type Project struct {
	ID                 string           `json:"id"`
	Code               string           `json:"code"`
	Name               string           `json:"name"`
	Site               string           `json:"site"`
	WindZone           string           `json:"wind_zone"`
	SnowZone           string           `json:"snow_zone"`
	Exposure           ExposureCategory `json:"exposure"`
	Importance         ImportanceCategory `json:"importance"`
	WindSpeed          int64            `json:"wind_speed"`           // deci-m/s
	Kzt                int64            `json:"kzt"`                   // centi (×100)
	Kd                 int64            `json:"kd"`                   // centi (×100), default 85
	GustG              int64            `json:"gust_g"`               // centi (×100), default 85
	GroundSnow         int64            `json:"ground_snow"`          // centi-kPa
	SnowExposureFactor int64            `json:"snow_exposure_factor"` // centi (Ce, default 100)
	ThermalFactor      int64            `json:"thermal_factor"`      // centi (Ct, default 100)
	Units              string           `json:"units"`
	CreatedAt          time.Time        `json:"created_at"`
}

// Level is a building floor with elevation and gross area.
type Level struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	Elevation int64     `json:"elevation"`  // mm
	GrossArea int64     `json:"gross_area"` // cm²
	CreatedAt time.Time `json:"created_at"`
}

// Component is a structural member with geometry and nominal strengths.
type Component struct {
	ID             string           `json:"id"`
	ProjectID      string           `json:"project_id"`
	LevelID        string           `json:"level_id"`        // may be empty
	Code           string           `json:"code"`
	Type           ComponentType    `json:"type"`
	Span           int64            `json:"span"`           // mm
	TributaryArea  int64            `json:"tributary_area"`  // cm²
	TributaryWidth int64            `json:"tributary_width"` // mm (for line loads)
	KLL            int64            `json:"kll"`
	NominalMoment  int64            `json:"nominal_moment"`  // centi-kN·m
	NominalAxial   int64            `json:"nominal_axial"`   // centi-kN
	NominalShear   int64            `json:"nominal_shear"`    // centi-kN
	PhiB           int64            `json:"phi_b"`            // centi (×100)
	PhiC           int64            `json:"phi_c"`            // centi (×100)
	PhiV           int64            `json:"phi_v"`            // centi (×100)
	WindSurface    WindSurface      `json:"wind_surface"`
	RoofSlope      int64            `json:"roof_slope"`       // deci-deg
	SectionLabel   string           `json:"section_label"`
	CreatedAt      time.Time        `json:"created_at"`
}

// LoadCase is one load of a given kind on a component.
type LoadCase struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	ComponentID   string    `json:"component_id"`
	Kind          LoadKind  `json:"kind"`
	LoadType      LoadType  `json:"load_type"`
	Magnitude     int64     `json:"magnitude"` // centi-kPa / centi-kN·m / centi-kN per LoadType
	Direction     Direction `json:"direction"`
	Origin        Origin    `json:"origin"`
	SourceEventID string    `json:"source_event_id,omitempty"` // for derived
	Note          string    `json:"note"`
	CreatedAt     time.Time `json:"created_at"`
}

// LoadCombination is an LRFD factor set over load kinds.
type LoadCombination struct {
	ID        string           `json:"id"`
	ProjectID string           `json:"project_id"`
	Name      string           `json:"name"`
	Kind      CombinationKind  `json:"kind"`
	CoeffD    int64            `json:"coeff_d"`  // centi (×100)
	CoeffL    int64            `json:"coeff_l"`
	CoeffLr   int64            `json:"coeff_lr"`
	CoeffS    int64            `json:"coeff_s"`
	CoeffW    int64            `json:"coeff_w"`
	CoeffR    int64            `json:"coeff_r"`
	CoeffE    int64            `json:"coeff_e"`
	CreatedAt time.Time        `json:"created_at"`
}

// HasAnyNonZero reports whether at least one factor is non-zero (a valid combination).
func (c LoadCombination) HasAnyNonZero() bool {
	return c.CoeffD != 0 || c.CoeffL != 0 || c.CoeffLr != 0 ||
		c.CoeffS != 0 || c.CoeffW != 0 || c.CoeffR != 0 || c.CoeffE != 0
}

// Coefficient returns the centi-coefficient for a load kind.
func (c LoadCombination) Coefficient(k LoadKind) int64 {
	switch k {
	case LoadDead:
		return c.CoeffD
	case LoadLive:
		return c.CoeffL
	case LoadRoof:
		return c.CoeffLr
	case LoadSnow:
		return c.CoeffS
	case LoadWind:
		return c.CoeffW
	case LoadRain:
		return c.CoeffR
	case LoadSeismic:
		return c.CoeffE
	}
	return 0
}

// Check is a sufficiency evaluation of one component under one combination.
type Check struct {
	ID             string         `json:"id"`
	ProjectID      string         `json:"project_id"`
	CombinationID  string         `json:"combination_id"`
	ComponentID    string         `json:"component_id"`
	DemandMoment   int64          `json:"demand_moment"`  // centi-kN·m
	DemandShear    int64          `json:"demand_shear"`   // centi-kN
	DemandAxial    int64          `json:"demand_axial"`   // centi-kN
	CapacityMoment int64          `json:"capacity_moment"` // centi-kN·m = phiB·Mn
	CapacityShear  int64          `json:"capacity_shear"`
	CapacityAxial  int64          `json:"capacity_axial"`
	UR             int64          `json:"ur"` // centi (×100)
	Status         CheckStatus    `json:"status"`
	GoverningKind  GoverningKind  `json:"governing_kind"`
	ComputedAt     time.Time      `json:"computed_at"`
}

// Override is a post-computation adjustment to an authoritative field.
type Override struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	ComponentID string    `json:"component_id"`
	Field       string    `json:"field"`
	OldValue    string    `json:"old_value"`
	NewValue    string    `json:"new_value"`
	Reason      string    `json:"reason"`
	AppliedAt   time.Time `json:"applied_at"`
}

// EventLog records every authoritative write for replay recovery.
type EventLog struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Seq       int64     `json:"seq"`
	EventType string    `json:"event_type"`
	Payload   string    `json:"payload"` // JSON
	CreatedAt time.Time `json:"created_at"`
}

// Summary is the compliance roll-up for a project.
type Summary struct {
	ProjectID         string  `json:"project_id"`
	TotalChecks       int     `json:"total_checks"`
	PassCount         int     `json:"pass_count"`
	MarginalCount     int     `json:"marginal_count"`
	FailCount         int     `json:"fail_count"`
	MaxUR             int64   `json:"max_ur"` // centi
	FirstFailComponent string `json:"first_fail_component,omitempty"`
	FirstFailCombo     string `json:"first_fail_combo,omitempty"`
}

// ReplayRequest rebuilds authoritative state from the event log.
type ReplayRequest struct {
	ProjectID string `json:"project_id"`
}

// AuditResult reports whether derived cases and checks are consistent with inputs.
type AuditResult struct {
	ProjectID         string `json:"project_id"`
	DerivedCases      int    `json:"derived_cases"`
	Checks            int    `json:"checks"`
	StaleDerivedCases int    `json:"stale_derived_cases"`
	StaleChecks       int    `json:"stale_checks"`
	Consistent        bool   `json:"consistent"`
}
