CREATE TABLE books (
                       id             BIGSERIAL PRIMARY KEY,
                       title          VARCHAR(255) NOT NULL,
                       author         VARCHAR(255) NOT NULL,
                       published_year INTEGER,
                       created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
                       updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);