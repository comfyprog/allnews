CREATE INDEX IF NOT EXISTS published_idx ON articles (published);
CREATE INDEX IF NOT EXISTS resource_name_idx ON articles (resource_name);
CREATE INDEX IF NOT EXISTS title_idx ON articles ((lower(title)));
