-- name: CreateSLGBattle :one
INSERT INTO slg_battle_records (
    test_id, battle_id, battle_type, player_ids, map_id, battle_data, start_time
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetSLGBattle :one
SELECT * FROM slg_battle_records WHERE battle_id = $1;

-- name: GetSLGBattles :many
SELECT * FROM slg_battle_records WHERE test_id = $1 ORDER BY start_time DESC;

-- name: UpdateSLGBattle :one
UPDATE slg_battle_records 
SET winner = $2, battle_duration = $3, init_latency = $4, 
    update_frequency = $5, sync_error_rate = $6, unity_fps = $7,
    battle_data = $8, end_time = $9, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetSLGBattleStats :one
SELECT 
    COUNT(*) as total_battles,
    COUNT(CASE WHEN end_time IS NOT NULL THEN 1 END) as completed_battles,
    AVG(battle_duration) as avg_duration,
    AVG(init_latency) as avg_init_latency,
    AVG(unity_fps) as avg_fps
FROM slg_battle_records WHERE test_id = $1;