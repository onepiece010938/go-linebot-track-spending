-- name: CreateUser :one
INSERT INTO users (
  lineid,name,month_limit,created_at,updated_at
) VALUES (
  $1,$2,$3,$4,$5
)
RETURNING *;