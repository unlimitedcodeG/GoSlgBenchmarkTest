-- name: CreateTest :one
INSERT INTO tests (
    name, type, status, config
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetTest :one
SELECT * FROM tests WHERE id = $1 AND deleted_at IS NULL;

-- name: GetTestList :many
SELECT * FROM tests 
WHERE deleted_at IS NULL
  AND ($1::text IS NULL OR type = $1)
  AND ($2::text IS NULL OR status = $2)
  AND ($3::text IS NULL OR name ILIKE '%' || $3 || '%')
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountTests :one
SELECT COUNT(*) FROM tests 
WHERE deleted_at IS NULL
  AND ($1::text IS NULL OR type = $1)
  AND ($2::text IS NULL OR status = $2)
  AND ($3::text IS NULL OR name ILIKE '%' || $3 || '%');

-- name: UpdateTest :one
UPDATE tests 
SET status = $2, start_time = $3, end_time = $4, duration = $5, 
    score = $6, grade = $7, error_message = $8, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteTest :exec
UPDATE tests SET deleted_at = NOW() WHERE id = $1;

-- name: GetTestStats :one
SELECT 
    COUNT(*) as total_tests,
    COUNT(CASE WHEN status = 'running' THEN 1 END) as running_tests,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_tests,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_tests
FROM tests WHERE deleted_at IS NULL;