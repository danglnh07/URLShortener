-- Create table url
CREATE TABLE IF NOT EXISTS url (
    id SERIAL PRIMARY KEY,
    original_url VARCHAR NOT NULL UNIQUE, -- Original full URL, no max size
    time_created TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create table visitor
CREATE TABLE IF NOT EXISTS visitor (
    ip VARCHAR(45) NOT NULL, -- IP address (both IPv4 and IPv6) can have a maximum of 45 characters
    url_id SERIAL NOT NULL REFERENCES url(id),
    PRIMARY KEY (Ip, url_id),
    time_visited TIMESTAMPTZ NOT NULL DEFAULT now()
)