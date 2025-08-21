-- name: CreateSessionMessage :one
INSERT INTO session_messages (
    session_id, direction, opcode, message_size, body_size, 
    sequence_num, message_hash, raw_data, timestamp
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetSessionMessages :many
SELECT * FROM session_messages 
WHERE session_id = $1 
ORDER BY timestamp DESC
LIMIT $2 OFFSET $3;

-- name: GetSessionMessagesByDirection :many
SELECT * FROM session_messages 
WHERE session_id = $1 AND direction = $2
ORDER BY timestamp DESC;

-- name: CountSessionMessages :one
SELECT COUNT(*) FROM session_messages WHERE session_id = $1;