-- name: CreateUnityClient :one
INSERT INTO unity_client_records (
    test_id, player_id, unity_version, platform, device_info, connect_time
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetUnityClient :one
SELECT * FROM unity_client_records 
WHERE test_id = $1 AND player_id = $2 
ORDER BY created_at DESC LIMIT 1;

-- name: GetUnityClients :many
SELECT * FROM unity_client_records WHERE test_id = $1 ORDER BY connect_time DESC;

-- name: UpdateUnityClient :one
UPDATE unity_client_records 
SET disconnect_time = $2, total_duration = $3, action_count = $4,
    avg_fps = $5, memory_usage = $6, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: IncrementActionCount :exec
UPDATE unity_client_records 
SET action_count = action_count + 1, updated_at = NOW()
WHERE test_id = $1 AND player_id = $2;

-- name: GetUnityClientStats :one
SELECT 
    COUNT(*) as total_clients,
    COUNT(CASE WHEN disconnect_time IS NOT NULL THEN 1 END) as disconnected_clients,
    AVG(total_duration) as avg_session_duration,
    AVG(action_count) as avg_actions,
    AVG(avg_fps) as avg_fps,
    AVG(memory_usage) as avg_memory
FROM unity_client_records WHERE test_id = $1;