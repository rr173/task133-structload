package store

// Schema is the authoritative DDL. Run on Open via Migrate (CREATE IF NOT EXISTS).
const Schema = `
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    site TEXT NOT NULL,
    wind_zone TEXT NOT NULL,
    snow_zone TEXT NOT NULL,
    exposure TEXT NOT NULL,
    importance TEXT NOT NULL,
    wind_speed INTEGER NOT NULL,
    kzt INTEGER NOT NULL,
    kd INTEGER NOT NULL,
    gust_g INTEGER NOT NULL,
    ground_snow INTEGER NOT NULL,
    snow_exposure_factor INTEGER NOT NULL,
    thermal_factor INTEGER NOT NULL,
    units TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS levels (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    elevation INTEGER NOT NULL,
    gross_area INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_levels_project ON levels(project_id);

CREATE TABLE IF NOT EXISTS components (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    level_id TEXT NOT NULL,
    code TEXT NOT NULL,
    type TEXT NOT NULL,
    span INTEGER NOT NULL,
    tributary_area INTEGER NOT NULL,
    tributary_width INTEGER NOT NULL,
    kll INTEGER NOT NULL,
    nominal_moment INTEGER NOT NULL,
    nominal_axial INTEGER NOT NULL,
    nominal_shear INTEGER NOT NULL,
    phi_b INTEGER NOT NULL,
    phi_c INTEGER NOT NULL,
    phi_v INTEGER NOT NULL,
    wind_surface TEXT NOT NULL,
    roof_slope INTEGER NOT NULL,
    section_label TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_components_project ON components(project_id);
CREATE INDEX IF NOT EXISTS idx_components_level ON components(level_id);

CREATE TABLE IF NOT EXISTS load_cases (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    component_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    load_type TEXT NOT NULL,
    magnitude INTEGER NOT NULL,
    direction TEXT NOT NULL,
    origin TEXT NOT NULL,
    source_event_id TEXT NOT NULL,
    note TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY(component_id) REFERENCES components(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_loadcases_component ON load_cases(component_id);
CREATE INDEX IF NOT EXISTS idx_loadcases_project ON load_cases(project_id);
CREATE INDEX IF NOT EXISTS idx_loadcases_origin ON load_cases(origin);

CREATE TABLE IF NOT EXISTS load_combinations (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    coeff_d INTEGER NOT NULL,
    coeff_l INTEGER NOT NULL,
    coeff_lr INTEGER NOT NULL,
    coeff_s INTEGER NOT NULL,
    coeff_w INTEGER NOT NULL,
    coeff_r INTEGER NOT NULL,
    coeff_e INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_combinations_project ON load_combinations(project_id);

CREATE TABLE IF NOT EXISTS checks (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    combination_id TEXT NOT NULL,
    component_id TEXT NOT NULL,
    demand_moment INTEGER NOT NULL,
    demand_shear INTEGER NOT NULL,
    demand_axial INTEGER NOT NULL,
    capacity_moment INTEGER NOT NULL,
    capacity_shear INTEGER NOT NULL,
    capacity_axial INTEGER NOT NULL,
    ur INTEGER NOT NULL,
    status TEXT NOT NULL,
    governing_kind TEXT NOT NULL,
    computed_at INTEGER NOT NULL,
    UNIQUE(combination_id, component_id),
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY(combination_id) REFERENCES load_combinations(id) ON DELETE CASCADE,
    FOREIGN KEY(component_id) REFERENCES components(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_checks_project ON checks(project_id);
CREATE INDEX IF NOT EXISTS idx_checks_component ON checks(component_id);

CREATE TABLE IF NOT EXISTS overrides (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    component_id TEXT NOT NULL,
    field TEXT NOT NULL,
    old_value TEXT NOT NULL,
    new_value TEXT NOT NULL,
    reason TEXT NOT NULL,
    applied_at INTEGER NOT NULL,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY(component_id) REFERENCES components(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_overrides_project ON overrides(project_id);

CREATE TABLE IF NOT EXISTS event_log (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    seq INTEGER NOT NULL,
    event_type TEXT NOT NULL,
    payload TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_eventlog_project ON event_log(project_id);
CREATE INDEX IF NOT EXISTS idx_eventlog_seq ON event_log(project_id, seq);
`
