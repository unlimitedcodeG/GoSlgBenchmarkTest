package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"GoSlgBenchmarkTest/internal/db"
)

// PgxPool 全局连接池
var PgxPool *pgxpool.Pool

// Queries 全局查询实例
var Queries *db.Queries

// Config 数据库配置
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "2020",
		DBName:   "postgres",
		SSLMode:  "disable",
	}
}

// ConnectPgx 连接PostgreSQL数据库
func ConnectPgx(config *Config) error {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.DBName, config.SSLMode,
	)

	// 配置连接池
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	// 设置连接池参数
	poolConfig.MaxConns = 25                       // 最大连接数
	poolConfig.MinConns = 5                        // 最小连接数
	poolConfig.MaxConnLifetime = time.Hour         // 连接最大生命周期
	poolConfig.MaxConnIdleTime = 30 * time.Minute  // 连接最大空闲时间
	poolConfig.HealthCheckPeriod = 1 * time.Minute // 健康检查周期

	// 创建连接池
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	PgxPool = pool
	Queries = db.New(pool)

	log.Println("✅ PostgreSQL连接池创建成功")
	return nil
}

// ClosePgx 关闭连接池
func ClosePgx() {
	if PgxPool != nil {
		PgxPool.Close()
		log.Println("✅ PostgreSQL连接池已关闭")
	}
}

// GetPgxPool 获取连接池实例
func GetPgxPool() *pgxpool.Pool {
	return PgxPool
}

// GetQueries 获取查询实例
func GetQueries() *db.Queries {
	return Queries
}

// TestConnectionPgx 测试数据库连接
func TestConnectionPgx() error {
	if PgxPool == nil {
		return fmt.Errorf("database pool not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return PgxPool.Ping(ctx)
}

// GetPoolStats 获取连接池统计信息
func GetPoolStats() *pgxpool.Stat {
	if PgxPool == nil {
		return nil
	}
	return PgxPool.Stat()
}
