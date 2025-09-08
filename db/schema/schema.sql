-- Create table url
CREATE TABLE IF NOT EXISTS url (
    id SERIAL PRIMARY KEY,
    domain VARCHAR(20) NOT NULL, -- Domain of the URL, like www.facebook.com, which won't change
    original_url VARCHAR(200) NOT NULL, -- Original full URL 
    shorten_id VARCHAR(64) NOT NULL, -- The generated ID used in the shorten URL.
    time_created TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create table visitor
CREATE TABLE IF NOT EXISTS visitor (
    ip VARCHAR(21) NOT NULL, -- IPv4 address can have a maximum of 21 characters
    url_id SERIAL NOT NULL REFERENCES url(id),
    PRIMARY KEY (Ip, url_id),
    time_visited TIMESTAMPTZ NOT NULL DEFAULT now()
)