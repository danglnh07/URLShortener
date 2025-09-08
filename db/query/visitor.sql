-- name: CreateVisitor :one
INSERT INTO visitor(ip, url_id)
VALUES ($1, $2)
RETURNING *;

-- name: ListVisitor :many
SELECT * FROM visitor 
WHERE url_id = $1
OFFSET $2
LIMIT $3;
