-- name: CreateURL :one
INSERT INTO url(original_url)
VALUES ($1)
RETURNING *;

-- name: ListURL :many
SELECT u.*, (SELECT COUNT(*) FROM visitor v WHERE v.url_id = u.id) AS total_visitors
FROM url u
OFFSET $1
LIMIT $2;