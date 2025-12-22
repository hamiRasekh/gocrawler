CREATE TABLE IF NOT EXISTS crawler_settings (
    settings_key TEXT PRIMARY KEY,
    value JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO crawler_settings (settings_key, value)
VALUES ('embroidery_payload_overrides', '{}'::jsonb)
ON CONFLICT (settings_key) DO NOTHING;

