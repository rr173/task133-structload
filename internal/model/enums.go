package model

// ComponentType enumerates structural member kinds.
type ComponentType string

const (
	ComponentBeam   ComponentType = "beam"
	ComponentColumn ComponentType = "column"
	ComponentSlab   ComponentType = "slab"
)

func (c ComponentType) Valid() bool {
	switch c {
	case ComponentBeam, ComponentColumn, ComponentSlab:
		return true
	}
	return false
}

// LoadKind enumerates load case families.
type LoadKind string

const (
	LoadDead    LoadKind = "D" // dead (permanent)
	LoadLive    LoadKind = "L" // live (reducible); manual stores Lo, derived stores reduced L
	LoadRoof    LoadKind = "Lr" // roof live
	LoadSnow    LoadKind = "S" // snow (derived)
	LoadWind    LoadKind = "W" // wind (derived)
	LoadRain    LoadKind = "R" // rain
	LoadSeismic LoadKind = "E" // seismic
)

func (k LoadKind) Valid() bool {
	switch k {
	case LoadDead, LoadLive, LoadRoof, LoadSnow, LoadWind, LoadRain, LoadSeismic:
		return true
	}
	return false
}

// IsDerived reports whether the kind is engine-derived (W/S) or reducible (L).
func (k LoadKind) IsDerivedMagnitude() bool {
	switch k {
	case LoadSnow, LoadWind:
		return true
	}
	return false
}

// LoadType describes how magnitude is interpreted.
type LoadType string

const (
	LoadAreaPressure LoadType = "area_pressure" // magnitude in centi-kPa
	LoadLineLoad     LoadType = "line_load"     // magnitude in centi-kN/m
	LoadPointLoad    LoadType = "point_load"    // magnitude in centi-kN
)

func (t LoadType) Valid() bool {
	switch t {
	case LoadAreaPressure, LoadLineLoad, LoadPointLoad:
		return true
	}
	return false
}

// Direction of a load effect.
type Direction string

const (
	DirGravity Direction = "gravity" // downward
	DirUplift  Direction = "uplift"   // upward (negative)
	DirLateral Direction = "lateral"  // horizontal
)

func (d Direction) Valid() bool {
	switch d {
	case DirGravity, DirUplift, DirLateral:
		return true
	}
	return false
}

// Origin distinguishes authoritative (manual) from computed (derived) load cases.
type Origin string

const (
	OriginManual  Origin = "manual"
	OriginDerived Origin = "derived"
)

// WindSurface labels how a component sees wind pressure.
type WindSurface string

const (
	WindNone       WindSurface = "none"
	WindWindward   WindSurface = "windward"   // Cp_ext +0.8
	WindLeeward    WindSurface = "leeward"    // Cp_ext -0.5
	WindSide       WindSurface = "side"       // Cp_ext -0.7
	WindRoofUplift WindSurface = "roof_uplift" // Cp_ext -1.0
	WindNetMWFRS   WindSurface = "net_mwfrs"  // net Cp +1.3
)

func (w WindSurface) Valid() bool {
	switch w {
	case WindNone, WindWindward, WindLeeward, WindSide, WindRoofUplift, WindNetMWFRS:
		return true
	}
	return false
}

// CpExternal returns the ASCE 7 external pressure coefficient for a wind surface.
func (w WindSurface) CpExternal() float64 {
	switch w {
	case WindWindward:
		return 0.8
	case WindLeeward:
		return -0.5
	case WindSide:
		return -0.7
	case WindRoofUplift:
		return -1.0
	case WindNetMWFRS:
		return 1.3
	default:
		return 0
	}
}

// DirectionForWind maps a wind surface to its load direction.
func (w WindSurface) DirectionForWind() Direction {
	switch w {
	case WindRoofUplift:
		return DirUplift
	case WindNone:
		return DirGravity
	default:
		return DirLateral
	}
}

// ExposureCategory for wind.
type ExposureCategory string

const (
	ExposureB ExposureCategory = "B"
	ExposureC ExposureCategory = "C"
	ExposureD ExposureCategory = "D"
)

func (e ExposureCategory) Valid() bool {
	switch e {
	case ExposureB, ExposureC, ExposureD:
		return true
	}
	return false
}

// ImportanceCategory for risk classification.
type ImportanceCategory string

const (
	ImportanceI   ImportanceCategory = "I"
	ImportanceII  ImportanceCategory = "II"
	ImportanceIII ImportanceCategory = "III"
	ImportanceIV  ImportanceCategory = "IV"
)

func (i ImportanceCategory) Valid() bool {
	switch i {
	case ImportanceI, ImportanceII, ImportanceIII, ImportanceIV:
		return true
	}
	return false
}

// ImportanceFactorIs returns the snow importance factor per ASCE 7.
func ImportanceFactorIs(i ImportanceCategory) float64 {
	switch i {
	case ImportanceI:
		return 1.00
	case ImportanceII:
		return 1.00
	case ImportanceIII:
		return 1.10
	case ImportanceIV:
		return 1.20
	}
	return 1.00
}

// CheckStatus of a sufficiency check.
type CheckStatus string

const (
	StatusPass    CheckStatus = "pass"    // UR <= 1.00
	StatusMarginal CheckStatus = "marginal" // 1.00 < UR <= 1.05
	StatusFail    CheckStatus = "fail"    // UR > 1.05
)

// GoverningKind names which demand ratio governs UR.
type GoverningKind string

const (
	GovNone GoverningKind = "none"
	GovM    GoverningKind = "moment"
	GovV    GoverningKind = "shear"
	GovP    GoverningKind = "axial"
)

// CombinationKind distinguishes strength vs service limit states.
type CombinationKind string

const (
	CombStrength CombinationKind = "strength"
	CombService  CombinationKind = "service"
)

func (c CombinationKind) Valid() bool {
	switch c {
	case CombStrength, CombService:
		return true
	}
	return false
}
