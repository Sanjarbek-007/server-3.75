CREATE EXTENSION postgres_fdw;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL
);

-- http://3.75.208.130
CREATE SERVER server2_fdw FOREIGN DATA WRAPPER postgres_fdw OPTIONS (host '3.120.39.160', port '5432', dbname 'server2_db');
CREATE USER MAPPING FOR postgres SERVER server2_fdw OPTIONS (user 'postgres', password '1111');
CREATE FOREIGN TABLE users_server2 (
    id INTEGER,
    username VARCHAR(50),
    email VARCHAR(100)
) SERVER server2_fdw OPTIONS (schema_name 'public', table_name 'users');