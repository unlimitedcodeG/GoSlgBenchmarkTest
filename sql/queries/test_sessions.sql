-- name: CreateTestSession :one
INSERT INTO test_sessions (
    test_id, session_id, client_type, client_info, start_time
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetTestSession :one
SELECT * FROM test_sessions WHERE session_id = $1;

-- name: GetTestSessions :many
SELECT * FROM test_sessions WHERE test_id = $1 ORDER BY start_time DESC;

-- name: UpdateTestSession :one
UPDATE test_sessions 
SET end_time = $2, connection_count = $3, message_count = $4, 
    bytes_transfer = $5, avg_latency = $6, min_latency = $7, 
    max_latency = $8, error_count = $9, reconnect_count = $10,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateTestSessionStats :exec
UPDATE test_sessions 
SET message_count = message_count + 1,
    bytes_transfer = bytes_transfer + $2,
    updated_at = NOW()
WHERE id = $1;