-- name: CreateURL :one
INSERT INTO url(original_url)
VALUES ($1)
RETURNING *;

-- name: GetURL :one
SELECT original_url FROM url 
WHERE id = $1;

-- name: ListURL :many
SELECT u.*, (SELECT COUNT(*) FROM visitor v WHERE v.url_id = u.id) AS total_visitors
FROM url u
OFFSET $1
LIMIT $2;