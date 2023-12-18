CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users
(
    id uuid DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
    email varchar(255) NOT NULL UNIQUE,
    username varchar(50) NOT NULL UNIQUE,
    password_hash varchar(255) NOT NULL,
    telegram_id varchar(20) UNIQUE,
    created_at timestamp DEFAULT now() NOT NULL
);

CREATE TABLE urls
(
    id uuid DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
    user_id uuid DEFAULT NULL REFERENCES users(id) ON DELETE CASCADE,
    long_url varchar(2048) NOT NULL,
    short_url varchar(20) NOT NULL UNIQUE,
    redirects int DEFAULT 0,
    created_at timestamp DEFAULT now() NOT NULL
);