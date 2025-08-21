package dashboard

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Dashboard 可视化仪表板
type Dashboard struct {
	templates map[string]*template.Template
	upgrader  websocket.Upgrader
	clients   map[string]*websocket.Conn
}

// DashboardData 仪表板数据
type DashboardData struct {
	Title       string            `json:"title"`
	LastUpdate  time.Time         `json:"last_update"`
	Summary     *SummaryCard      `json:"summary"`
	Charts      []*ChartData      `json:"charts"`
	Tables      []*TableData      `json:"tables"`
	Alerts      []*AlertData      `json:"alerts"`
	SLGSpecific *SLGDashboardData `json:"slg_specific"`
}

// SummaryCard 摘要卡片
type SummaryCard struct {
	TotalTests   int     `json:"total_tests"`
	RunningTests int     `json:"running_tests"`
	PassedTests  int     `json:"passed_tests"`
	FailedTests  int     `json:"failed_tests"`
	AverageScore float64 `json:"average_score"`
	SuccessRate  float64 `json:"success_rate"`
}

// ChartData 图表数据
type ChartData struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // "line", "bar", "pie", "gauge", "heatmap"
	Title    string                 `json:"title"`
	Data     interface{}            `json:"data"`
	Options  map[string]interface{} `json:"options"`
	RealTime bool                   `json:"real_time"`
}

// TableData 表格数据
type TableData struct {
	ID      string          `json:"id"`
	Title   string          `json:"title"`
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
	Actions []string        `json:"actions,omitempty"`
}

// AlertData 警报数据
type AlertData struct {
	ID        string    `json:"id"`
	Level     string    `json:"level"` // "success", "info", "warning", "danger"
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Actions   []string  `json:"actions,omitempty"`
}

// SLGDashboardData SLG游戏特定仪表板数据
type SLGDashboardData struct {
	UnityConnections    int                   `json:"unity_connections"`
	ActiveBattles       int                   `json:"active_battles"`
	AverageFPS          float64               `json:"average_fps"`
	BattleMetrics       *BattleMetricsWidget  `json:"battle_metrics"`
	PlayerActionMetrics *PlayerActionWidget   `json:"player_action_metrics"`
	NetworkQuality      *NetworkQualityWidget `json:"network_quality"`
}

// BattleMetricsWidget 战斗指标组件
type BattleMetricsWidget struct {
	BattlesStarted   int           `json:"battles_started"`
	BattlesCompleted int           `json:"battles_completed"`
	AverageInitTime  time.Duration `json:"average_init_time"`
	AverageDuration  time.Duration `json:"average_duration"`
	SyncErrorRate    float64       `json:"sync_error_rate"`
	UpdateFrequency  float64       `json:"update_frequency"`
}

// PlayerActionWidget 玩家操作指标组件
type PlayerActionWidget struct {
	TotalActions     int            `json:"total_actions"`
	ActionsPerSecond float64        `json:"actions_per_second"`
	AverageLatency   time.Duration  `json:"average_latency"`
	ActionTypes      map[string]int `json:"action_types"`
	ErrorRate        float64        `json:"error_rate"`
}

// NetworkQualityWidget 网络质量组件
type NetworkQualityWidget struct {
	QualityScore        float64       `json:"quality_score"`
	PacketLoss          float64       `json:"packet_loss"`
	Jitter              time.Duration `json:"jitter"`
	Bandwidth           float64       `json:"bandwidth"`
	ConnectionStability float64       `json:"connection_stability"`
}

// TimeSeriesData 时序数据
type TimeSeriesData struct {
	Labels []string     `json:"labels"`
	Series []SeriesData `json:"series"`
}

// SeriesData 系列数据
type SeriesData struct {
	Name  string    `json:"name"`
	Data  []float64 `json:"data"`
	Color string    `json:"color,omitempty"`
}

// NewDashboard 创建新的仪表板
func NewDashboard() *Dashboard {
	return &Dashboard{
		templates: make(map[string]*template.Template),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[string]*websocket.Conn),
	}
}

// RegisterRoutes 注册路由
func (d *Dashboard) RegisterRoutes(router *mux.Router) {
	// HTML页面路由
	router.HandleFunc("/dashboard", d.handleDashboard).Methods("GET")
	router.HandleFunc("/dashboard/test/{id}", d.handleTestDetail).Methods("GET")
	router.HandleFunc("/dashboard/slg", d.handleSLGDashboard).Methods("GET")
	router.HandleFunc("/dashboard/reports", d.handleReports).Methods("GET")

	// API路由
	router.HandleFunc("/api/dashboard/data", d.handleDashboardData).Methods("GET")
	router.HandleFunc("/api/dashboard/test/{id}/data", d.handleTestData).Methods("GET")
	router.HandleFunc("/api/dashboard/slg/data", d.handleSLGData).Methods("GET")

	// WebSocket路由
	router.HandleFunc("/ws/dashboard", d.handleWebSocket)

	// 静态资源
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir("./web/static/"))))
}

// handleDashboard 处理主仪表板页面
func (d *Dashboard) handleDashboard(w http.ResponseWriter, r *http.Request) {
	data := d.generateDashboardData()
	d.renderTemplate(w, "dashboard.html", data)
}

// handleTestDetail 处理测试详情页面
func (d *Dashboard) handleTestDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testID := vars["id"]

	data := d.generateTestDetailData(testID)
	d.renderTemplate(w, "test_detail.html", data)
}

// handleSLGDashboard 处理SLG专用仪表板
func (d *Dashboard) handleSLGDashboard(w http.ResponseWriter, r *http.Request) {
	data := d.generateSLGDashboardData()
	d.renderTemplate(w, "slg_dashboard.html", data)
}

// handleReports 处理报告页面
func (d *Dashboard) handleReports(w http.ResponseWriter, r *http.Request) {
	data := d.generateReportsData()
	d.renderTemplate(w, "reports.html", data)
}

// handleDashboardData 处理仪表板数据API
func (d *Dashboard) handleDashboardData(w http.ResponseWriter, r *http.Request) {
	data := d.generateDashboardData()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// handleTestData 处理测试数据API
func (d *Dashboard) handleTestData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testID := vars["id"]

	data := d.generateTestDetailData(testID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// handleSLGData 处理SLG数据API
func (d *Dashboard) handleSLGData(w http.ResponseWriter, r *http.Request) {
	data := d.generateSLGDashboardData()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// handleWebSocket 处理WebSocket连接
func (d *Dashboard) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := d.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	clientID := fmt.Sprintf("client_%d", time.Now().Unix())
	d.clients[clientID] = conn
	defer delete(d.clients, clientID)

	// 定期推送实时数据
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			realTimeData := d.generateRealTimeData()
			if err := conn.WriteJSON(realTimeData); err != nil {
				return
			}
		}
	}
}

// generateDashboardData 生成主仪表板数据
func (d *Dashboard) generateDashboardData() *DashboardData {
	return &DashboardData{
		Title:      "Unity SLG 测试平台",
		LastUpdate: time.Now(),
		Summary: &SummaryCard{
			TotalTests:   156,
			RunningTests: 3,
			PassedTests:  142,
			FailedTests:  11,
			AverageScore: 87.5,
			SuccessRate:  91.0,
		},
		Charts: []*ChartData{
			{
				ID:       "performance_trend",
				Type:     "line",
				Title:    "性能趋势",
				RealTime: true,
				Data: &TimeSeriesData{
					Labels: []string{"00:00", "00:05", "00:10", "00:15", "00:20"},
					Series: []SeriesData{
						{
							Name:  "延迟",
							Data:  []float64{45, 42, 38, 41, 39},
							Color: "#3498db",
						},
						{
							Name:  "吞吐量",
							Data:  []float64{1200, 1250, 1300, 1275, 1320},
							Color: "#2ecc71",
						},
					},
				},
			},
			{
				ID:    "test_distribution",
				Type:  "pie",
				Title: "测试类型分布",
				Data: map[string]interface{}{
					"labels": []string{"集成测试", "压力测试", "基准测试", "模糊测试"},
					"data":   []float64{45, 25, 20, 10},
					"colors": []string{"#3498db", "#e74c3c", "#f39c12", "#9b59b6"},
				},
			},
			{
				ID:    "slg_battle_metrics",
				Type:  "gauge",
				Title: "SLG战斗性能评分",
				Data: map[string]interface{}{
					"value": 87.5,
					"min":   0,
					"max":   100,
					"thresholds": map[string]float64{
						"excellent": 90,
						"good":      80,
						"average":   70,
						"poor":      60,
					},
				},
			},
		},
		Tables: []*TableData{
			{
				ID:      "recent_tests",
				Title:   "最近测试",
				Headers: []string{"测试ID", "类型", "状态", "得分", "开始时间", "操作"},
				Rows: [][]interface{}{
					{"test_001", "Unity集成", "完成", 92.5, "2024-01-27 10:30:00", "查看详情"},
					{"test_002", "压力测试", "运行中", "-", "2024-01-27 10:45:00", "实时监控"},
					{"test_003", "基准测试", "完成", 85.2, "2024-01-27 09:15:00", "查看报告"},
				},
				Actions: []string{"view", "monitor", "report"},
			},
		},
		Alerts: []*AlertData{
			{
				ID:        "alert_001",
				Level:     "warning",
				Title:     "延迟异常",
				Message:   "测试 test_002 中检测到延迟峰值 >150ms",
				Timestamp: time.Now().Add(-5 * time.Minute),
			},
			{
				ID:        "alert_002",
				Level:     "info",
				Title:     "测试完成",
				Message:   "Unity集成测试 test_001 已完成，得分 92.5分",
				Timestamp: time.Now().Add(-10 * time.Minute),
			},
		},
		SLGSpecific: &SLGDashboardData{
			UnityConnections: 12,
			ActiveBattles:    3,
			AverageFPS:       58.5,
			BattleMetrics: &BattleMetricsWidget{
				BattlesStarted:   45,
				BattlesCompleted: 42,
				AverageInitTime:  850 * time.Millisecond,
				AverageDuration:  125 * time.Second,
				SyncErrorRate:    0.003,
				UpdateFrequency:  20.0,
			},
			PlayerActionMetrics: &PlayerActionWidget{
				TotalActions:     1250,
				ActionsPerSecond: 15.5,
				AverageLatency:   25 * time.Millisecond,
				ActionTypes: map[string]int{
					"move":   450,
					"attack": 320,
					"skill":  280,
					"chat":   200,
				},
				ErrorRate: 0.02,
			},
			NetworkQuality: &NetworkQualityWidget{
				QualityScore:        92.5,
				PacketLoss:          0.001,
				Jitter:              5 * time.Millisecond,
				Bandwidth:           125.5,
				ConnectionStability: 98.5,
			},
		},
	}
}

// generateTestDetailData 生成测试详情数据
func (d *Dashboard) generateTestDetailData(testID string) *DashboardData {
	return &DashboardData{
		Title:      fmt.Sprintf("测试详情 - %s", testID),
		LastUpdate: time.Now(),
		Charts: []*ChartData{
			{
				ID:       "latency_timeline",
				Type:     "line",
				Title:    "延迟时间线",
				RealTime: true,
				Data: &TimeSeriesData{
					Labels: []string{"10:00", "10:05", "10:10", "10:15", "10:20", "10:25"},
					Series: []SeriesData{
						{
							Name:  "平均延迟",
							Data:  []float64{45, 42, 48, 41, 39, 43},
							Color: "#3498db",
						},
						{
							Name:  "P95延迟",
							Data:  []float64{85, 82, 95, 88, 79, 83},
							Color: "#e74c3c",
						},
					},
				},
			},
			{
				ID:    "message_distribution",
				Type:  "bar",
				Title: "消息类型分布",
				Data: map[string]interface{}{
					"labels": []string{"LOGIN", "HEARTBEAT", "BATTLE", "CHAT", "ERROR"},
					"data":   []float64{150, 800, 450, 200, 12},
					"colors": []string{"#2ecc71", "#3498db", "#e74c3c", "#f39c12", "#95a5a6"},
				},
			},
		},
	}
}

// generateSLGDashboardData 生成SLG专用仪表板数据
func (d *Dashboard) generateSLGDashboardData() *SLGDashboardData {
	return &SLGDashboardData{
		UnityConnections: 15,
		ActiveBattles:    5,
		AverageFPS:       59.2,
		BattleMetrics: &BattleMetricsWidget{
			BattlesStarted:   67,
			BattlesCompleted: 62,
			AverageInitTime:  750 * time.Millisecond,
			AverageDuration:  135 * time.Second,
			SyncErrorRate:    0.002,
			UpdateFrequency:  22.0,
		},
		PlayerActionMetrics: &PlayerActionWidget{
			TotalActions:     2350,
			ActionsPerSecond: 18.5,
			AverageLatency:   22 * time.Millisecond,
			ActionTypes: map[string]int{
				"move":   850,
				"attack": 620,
				"skill":  520,
				"chat":   360,
			},
			ErrorRate: 0.015,
		},
		NetworkQuality: &NetworkQualityWidget{
			QualityScore:        94.2,
			PacketLoss:          0.0008,
			Jitter:              3 * time.Millisecond,
			Bandwidth:           145.8,
			ConnectionStability: 99.1,
		},
	}
}

// generateReportsData 生成报告数据
func (d *Dashboard) generateReportsData() *DashboardData {
	return &DashboardData{
		Title:      "测试报告中心",
		LastUpdate: time.Now(),
		Tables: []*TableData{
			{
				ID:      "test_reports",
				Title:   "测试报告列表",
				Headers: []string{"报告ID", "测试类型", "测试时间", "得分", "等级", "状态", "操作"},
				Rows: [][]interface{}{
					{"RPT_001", "Unity集成", "2024-01-27 10:30", 92.5, "A", "已完成", "下载PDF"},
					{"RPT_002", "压力测试", "2024-01-27 09:15", 85.2, "B+", "已完成", "查看详情"},
					{"RPT_003", "基准测试", "2024-01-27 08:45", 78.8, "C+", "已完成", "对比分析"},
				},
				Actions: []string{"download", "view", "compare"},
			},
		},
	}
}

// generateRealTimeData 生成实时数据
func (d *Dashboard) generateRealTimeData() map[string]interface{} {
	return map[string]interface{}{
		"timestamp": time.Now(),
		"metrics": map[string]interface{}{
			"active_connections": 12 + (time.Now().Unix() % 5),
			"current_latency":    40 + float64(time.Now().Unix()%20),
			"throughput":         1200 + float64(time.Now().Unix()%100),
			"error_rate":         0.01 + float64(time.Now().Unix()%10)/1000,
		},
		"slg_metrics": map[string]interface{}{
			"unity_fps":      55 + float64(time.Now().Unix()%10),
			"battle_count":   3 + (time.Now().Unix() % 3),
			"player_actions": 15.5 + float64(time.Now().Unix()%5),
		},
	}
}

// renderTemplate 渲染模板
func (d *Dashboard) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	tmpl, exists := d.templates[templateName]
	if !exists {
		// 动态加载模板
		tmpl = template.Must(template.ParseFiles(fmt.Sprintf("web/templates/%s", templateName)))
		d.templates[templateName] = tmpl
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// BroadcastUpdate 广播更新到所有客户端
func (d *Dashboard) BroadcastUpdate(data interface{}) {
	for clientID, conn := range d.clients {
		if err := conn.WriteJSON(data); err != nil {
			// 连接已断开，清理
			delete(d.clients, clientID)
		}
	}
}
