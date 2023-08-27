CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users
(
    id uuid DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
    email varchar(255) NOT NULL UNIQUE,
    username varchar(50) NOT NULL UNIQUE,
    password_hash varchar(255) NOT NULL,
    created_at timestamp DEFAULT now() NOT NULL
);

CREATE TABLE urls
(
    id uuid DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
    user_id uuid REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    long_url varchar(2048) NOT NULL,
    short_url varchar(10) NOT NULL UNIQUE,
    redirects int DEFAULT 0,
    created_at timestamp DEFAULT now() NOT NULL
);