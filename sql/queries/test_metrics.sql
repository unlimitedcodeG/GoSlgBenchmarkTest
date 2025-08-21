-- name: CreateTestMetric :one
INSERT INTO test_metrics (
    test_id, metric_type, metric_name, metric_value, metric_unit, timestamp
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetTestMetrics :many
SELECT * FROM test_metrics 
WHERE test_id = $1 
ORDER BY timestamp DESC;

-- name: GetTestMetricsByType :many
SELECT * FROM test_metrics 
WHERE test_id = $1 AND metric_type = $2
ORDER BY timestamp DESC;

-- name: GetRecentMetrics :many
SELECT * FROM test_metrics 
WHERE timestamp > $1
ORDER BY timestamp DESC
LIMIT $2;