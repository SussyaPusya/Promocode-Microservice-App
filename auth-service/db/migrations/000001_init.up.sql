

CREATE TABLE IF NOT EXISTS platform_user (
    id uuid PRIMARY KEY,
    email VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL
)