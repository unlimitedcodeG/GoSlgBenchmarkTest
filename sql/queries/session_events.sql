-- name: CreateSessionEvent :one
INSERT INTO session_events (
    session_id, event_type, event_data, timestamp
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetSessionEvents :many
SELECT * FROM session_events 
WHERE session_id = $1 
ORDER BY timestamp DESC;

-- name: GetSessionEventsByType :many
SELECT * FROM session_events 
WHERE session_id = $1 AND event_type = $2
ORDER BY timestamp DESC;