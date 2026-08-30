-- DOWN
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    sector TEXT NOT NULL,
    industry TEXT
);

ALTER TABLE assets ADD COLUMN category_id UUID REFERENCES categories(id);
