CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    surname VARCHAR(100) NOT NULL,
    avatar_url TEXT NOT NULL,
    age INTEGER NOT NULL,
    country VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS buisness(
    id uuid PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);