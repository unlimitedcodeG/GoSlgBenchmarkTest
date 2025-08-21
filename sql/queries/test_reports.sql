-- name: CreateTestReport :one
INSERT INTO test_reports (
    test_id, summary, issues, suggestions, raw_data
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetTestReports :many
SELECT * FROM test_reports WHERE test_id = $1 ORDER BY created_at DESC;

-- name: GetTestReport :one
SELECT * FROM test_reports WHERE id = $1;