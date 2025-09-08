-- name: CreateVisitor :one
INSERT INTO visitor(ip, url_id)
VALUES ($1, $2)
RETURNING *;

-- name: ListVisitor :many
SELECT v.ip, v.time_visited, v.url_id, u.original_url FROM visitor v
JOIN url u ON u.id = v.url_id
WHERE url_id = $1
OFFSET $2
LIMIT $3;
