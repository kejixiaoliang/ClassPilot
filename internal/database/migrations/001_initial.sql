CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS classes (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    stage TEXT NOT NULL,
    grade TEXT NOT NULL,
    school_year TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('current', 'archived')),
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    archived_at TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS one_current_class
ON classes(status)
WHERE status = 'current';

CREATE TABLE IF NOT EXISTS students (
    id TEXT PRIMARY KEY,
    class_id TEXT NOT NULL REFERENCES classes(id) ON DELETE RESTRICT,
    name TEXT NOT NULL,
    gender TEXT NOT NULL DEFAULT '',
    student_number TEXT,
    birth_date TEXT,
    enrollment_date TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    deleted_at TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_student_number_per_class
ON students(class_id, student_number)
WHERE student_number IS NOT NULL
  AND student_number <> ''
  AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS students_class_id ON students(class_id);
CREATE INDEX IF NOT EXISTS students_name ON students(name);

CREATE TABLE IF NOT EXISTS custom_fields (
    id TEXT PRIMARY KEY,
    class_id TEXT NOT NULL REFERENCES classes(id) ON DELETE RESTRICT,
    name TEXT NOT NULL,
    field_type TEXT NOT NULL CHECK (
        field_type IN (
            'text',
            'long_text',
            'number',
            'date',
            'single_choice',
            'multi_choice',
            'boolean'
        )
    ),
    description TEXT NOT NULL DEFAULT '',
    required INTEGER NOT NULL DEFAULT 0 CHECK (required IN (0, 1)),
    enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
    show_in_list INTEGER NOT NULL DEFAULT 0 CHECK (show_in_list IN (0, 1)),
    default_value_json TEXT,
    options_json TEXT NOT NULL DEFAULT '[]',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_custom_field_name
ON custom_fields(class_id, name);

CREATE TABLE IF NOT EXISTS student_field_values (
    student_id TEXT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    field_id TEXT NOT NULL REFERENCES custom_fields(id) ON DELETE RESTRICT,
    value_json TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (student_id, field_id)
);

CREATE TABLE IF NOT EXISTS attachments (
    id TEXT PRIMARY KEY,
    student_id TEXT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    category TEXT NOT NULL,
    original_name TEXT NOT NULL,
    relative_path TEXT NOT NULL UNIQUE,
    size INTEGER NOT NULL CHECK (size >= 0),
    mime_type TEXT NOT NULL,
    sha256 TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    deleted_at TEXT
);

CREATE INDEX IF NOT EXISTS attachments_student_id ON attachments(student_id);

CREATE TABLE IF NOT EXISTS audit_logs (
    id TEXT PRIMARY KEY,
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    summary TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS audit_logs_entity
ON audit_logs(entity_type, entity_id, created_at DESC);

INSERT OR IGNORE INTO schema_migrations (version, applied_at)
VALUES (1, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'));
