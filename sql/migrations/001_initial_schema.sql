-- +goose Up
-- 创建测试主表
CREATE TABLE tests (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    duration BIGINT,
    score DECIMAL,
    grade VARCHAR(10),
    config TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_tests_deleted_at ON tests (deleted_at);
CREATE INDEX idx_tests_status ON tests (status);
CREATE INDEX idx_tests_type ON tests (type);

-- 创建测试报告表
CREATE TABLE test_reports (
    id BIGSERIAL PRIMARY KEY,
    test_id BIGINT NOT NULL REFERENCES tests(id),
    summary TEXT,
    issues JSONB,
    suggestions JSONB,
    raw_data TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_reports_test_id ON test_reports (test_id);

-- 创建测试指标表
CREATE TABLE test_metrics (
    id BIGSERIAL PRIMARY KEY,
    test_id BIGINT NOT NULL REFERENCES tests(id),
    metric_type VARCHAR(100) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL,
    metric_unit VARCHAR(50),
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_metrics_test_id ON test_metrics (test_id);
CREATE INDEX idx_test_metrics_type ON test_metrics (metric_type);
CREATE INDEX idx_test_metrics_timestamp ON test_metrics (timestamp);

-- 创建测试会话表
CREATE TABLE test_sessions (
    id BIGSERIAL PRIMARY KEY,
    test_id BIGINT NOT NULL REFERENCES tests(id),
    session_id VARCHAR(255) NOT NULL,
    client_type VARCHAR(100),
    client_info JSONB,
    connection_count INTEGER DEFAULT 0,
    message_count INTEGER DEFAULT 0,
    bytes_transfer BIGINT DEFAULT 0,
    avg_latency BIGINT DEFAULT 0,
    min_latency BIGINT DEFAULT 0,
    max_latency BIGINT DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    reconnect_count INTEGER DEFAULT 0,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_sessions_test_id ON test_sessions (test_id);
CREATE INDEX idx_test_sessions_session_id ON test_sessions (session_id);

-- 创建会话事件表
CREATE TABLE session_events (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES test_sessions(id),
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_session_events_session_id ON session_events (session_id);
CREATE INDEX idx_session_events_type ON session_events (event_type);
CREATE INDEX idx_session_events_timestamp ON session_events (timestamp);

-- 创建会话消息表
CREATE TABLE session_messages (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES test_sessions(id),
    direction VARCHAR(20) NOT NULL,
    opcode INTEGER,
    message_size INTEGER,
    body_size INTEGER,
    sequence_num BIGINT,
    message_hash VARCHAR(64),
    raw_data BYTEA,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_session_messages_session_id ON session_messages (session_id);
CREATE INDEX idx_session_messages_direction ON session_messages (direction);
CREATE INDEX idx_session_messages_timestamp ON session_messages (timestamp);

-- 创建SLG战斗记录表
CREATE TABLE slg_battle_records (
    id BIGSERIAL PRIMARY KEY,
    test_id BIGINT NOT NULL REFERENCES tests(id),
    battle_id VARCHAR(255) NOT NULL,
    battle_type VARCHAR(100),
    player_ids JSONB,
    map_id VARCHAR(100),
    winner VARCHAR(255),
    battle_duration BIGINT,
    init_latency BIGINT,
    update_frequency DECIMAL,
    sync_error_rate DECIMAL,
    unity_fps DECIMAL,
    battle_data JSONB,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_slg_battle_records_test_id ON slg_battle_records (test_id);
CREATE INDEX idx_slg_battle_records_battle_id ON slg_battle_records (battle_id);

-- 创建Unity客户端记录表
CREATE TABLE unity_client_records (
    id BIGSERIAL PRIMARY KEY,
    test_id BIGINT NOT NULL REFERENCES tests(id),
    player_id VARCHAR(255) NOT NULL,
    unity_version VARCHAR(100),
    platform VARCHAR(100),
    device_info JSONB,
    connect_time TIMESTAMPTZ NOT NULL,
    disconnect_time TIMESTAMPTZ,
    total_duration BIGINT,
    action_count INTEGER DEFAULT 0,
    avg_fps DECIMAL,
    memory_usage DECIMAL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_unity_client_records_test_id ON unity_client_records (test_id);
CREATE INDEX idx_unity_client_records_player_id ON unity_client_records (player_id);

-- +goose Down
DROP TABLE IF EXISTS unity_client_records;
DROP TABLE IF EXISTS slg_battle_records;
DROP TABLE IF EXISTS session_messages;
DROP TABLE IF EXISTS session_events;
DROP TABLE IF EXISTS test_sessions;
DROP TABLE IF EXISTS test_metrics;
DROP TABLE IF EXISTS test_reports;
DROP TABLE IF EXISTS tests;