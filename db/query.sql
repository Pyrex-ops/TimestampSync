-- name: GetAllTimestamps :many
SELECT *
FROM timestamps;

-- name: GetTimestampById :one
SELECT *
FROM timestamps
WHERE id = ?;

-- name: GetTimestampsByName :many
SELECT *
FROM timestamps
WHERE name = ?;

-- name: GetTimestampsByTimestamp :many
SELECT *
FROM timestamps
WHERE timestamp = ?;

-- name: GetTimestampsByTimestampRange :many
SELECT *
FROM timestamps
WHERE timestamp BETWEEN ? AND ?;

-- name: InsertTimestamp :exec
INSERT INTO timestamps (id, name, seconds, timestamp)
VALUES (?, ?, ?, ?);

-- name: UpdateTimestampById :exec
UPDATE timestamps
SET name      = ?,
    timestamp = ?,
    seconds   = ?
WHERE id = ?;

-- name: UpdateTimestampByName :exec
UPDATE timestamps
SET timestamp = ?,
    seconds   = ?
WHERE name = ?;

-- name: UpdateTimestampOnly :exec
UPDATE timestamps
SET timestamp = ?
WHERE id = ?;

-- name: UpdateNameOnly :exec
UPDATE timestamps
SET name = ?
WHERE id = ?;

-- name: DeleteTimestampById :exec
DELETE
FROM timestamps
WHERE id = ?;

-- name: DeleteTimestampsByName :exec
DELETE
FROM timestamps
WHERE name = ?;

-- name: CountAllTimestamps :one
SELECT COUNT(*)
FROM timestamps;

-- name: CountTimestampsByName :one
SELECT COUNT(*)
FROM timestamps
WHERE name = ?;

-- name: DeleteOldTimestamps :exec
DELETE
FROM timestamps
WHERE timestamp < DATE('now', '-30 days');