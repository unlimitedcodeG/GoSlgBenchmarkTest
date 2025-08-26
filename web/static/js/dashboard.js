// Dashboard JavaScript - Unity SLG 测试平台

// Axios配置和日志记录
class AxiosLogger {
    constructor() {
        this.enableLogging = this.getLoggingEnabled();
        this.setupAxiosInterceptors();
    }

    getLoggingEnabled() {
        // 检查环境变量或localStorage
        return localStorage.getItem('ENABLE_AXIOS_LOGGING') === 'true' ||
               window.location.search.includes('debug=axios');
    }

    setupAxiosInterceptors() {
        // 请求拦截器
        axios.interceptors.request.use(
            (config) => {
                if (this.enableLogging) {
                    console.log('🚀 Axios Request:', {
                        method: config.method?.toUpperCase(),
                        url: config.url,
                        headers: config.headers,
                        data: config.data,
                        timestamp: new Date().toISOString()
                    });
                }
                return config;
            },
            (error) => {
                if (this.enableLogging) {
                    console.error('❌ Axios Request Error:', error);
                }
                return Promise.reject(error);
            }
        );

        // 响应拦截器
        axios.interceptors.response.use(
            (response) => {
                if (this.enableLogging) {
                    console.log('✅ Axios Response:', {
                        status: response.status,
                        statusText: response.statusText,
                        url: response.config.url,
                        method: response.config.method?.toUpperCase(),
                        headers: response.headers,
                        data: response.data,
                        duration: Date.now() - new Date(response.config.timestamp || Date.now()),
                        timestamp: new Date().toISOString()
                    });
                }
                return response;
            },
            (error) => {
                if (this.enableLogging) {
                    console.error('❌ Axios Response Error:', {
                        status: error.response?.status,
                        statusText: error.response?.statusText,
                        url: error.config?.url,
                        method: error.config?.method?.toUpperCase(),
                        data: error.response?.data,
                        message: error.message,
                        timestamp: new Date().toISOString()
                    });
                }
                return Promise.reject(error);
            }
        );
    }

    toggleLogging() {
        this.enableLogging = !this.enableLogging;
        localStorage.setItem('ENABLE_AXIOS_LOGGING', this.enableLogging);
        console.log(`🔧 Axios logging ${this.enableLogging ? 'enabled' : 'disabled'}`);
        return this.enableLogging;
    }

    logConfig() {
        if (this.enableLogging) {
            console.log('📋 Axios Configuration:', {
                baseURL: axios.defaults.baseURL,
                timeout: axios.defaults.timeout,
                headers: axios.defaults.headers,
                logging: this.enableLogging
            });
        }
    }
}

// 创建全局axios日志记录器实例
const axiosLogger = new AxiosLogger();

class DashboardManager {
    constructor() {
        this.wsConnection = null;
        this.charts = {};
        this.updateInterval = null;
        this.init();
    }

    init() {
        this.connectWebSocket();
        this.setupEventListeners();
    }

    // WebSocket连接
    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/dashboard`;
        
        this.wsConnection = new WebSocket(wsUrl);
        
        this.wsConnection.onopen = () => {
            console.log('✅ WebSocket连接已建立');
            this.showNotification('连接成功', 'WebSocket实时连接已建立', 'success');
        };
        
        this.wsConnection.onmessage = (event) => {
            const data = JSON.parse(event.data);
            this.handleRealtimeUpdate(data);
        };
        
        this.wsConnection.onerror = (error) => {
            console.error('❌ WebSocket错误:', error);
            this.showNotification('连接错误', 'WebSocket连接出现问题', 'warning');
        };
        
        this.wsConnection.onclose = () => {
            console.log('🔌 WebSocket连接已关闭');
            // 尝试重连
            setTimeout(() => this.connectWebSocket(), 5000);
        };
    }

    // 事件监听器设置
    setupEventListeners() {
        // 时间段切换按钮
        document.querySelectorAll('[data-period]').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const period = e.target.dataset.period;
                this.updatePeriod(period);
                
                // 更新按钮状态
                document.querySelectorAll('[data-period]').forEach(b => b.classList.remove('active'));
                e.target.classList.add('active');
            });
        });
    }

    // 处理实时数据更新
    handleRealtimeUpdate(data) {
        console.log('📊 收到实时数据:', data);
        
        // 更新摘要卡片
        if (data.metrics) {
            this.updateSummaryCards(data.metrics);
        }
        
        // 更新SLG指标
        if (data.slg_metrics) {
            this.updateSLGMetrics(data.slg_metrics);
        }
        
        // 更新图表
        this.updateChartsWithRealtimeData(data);
    }

    // 更新摘要卡片
    updateSummaryCards(metrics) {
        // 更新延迟显示
        const latencyElement = document.querySelector('.metric-value[data-metric="latency"]');
        if (latencyElement && metrics.current_latency) {
            latencyElement.textContent = `${metrics.current_latency.toFixed(1)}ms`;
        }
        
        // 更新吞吐量
        const throughputElement = document.querySelector('.metric-value[data-metric="throughput"]');
        if (throughputElement && metrics.throughput) {
            throughputElement.textContent = metrics.throughput.toFixed(0);
        }
    }

    // 更新SLG指标
    updateSLGMetrics(slgMetrics) {
        // 更新Unity FPS
        const fpsElement = document.querySelector('.metric-value[data-metric="fps"]');
        if (fpsElement && slgMetrics.unity_fps) {
            fpsElement.textContent = slgMetrics.unity_fps.toFixed(1);
        }
        
        // 更新战斗数量
        const battleElement = document.querySelector('.metric-value[data-metric="battles"]');
        if (battleElement && slgMetrics.battle_count) {
            battleElement.textContent = slgMetrics.battle_count;
        }
    }

    // 显示通知
    showNotification(title, message, type = 'info') {
        const alertsContainer = document.querySelector('.alerts-container');
        if (!alertsContainer) return;

        const alertClass = type === 'success' ? 'alert-success' : 
                          type === 'warning' ? 'alert-warning' : 
                          type === 'danger' ? 'alert-danger' : 'alert-info';

        const alertHtml = `
            <div class="alert ${alertClass} alert-dismissible fade show" role="alert">
                <h6 class="alert-heading">
                    <i class="fas fa-${type === 'success' ? 'check-circle' : 'info-circle'}"></i>
                    ${title}
                </h6>
                <p class="mb-1">${message}</p>
                <small class="text-muted">${new Date().toLocaleTimeString()}</small>
                <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
            </div>
        `;
        
        alertsContainer.insertAdjacentHTML('afterbegin', alertHtml);
        
        // 自动删除旧警报
        const alerts = alertsContainer.querySelectorAll('.alert');
        if (alerts.length > 5) {
            alerts[alerts.length - 1].remove();
        }
    }

    // 更新时间段
    updatePeriod(period) {
        console.log(`📅 切换到时间段: ${period}`);
        // 这里可以发送API请求获取对应时间段的数据
        this.fetchPeriodData(period).then(data => {
            this.updateChartsWithPeriodData(data);
        });
    }

    // 获取时间段数据
    async fetchPeriodData(period) {
        try {
            const response = await axios.get(`/api/dashboard/data`, {
                params: { period },
                timeout: 10000
            });
            return response.data;
        } catch (error) {
            console.error('获取时间段数据失败:', error);
            return null;
        }
    }

    // 更新图表数据
    updateChartsWithPeriodData(data) {
        if (data && data.charts) {
            data.charts.forEach(chartData => {
                if (this.charts[chartData.id]) {
                    this.updateChart(chartData.id, chartData.data);
                }
            });
        }
    }

    // 更新图表（实时数据）
    updateChartsWithRealtimeData(data) {
        if (data.timestamp && this.charts.performanceTrendChart) {
            const chart = this.charts.performanceTrendChart;
            const option = chart.getOption();
            
            // 添加新数据点
            const newTime = new Date(data.timestamp).toLocaleTimeString();
            option.xAxis[0].data.push(newTime);
            option.series[0].data.push(data.metrics?.current_latency || 0);
            option.series[1].data.push(data.metrics?.throughput || 0);
            
            // 保持最近20个数据点
            if (option.xAxis[0].data.length > 20) {
                option.xAxis[0].data.shift();
                option.series[0].data.shift();
                option.series[1].data.shift();
            }
            
            chart.setOption(option);
        }
    }

    // 更新单个图表
    updateChart(chartId, newData) {
        const chart = this.charts[chartId];
        if (chart) {
            chart.setOption({
                series: [{
                    data: newData
                }]
            });
        }
    }
}

// 初始化性能趋势图
function initPerformanceTrendChart(data) {
    const chartDom = document.getElementById('performanceTrendChart');
    const myChart = echarts.init(chartDom);
    
    const option = {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'cross',
                label: {
                    backgroundColor: '#6a7985'
                }
            }
        },
        legend: {
            data: ['延迟 (ms)', '吞吐量']
        },
        toolbox: {
            feature: {
                saveAsImage: {}
            }
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
        },
        xAxis: [
            {
                type: 'category',
                boundaryGap: false,
                data: data.labels || ['00:00', '00:05', '00:10', '00:15', '00:20']
            }
        ],
        yAxis: [
            {
                type: 'value',
                name: '延迟 (ms)',
                position: 'left',
                axisLine: {
                    lineStyle: {
                        color: '#3498db'
                    }
                }
            },
            {
                type: 'value',
                name: '吞吐量',
                position: 'right',
                axisLine: {
                    lineStyle: {
                        color: '#2ecc71'
                    }
                }
            }
        ],
        series: [
            {
                name: '延迟 (ms)',
                type: 'line',
                yAxisIndex: 0,
                data: data.series ? data.series[0].data : [45, 42, 38, 41, 39],
                smooth: true,
                lineStyle: {
                    color: '#3498db'
                },
                areaStyle: {
                    color: 'rgba(52, 152, 219, 0.1)'
                }
            },
            {
                name: '吞吐量',
                type: 'line',
                yAxisIndex: 1,
                data: data.series ? data.series[1].data : [1200, 1250, 1300, 1275, 1320],
                smooth: true,
                lineStyle: {
                    color: '#2ecc71'
                },
                areaStyle: {
                    color: 'rgba(46, 204, 113, 0.1)'
                }
            }
        ]
    };
    
    myChart.setOption(option);
    dashboard.charts.performanceTrendChart = myChart;
    
    // 响应式调整
    window.addEventListener('resize', () => {
        myChart.resize();
    });
}

// 初始化测试分布图
function initTestDistributionChart(data) {
    const chartDom = document.getElementById('testDistributionChart');
    const myChart = echarts.init(chartDom);
    
    const option = {
        tooltip: {
            trigger: 'item',
            formatter: '{a} <br/>{b}: {c} ({d}%)'
        },
        legend: {
            orient: 'vertical',
            left: 'left',
            data: data.labels || ['集成测试', '压力测试', '基准测试', '模糊测试']
        },
        series: [
            {
                name: '测试类型',
                type: 'pie',
                radius: ['40%', '70%'],
                avoidLabelOverlap: false,
                itemStyle: {
                    borderRadius: 10,
                    borderColor: '#fff',
                    borderWidth: 2
                },
                label: {
                    show: false,
                    position: 'center'
                },
                emphasis: {
                    label: {
                        show: true,
                        fontSize: '30',
                        fontWeight: 'bold'
                    }
                },
                labelLine: {
                    show: false
                },
                data: (data.labels || []).map((label, index) => ({
                    value: data.data ? data.data[index] : 0,
                    name: label,
                    itemStyle: {
                        color: data.colors ? data.colors[index] : undefined
                    }
                }))
            }
        ]
    };
    
    myChart.setOption(option);
    dashboard.charts.testDistributionChart = myChart;
    
    window.addEventListener('resize', () => {
        myChart.resize();
    });
}

// 初始化SLG战斗性能仪表
function initSLGBattleGauge(data) {
    const chartDom = document.getElementById('slgBattleGauge');
    const myChart = echarts.init(chartDom);
    
    const option = {
        series: [
            {
                type: 'gauge',
                center: ['50%', '60%'],
                startAngle: 200,
                endAngle: -20,
                min: 0,
                max: 100,
                splitNumber: 10,
                itemStyle: {
                    color: '#58D9F9',
                    shadowColor: 'rgba(0,138,255,0.45)',
                    shadowBlur: 10,
                    shadowOffsetX: 2,
                    shadowOffsetY: 2
                },
                progress: {
                    show: true,
                    roundCap: true,
                    width: 18
                },
                pointer: {
                    icon: 'path://M2090.36389,615.30999 L2090.36389,615.30999 C2091.48372,615.30999 2092.40383,616.194028 2092.44859,617.312956 L2096.90698,728.755929 C2097.05155,732.369577 2094.2393,735.416212 2090.62566,735.56078 C2090.53845,735.564269 2090.45117,735.566014 2090.36389,735.566014 L2090.36389,735.566014 C2086.74078,735.566014 2083.81219,732.63742 2083.81219,729.014312 C2083.81219,728.927025 2083.81394,728.839778 2083.81744,728.752531 L2088.27583,617.312956 C2088.32059,616.194028 2089.24069,615.30999 2090.36389,615.30999 Z',
                    length: '75%',
                    width: 16,
                    offsetCenter: [0, '5%']
                },
                axisLine: {
                    roundCap: true,
                    lineStyle: {
                        width: 18
                    }
                },
                axisTick: {
                    splitNumber: 2,
                    lineStyle: {
                        width: 2,
                        color: '#999'
                    }
                },
                splitLine: {
                    length: 12,
                    lineStyle: {
                        width: 3,
                        color: '#999'
                    }
                },
                axisLabel: {
                    distance: 30,
                    color: '#999',
                    fontSize: 20
                },
                title: {
                    show: false
                },
                detail: {
                    backgroundColor: '#fff',
                    borderColor: '#999',
                    borderWidth: 2,
                    width: '60%',
                    lineHeight: 40,
                    height: 40,
                    borderRadius: 8,
                    offsetCenter: [0, '35%'],
                    valueAnimation: true,
                    formatter: function (value) {
                        return '{value|' + value.toFixed(0) + '}{unit|分}';
                    },
                    rich: {
                        value: {
                            fontSize: 50,
                            fontWeight: 'bolder',
                            color: '#777'
                        },
                        unit: {
                            fontSize: 20,
                            color: '#999',
                            padding: [0, 0, -20, 10]
                        }
                    }
                },
                data: [
                    {
                        value: data.value || 87.5
                    }
                ]
            }
        ]
    };
    
    myChart.setOption(option);
    dashboard.charts.slgBattleGauge = myChart;
    
    window.addEventListener('resize', () => {
        myChart.resize();
    });
}

// 启动实时数据更新
function startRealtimeUpdates() {
    console.log('🔄 启动实时数据更新');
    // WebSocket会自动推送数据，这里不需要轮询
}

// 创建新测试
function createNewTest() {
    const modal = new bootstrap.Modal(document.getElementById('newTestModal'));
    modal.show();
}

// 提交新测试
async function submitNewTest() {
    const form = document.getElementById('newTestForm');
    const formData = new FormData(form);
    
    const testData = {
        name: document.getElementById('testName').value,
        type: document.getElementById('testType').value,
        slg_mode: document.getElementById('slgMode').checked,
        unity_url: document.getElementById('unityUrl').value,
        game_url: document.getElementById('gameUrl').value,
        config: {}
    };
    
    // 解析配置JSON
    const configText = document.getElementById('testConfig').value;
    if (configText) {
        try {
            testData.config = JSON.parse(configText);
        } catch (e) {
            alert('配置JSON格式错误：' + e.message);
            return;
        }
    }
    
    try {
        const response = await axios.post('/api/v1/tests', testData, {
            timeout: 15000,
            headers: {
                'Content-Type': 'application/json'
            }
        });

        const result = response.data;

        if (response.status >= 200 && response.status < 300) {
            // 关闭模态框
            bootstrap.Modal.getInstance(document.getElementById('newTestModal')).hide();

            // 显示成功通知
            dashboard.showNotification('测试创建成功', `测试 ${result.test_id} 已创建`, 'success');

            // 询问是否立即启动
            if (confirm('是否立即启动测试？')) {
                startTest(result.test_id);
            }

            // 刷新测试列表
            setTimeout(() => {
                window.location.reload();
            }, 2000);

        } else {
            alert('创建测试失败：' + result.message);
        }
    } catch (error) {
        console.error('创建测试出错:', error);
        const message = error.response?.data?.message || error.message;
        alert('创建测试出错：' + message);
    }
}

// 启动测试
async function startTest(testId) {
    try {
        const response = await axios.post(`/api/v1/tests/${testId}/start`, {}, {
            timeout: 10000
        });

        const result = response.data;

        if (response.status >= 200 && response.status < 300) {
            dashboard.showNotification('测试已启动', `测试 ${testId} 正在运行`, 'success');
        } else {
            alert('启动测试失败：' + result.message);
        }
    } catch (error) {
        console.error('启动测试出错:', error);
        const message = error.response?.data?.message || error.message;
        alert('启动测试出错：' + message);
    }
}

// 查看测试详情
function viewTestDetail(testId) {
    window.open(`/dashboard/test/${testId}`, '_blank');
}

// 全局dashboard实例
let dashboard;

// DOM加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    dashboard = new DashboardManager();
    console.log('🎮 Unity SLG 测试平台 Dashboard 已初始化');

    // 添加axios日志控制
    setupAxiosLoggingControls();
});

// 设置axios日志控制
function setupAxiosLoggingControls() {
    // 添加键盘快捷键 (Ctrl+Shift+L) 来切换日志
    document.addEventListener('keydown', function(event) {
        if (event.ctrlKey && event.shiftKey && event.key === 'L') {
            event.preventDefault();
            const enabled = axiosLogger.toggleLogging();
            dashboard.showNotification(
                'Axios日志',
                `请求/响应日志已${enabled ? '启用' : '禁用'}`,
                'info'
            );
        }
    });

    // 添加日志状态显示
    if (axiosLogger.enableLogging) {
        console.log('🔍 Axios logging is ENABLED. Use Ctrl+Shift+L to toggle.');
        axiosLogger.logConfig();
    }
}