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
    created_at timestamp DEFAULT now() NOT NULL
);

CREATE TABLE statistics
(
    id uuid DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
    url_id uuid REFERENCES urls(id) ON DELETE CASCADE NOT NULL,
    ip varchar(15) NOT NULL,
    user_agent varchar(255) NOT NULL,
    timestamp timestamp DEFAULT now() NOT NULL
);