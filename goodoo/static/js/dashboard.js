// Dashboard JavaScript functionality
class GoodooDashboard {
    constructor() {
        this.currentSection = 'overview';
        this.charts = {};
        this.refreshInterval = null;
        this.apiBaseUrl = '/api';
        
        this.init();
    }

    init() {
        this.setupNavigation();
        this.setupCharts();
        this.loadDashboardData();
        this.startAutoRefresh();
        this.setupEventListeners();
    }

    setupNavigation() {
        const navItems = document.querySelectorAll('.nav-item');
        
        navItems.forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const section = item.dataset.section;
                this.switchSection(section);
            });
        });
    }

    switchSection(sectionName) {
        // Update active nav item
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-section="${sectionName}"]`).classList.add('active');

        // Update active section
        document.querySelectorAll('.dashboard-section').forEach(section => {
            section.classList.remove('active');
        });
        document.getElementById(`${sectionName}-section`).classList.add('active');

        // Update header
        this.updateSectionHeader(sectionName);
        
        // Load section-specific data
        this.loadSectionData(sectionName);
        
        this.currentSection = sectionName;
    }

    updateSectionHeader(sectionName) {
        const titles = {
            'overview': {
                title: 'Dashboard Overview',
                description: 'Monitor your Goodoo framework tools and system status'
            },
            'llm-tools': {
                title: 'LLM Tools',
                description: 'Select and configure AI tools for your workflow'
            },
            'chat': {
                title: 'AI Chat',
                description: 'Communicate with your AI assistant'
            },
            'api': {
                title: 'API Performance',
                description: 'Track API metrics and performance indicators'
            },
            'logs': {
                title: 'System Logs',
                description: 'View and filter system logs and events'
            },
            'settings': {
                title: 'System Settings',
                description: 'Configure system parameters and preferences'
            }
        };

        const config = titles[sectionName] || titles['overview'];
        document.getElementById('section-title').textContent = config.title;
        document.getElementById('section-description').textContent = config.description;
    }

    setupCharts() {
        // Request Volume Chart
        const requestCtx = document.getElementById('requestChart');
        if (requestCtx) {
            this.charts.requests = new Chart(requestCtx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [{
                        label: 'Requests/min',
                        data: [],
                        borderColor: '#3b82f6',
                        backgroundColor: 'rgba(59, 130, 246, 0.1)',
                        tension: 0.4,
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true
                        }
                    },
                    plugins: {
                        legend: {
                            display: false
                        }
                    }
                }
            });
        }

        // Response Time Chart
        const responseCtx = document.getElementById('responseChart');
        if (responseCtx) {
            this.charts.response = new Chart(responseCtx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [{
                        label: 'Response Time (ms)',
                        data: [],
                        borderColor: '#10b981',
                        backgroundColor: 'rgba(16, 185, 129, 0.1)',
                        tension: 0.4,
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true
                        }
                    },
                    plugins: {
                        legend: {
                            display: false
                        }
                    }
                }
            });
        }
    }

    async loadDashboardData() {
        // Show loading state
        this.showLoadingState(true);
        
        try {
            // Load dashboard data with timeout and retry logic
            const promises = [
                this.loadToolsOverview(),
                this.loadChartData(),
                this.loadActivityFeed()
            ];
            
            // Add timeout to prevent hanging
            const timeoutPromise = new Promise((_, reject) => 
                setTimeout(() => reject(new Error('Request timeout')), 10000)
            );
            
            await Promise.race([
                Promise.all(promises),
                timeoutPromise
            ]);
            
            // Hide loading state after successful load
            setTimeout(() => this.showLoadingState(false), 500);
            
        } catch (error) {
            console.error('Error loading dashboard data:', error);
            let errorMessage = 'Error loading dashboard data';
            
            if (error.message === 'Request timeout') {
                errorMessage = 'Dashboard loading timed out - check your connection';
            } else if (error.message.includes('Failed to fetch')) {
                errorMessage = 'Unable to connect to server';
            }
            
            this.showNotification(errorMessage, 'error');
            this.showLoadingState(false);
            
            // Attempt to load cached or fallback data
            this.loadFallbackData();
        }
    }

    showLoadingState(show) {
        const metrics = document.querySelectorAll('.metric-value');
        metrics.forEach(metric => {
            if (show) {
                metric.style.opacity = '0.6';
                metric.style.transform = 'scale(0.95)';
            } else {
                metric.style.opacity = '1';
                metric.style.transform = 'scale(1)';
            }
        });
        
        // Add pulse animation to refresh button when loading
        const refreshBtn = document.querySelector('.refresh-btn');
        if (refreshBtn) {
            if (show) {
                refreshBtn.style.animation = 'pulse 1.5s infinite';
            } else {
                refreshBtn.style.animation = '';
            }
        }
    }

    async loadMetrics() {
        try {
            const response = await fetch('/api/metrics');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }
            const data = await response.json();
            
            // Calculate percentage changes (simulated)
            const userChange = data.active_users > 0 ? '+' + (5 + data.active_users % 10) + '%' : '0%';
            const requestChange = '+' + (10 + data.request_count % 15) + '%';
            const responseChange = '-' + (5 + data.avg_response_time % 8) + '%';
            
            // Update metric cards with real data
            this.updateMetric('active-users', data.active_users || 0, userChange);
            this.updateMetric('api-requests', data.request_count || 0, requestChange);
            this.updateMetric('response-time', `${data.avg_response_time || 0}ms`, responseChange);
            this.updateMetric('system-health', data.system_health || 'Healthy');
            
            // Update system health indicator
            const healthStatus = document.getElementById('health-status');
            if (healthStatus) {
                healthStatus.className = `metric-status ${data.status === 'healthy' ? 'online' : 'warning'}`;
            }
            
        } catch (error) {
            console.error('Error loading metrics:', error);
            this.showNotification('Failed to load metrics', 'error');
            // Fallback to mock data
            this.loadMockMetrics();
        }
    }


    async loadChartData() {
        try {
            const response = await fetch('/api/metrics/charts');
            const data = await response.json();
            
            if (this.charts.requests && data.requests) {
                this.charts.requests.data.labels = data.requests.labels;
                this.charts.requests.data.datasets[0].data = data.requests.data;
                this.charts.requests.update();
            }
            
            if (this.charts.response && data.response_times) {
                this.charts.response.data.labels = data.response_times.labels;
                this.charts.response.data.datasets[0].data = data.response_times.data;
                this.charts.response.update();
            }
            
        } catch (error) {
            // Generate mock data if API not available
            this.generateMockChartData();
        }
    }

    generateMockChartData() {
        const now = new Date();
        const labels = [];
        const requestData = [];
        const responseData = [];
        
        for (let i = 23; i >= 0; i--) {
            const time = new Date(now.getTime() - i * 60 * 60 * 1000);
            labels.push(time.getHours().toString().padStart(2, '0') + ':00');
            requestData.push(Math.floor(Math.random() * 100) + 20);
            responseData.push(Math.floor(Math.random() * 200) + 50);
        }
        
        if (this.charts.requests) {
            this.charts.requests.data.labels = labels;
            this.charts.requests.data.datasets[0].data = requestData;
            this.charts.requests.update();
        }
        
        if (this.charts.response) {
            this.charts.response.data.labels = labels;
            this.charts.response.data.datasets[0].data = responseData;
            this.charts.response.update();
        }
    }

    async loadActivityFeed() {
        try {
            const response = await fetch('/api/activity/recent');
            const activities = await response.json();
            
            const activityFeed = document.getElementById('activity-feed');
            if (activityFeed && activities) {
                activityFeed.innerHTML = activities.map(activity => `
                    <div class="activity-item">
                        <span class="activity-time">${this.formatTime(activity.timestamp)}</span>
                        <span class="activity-text">${activity.message}</span>
                        <span class="activity-type ${activity.level.toLowerCase()}">${activity.level}</span>
                    </div>
                `).join('');
            }
        } catch (error) {
            // Keep mock data if API not available
            console.log('Using mock activity data');
        }
    }

    async loadSectionData(sectionName) {
        switch (sectionName) {
            case 'llm-tools':
                await this.loadLLMTools();
                break;
            case 'chat':
                await this.loadChatSection();
                break;
            case 'api':
                await this.loadAPIMetrics();
                break;
            case 'logs':
                await this.loadLogs();
                break;
            case 'settings':
                await this.loadSettings();
                break;
        }
    }

    async loadToolsOverview() {
        // Load active tools information
        try {
            const toolCount = document.getElementById('tool-count');
            const toolItems = document.querySelectorAll('.tool-item');
            
            // Update tool count
            if (toolCount) {
                toolCount.textContent = toolItems.length;
            }
            
            // Check tool status (simulate API check)
            toolItems.forEach((item, index) => {
                const status = item.querySelector('.tool-status');
                const isActive = Math.random() > 0.1; // 90% chance tools are active
                
                if (isActive) {
                    item.classList.add('active');
                    status.textContent = 'Active';
                    status.className = 'tool-status active';
                } else {
                    item.classList.remove('active');
                    status.textContent = 'Inactive';
                    status.className = 'tool-status inactive';
                }
            });
            
        } catch (error) {
            console.error('Error loading tools overview:', error);
        }
    }

    async loadAPIMetrics() {
        try {
            const response = await fetch('/api/metrics/api');
            const data = await response.json();
            
            document.getElementById('total-requests').textContent = data.total_requests || 0;
            document.getElementById('success-rate').textContent = `${data.success_rate || 0}%`;
            document.getElementById('error-rate').textContent = `${data.error_rate || 0}%`;
            document.getElementById('avg-response-time').textContent = `${data.avg_response_time || 0}ms`;
        } catch (error) {
            // Mock data
            document.getElementById('total-requests').textContent = '15,432';
            document.getElementById('success-rate').textContent = '98.7%';
            document.getElementById('error-rate').textContent = '1.3%';
            document.getElementById('avg-response-time').textContent = '127ms';
        }
    }


    async loadLogs() {
        try {
            const response = await fetch('/api/logs/recent');
            const logs = await response.json();
            
            const logsDisplay = document.getElementById('logs-display');
            if (logsDisplay && logs) {
                logsDisplay.innerHTML = logs.map(log => `
                    <div class="log-entry">
                        <span class="log-time">${this.formatDateTime(log.timestamp)}</span>
                        <span class="log-level ${log.level.toLowerCase()}">${log.level.toUpperCase()}</span>
                        <span class="log-message">${log.message}</span>
                    </div>
                `).join('');
            }
        } catch (error) {
            // Keep existing mock log entry
            console.log('Using mock log data');
        }
    }

    async loadSettings() {
        try {
            const response = await fetch('/api/settings');
            const settings = await response.json();
            
            if (settings) {
                document.getElementById('settings-log-level').value = settings.log_level || 'info';
                document.getElementById('session-timeout').value = settings.session_timeout || 60;
                document.getElementById('performance-monitoring').checked = settings.performance_monitoring !== false;
            }
        } catch (error) {
            console.log('Using default settings');
        }
    }

    setupEventListeners() {
        // Time range filter
        const timeRange = document.getElementById('timeRange');
        if (timeRange) {
            timeRange.addEventListener('change', () => {
                this.updateTimeRange();
            });
        }

        // Log level filter
        const logLevel = document.getElementById('log-level');
        if (logLevel) {
            logLevel.addEventListener('change', () => {
                this.filterLogs();
            });
        }

        // Refresh button
        const refreshBtn = document.querySelector('.refresh-btn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.refreshData();
            });
        }

        // LLM Tools toggle switches
        const toolToggles = document.querySelectorAll('.tool-toggle');
        toolToggles.forEach(toggle => {
            toggle.addEventListener('change', (e) => {
                this.handleToolToggle(e.target);
            });
        });
    }

    startAutoRefresh() {
        // Refresh data every 30 seconds
        this.refreshInterval = setInterval(() => {
            this.loadDashboardData();
        }, 30000);
    }

    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    refreshData() {
        const refreshBtn = document.querySelector('.refresh-btn');
        const originalText = refreshBtn.innerHTML;
        refreshBtn.innerHTML = '<span class="refresh-icon">⏳</span> Refreshing...';
        refreshBtn.disabled = true;

        this.loadDashboardData().finally(() => {
            refreshBtn.innerHTML = originalText;
            refreshBtn.disabled = false;
        });
    }

    updateTimeRange() {
        const timeRange = document.getElementById('timeRange').value;
        console.log('Time range changed to:', timeRange);
        // Reload chart data with new time range
        this.loadChartData();
    }

    filterLogs() {
        const selectedLevel = document.getElementById('log-level').value;
        const logEntries = document.querySelectorAll('.log-entry');
        
        logEntries.forEach(entry => {
            const logLevel = entry.querySelector('.log-level').textContent.toLowerCase();
            if (selectedLevel === 'all' || logLevel === selectedLevel) {
                entry.style.display = 'flex';
            } else {
                entry.style.display = 'none';
            }
        });
    }

    saveSettings() {
        const settings = {
            log_level: document.getElementById('settings-log-level').value,
            session_timeout: parseInt(document.getElementById('session-timeout').value),
            performance_monitoring: document.getElementById('performance-monitoring').checked
        };

        fetch('/api/settings', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(settings)
        })
        .then(response => response.json())
        .then(data => {
            this.showNotification('Settings saved successfully', 'success');
        })
        .catch(error => {
            console.error('Error saving settings:', error);
            this.showNotification('Error saving settings', 'error');
        });
    }


    async loadLLMTools() {
        try {
            // Load real LLM data from backend
            const response = await fetch('/api/llm/tools');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }
            const data = await response.json();
            
            // Update tool cards based on real addon status
            this.updateToolCardsFromBackend(data.addons);
            
            // Update summary statistics
            this.updateLLMSummary(data.summary);
            
            // Load saved tool configuration from localStorage
            this.loadToolConfiguration();
            this.updateActiveToolsList();
            
        } catch (error) {
            console.error('Error loading LLM tools:', error);
            this.showNotification('Failed to load LLM tools data', 'error');
            // Fall back to existing functionality
            this.loadToolConfiguration();
            this.updateActiveToolsList();
        }
    }

    loadToolConfiguration() {
        // Load from localStorage or use defaults
        const savedConfig = localStorage.getItem('llm-tools-config');
        if (savedConfig) {
            try {
                const config = JSON.parse(savedConfig);
                Object.keys(config).forEach(toolId => {
                    const toggle = document.querySelector(`.tool-toggle[data-tool="${toolId}"]`);
                    if (toggle) {
                        toggle.checked = config[toolId].enabled;
                        this.handleToolToggle(toggle, false); // Don't update list yet
                        
                        // Load config values
                        const toolCard = toggle.closest('.llm-tool-card');
                        const configInputs = toolCard.querySelectorAll('.config-input');
                        configInputs.forEach((input, index) => {
                            if (config[toolId].config && config[toolId].config[index] !== undefined) {
                                input.value = config[toolId].config[index];
                            }
                        });
                    }
                });
            } catch (error) {
                console.error('Error loading tool configuration:', error);
            }
        }
    }

    handleToolToggle(toggle, updateList = true) {
        const toolCard = toggle.closest('.llm-tool-card');
        const configSection = toolCard.querySelector('.tool-config');
        
        if (toggle.checked) {
            toolCard.classList.add('enabled');
            configSection.style.display = 'block';
        } else {
            toolCard.classList.remove('enabled');
            configSection.style.display = 'none';
        }
        
        if (updateList) {
            this.updateActiveToolsList();
        }
    }

    updateActiveToolsList() {
        const activeToolsList = document.getElementById('active-tools-list');
        const enabledToggles = document.querySelectorAll('.tool-toggle:checked');
        
        if (enabledToggles.length === 0) {
            activeToolsList.innerHTML = '<p class="no-tools">No tools selected. Choose tools from the categories above.</p>';
            return;
        }
        
        const toolsHtml = Array.from(enabledToggles).map(toggle => {
            const toolCard = toggle.closest('.llm-tool-card');
            const toolName = toolCard.querySelector('h5').textContent;
            const toolProvider = toolCard.querySelector('.tool-provider').textContent;
            
            return `
                <div class="active-tool-item">
                    <span class="active-tool-name">${toolName}</span>
                    <span class="active-tool-provider">${toolProvider}</span>
                    <span class="active-tool-status">🟢 Active</span>
                </div>
            `;
        }).join('');
        
        activeToolsList.innerHTML = toolsHtml;
    }

    async saveToolConfiguration() {
        const config = {};
        const toolToggles = document.querySelectorAll('.tool-toggle');
        
        toolToggles.forEach(toggle => {
            const toolId = toggle.dataset.tool;
            const isEnabled = toggle.checked;
            const toolCard = toggle.closest('.llm-tool-card');
            const configInputs = toolCard.querySelectorAll('.config-input');
            
            config[toolId] = {
                enabled: isEnabled,
                config: Array.from(configInputs).map(input => input.value)
            };
        });
        
        try {
            // Save to localStorage
            localStorage.setItem('llm-tools-config', JSON.stringify(config));
            
            // Also save to backend if possible
            try {
                const response = await fetch('/api/llm/config', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        config: config
                    })
                });
                
                if (response.ok) {
                    this.showNotification('Tool configuration saved successfully!', 'success');
                } else {
                    this.showNotification('Configuration saved locally (server sync failed)', 'warning');
                }
            } catch (serverError) {
                console.warn('Failed to sync config to server:', serverError);
                this.showNotification('Configuration saved locally (server sync failed)', 'warning');
            }
            
            // Update the tools overview in the main dashboard
            this.updateToolsOverview();
            
        } catch (error) {
            console.error('Error saving tool configuration:', error);
            this.showNotification('Failed to save configuration', 'error');
        }
    }

    async testSelectedTools() {
        const enabledToggles = document.querySelectorAll('.tool-toggle:checked');
        
        if (enabledToggles.length === 0) {
            this.showNotification('No tools selected to test', 'warning');
            return;
        }
        
        this.showNotification('Testing selected tools...', 'info');
        
        try {
            // Get provider data to test actual connections
            const providersResponse = await fetch('/api/llm/providers');
            const providers = await providersResponse.json();
            
            const testResults = [];
            
            for (const toggle of enabledToggles) {
                const toolId = toggle.dataset.tool;
                const toolCard = toggle.closest('.llm-tool-card');
                const toolName = toolCard.querySelector('h5').textContent;
                
                // Find corresponding provider for testing
                const provider = providers.find(p => 
                    toolId.includes(p.service) || 
                    p.name.toLowerCase().includes(toolId.replace('llm_', ''))
                );
                
                if (provider && provider.active) {
                    try {
                        const testResponse = await fetch('/api/llm/test', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json'
                            },
                            body: JSON.stringify({
                                provider_id: provider.id
                            })
                        });
                        
                        const testResult = await testResponse.json();
                        testResults.push({
                            name: toolName,
                            success: testResult.success,
                            responseTime: testResult.response_time_ms,
                            error: testResult.error
                        });
                    } catch (error) {
                        testResults.push({
                            name: toolName,
                            success: false,
                            error: 'Connection failed'
                        });
                    }
                } else {
                    testResults.push({
                        name: toolName,
                        success: false,
                        error: 'Provider not active or configured'
                    });
                }
            }
            
            const successCount = testResults.filter(r => r.success).length;
            const avgResponseTime = testResults
                .filter(r => r.success && r.responseTime)
                .reduce((sum, r) => sum + r.responseTime, 0) / 
                testResults.filter(r => r.success && r.responseTime).length;
            
            let message = `Testing completed: ${successCount}/${testResults.length} tools working correctly`;
            if (successCount > 0 && avgResponseTime) {
                message += ` (avg response: ${Math.round(avgResponseTime)}ms)`;
            }
            
            const type = successCount === testResults.length ? 'success' : 'warning';
            this.showNotification(message, type);
            
        } catch (error) {
            console.error('Error testing tools:', error);
            this.showNotification('Failed to test tools', 'error');
        }
    }

    resetToolConfiguration() {
        if (confirm('Are you sure you want to reset all tool configurations to defaults?')) {
            localStorage.removeItem('llm-tools-config');
            
            // Reset all toggles
            const toolToggles = document.querySelectorAll('.tool-toggle');
            toolToggles.forEach(toggle => {
                toggle.checked = false;
                this.handleToolToggle(toggle, false);
            });
            
            // Reset all config inputs
            const configInputs = document.querySelectorAll('.config-input');
            configInputs.forEach(input => {
                if (input.type === 'number') {
                    input.value = input.defaultValue || input.getAttribute('value') || '';
                } else {
                    input.value = input.defaultValue || input.getAttribute('value') || '';
                }
            });
            
            this.updateActiveToolsList();
            this.showNotification('Configuration reset to defaults', 'success');
        }
    }

    updateToolsOverview() {
        // Update the main overview section with current tool count
        const enabledToggles = document.querySelectorAll('.tool-toggle:checked');
        const toolCountElement = document.getElementById('tool-count');
        
        if (toolCountElement) {
            // Add 4 to account for the core dashboard tools (API Metrics, Log Monitor, Settings Manager, Dashboard Core)
            const totalTools = enabledToggles.length + 4;
            toolCountElement.textContent = totalTools;
        }
    }

    updateToolCardsFromBackend(addons) {
        // Update tool card status based on real addon data
        addons.forEach(addon => {
            const toolCard = document.querySelector(`[data-tool="${addon.name}"]`);
            if (toolCard) {
                const card = toolCard.closest('.llm-tool-card');
                const toolInfo = card.querySelector('.tool-info h5');
                const toolProvider = card.querySelector('.tool-provider');
                const toolDescription = card.querySelector('.tool-description');
                
                // Update display name and status
                if (toolInfo) toolInfo.textContent = addon.display_name;
                if (toolProvider) toolProvider.textContent = addon.name;
                if (toolDescription) {
                    let description = addon.description || toolDescription.textContent;
                    if (!addon.installed) {
                        description += ' (Not Installed)';
                    } else if (!addon.active) {
                        description += ' (Inactive)';
                    }
                    toolDescription.textContent = description;
                }
                
                // Update card styling based on status
                if (!addon.installed) {
                    card.classList.add('not-installed');
                    card.style.opacity = '0.6';
                    toolCard.disabled = true;
                } else if (!addon.active) {
                    card.classList.add('inactive');
                    card.style.opacity = '0.8';
                } else {
                    card.classList.remove('not-installed', 'inactive');
                    card.style.opacity = '1';
                    toolCard.disabled = false;
                }
            }
        });
    }

    updateLLMSummary(summary) {
        // Update the overview section with real LLM statistics
        const toolCountElement = document.getElementById('tool-count');
        if (toolCountElement && summary) {
            // Show actual installed addons + core dashboard tools
            const totalTools = summary.installed_addons + 4;
            toolCountElement.textContent = totalTools;
        }
        
        // Update tools list if we're on the overview section
        if (this.currentSection === 'overview') {
            this.updateToolsListFromSummary(summary);
        }
    }

    updateToolsListFromSummary(summary) {
        // Update the tools list in overview to show LLM tool status
        const toolsList = document.querySelector('.tools-list');
        if (toolsList && summary) {
            // Add LLM-specific tools to the existing list
            const llmTools = [
                {
                    icon: '🤖',
                    name: 'LLM Providers',
                    status: `${summary.active_providers}/${summary.total_providers} Active`
                },
                {
                    icon: '🧠',
                    name: 'AI Models',
                    status: `${summary.active_models}/${summary.total_models} Available`
                },
                {
                    icon: '📦',
                    name: 'LLM Addons',
                    status: `${summary.installed_addons} Installed`
                }
            ];
            
            llmTools.forEach(tool => {
                const existingTool = Array.from(toolsList.children).find(item => 
                    item.querySelector('.tool-name').textContent === tool.name
                );
                
                if (existingTool) {
                    // Update existing tool status
                    const statusElement = existingTool.querySelector('.tool-status');
                    if (statusElement) {
                        statusElement.textContent = tool.status;
                    }
                } else {
                    // Add new tool if it doesn't exist
                    const toolItem = document.createElement('div');
                    toolItem.className = 'tool-item active';
                    toolItem.innerHTML = `
                        <span class="tool-icon">${tool.icon}</span>
                        <span class="tool-name">${tool.name}</span>
                        <span class="tool-status">${tool.status}</span>
                    `;
                    toolsList.appendChild(toolItem);
                }
            });
        }
    }

    showNotification(message, type = 'info') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.textContent = message;
        
        // Style the notification
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 12px 20px;
            border-radius: 4px;
            color: white;
            z-index: 10000;
            transition: all 0.3s ease;
            max-width: 300px;
        `;
        
        // Set background color based on type
        switch (type) {
            case 'success':
                notification.style.backgroundColor = '#10b981';
                break;
            case 'error':
                notification.style.backgroundColor = '#ef4444';
                break;
            case 'warning':
                notification.style.backgroundColor = '#f59e0b';
                break;
            default:
                notification.style.backgroundColor = '#3b82f6';
        }
        
        document.body.appendChild(notification);
        
        // Remove after 3 seconds
        setTimeout(() => {
            notification.style.opacity = '0';
            setTimeout(() => {
                document.body.removeChild(notification);
            }, 300);
        }, 3000);
    }

    formatTime(timestamp) {
        const date = new Date(timestamp);
        const now = new Date();
        const diff = now - date;
        
        if (diff < 60000) {
            return 'Just now';
        } else if (diff < 3600000) {
            return `${Math.floor(diff / 60000)} min ago`;
        } else if (diff < 86400000) {
            return `${Math.floor(diff / 3600000)} hours ago`;
        } else {
            return date.toLocaleDateString();
        }
    }

    formatDate(timestamp) {
        return new Date(timestamp).toLocaleString();
    }

    formatDateTime(timestamp) {
        return new Date(timestamp).toISOString().slice(0, 19).replace('T', ' ');
    }

    loadFallbackData() {
        // Load basic fallback data when API calls fail
        try {
            // Check localStorage for cached data
            const cachedConfig = localStorage.getItem('llm-tools-config');
            if (cachedConfig) {
                this.loadToolConfiguration();
                this.showNotification('Loaded cached configuration', 'info');
            }
            
            // Generate minimal mock chart data
            this.generateMockChartData();
            
            // Show offline status
            this.showOfflineIndicator(true);
            
        } catch (error) {
            console.error('Error loading fallback data:', error);
        }
    }

    showOfflineIndicator(show) {
        let indicator = document.getElementById('offline-indicator');
        
        if (show && !indicator) {
            // Create offline indicator
            indicator = document.createElement('div');
            indicator.id = 'offline-indicator';
            indicator.className = 'offline-indicator';
            indicator.innerHTML = `
                <span class="offline-icon">⚠️</span>
                <span class="offline-text">Working Offline</span>
                <button class="retry-btn" onclick="dashboard.retryConnection()">Retry</button>
            `;
            indicator.style.cssText = `
                position: fixed;
                top: 10px;
                right: 10px;
                background: #fbbf24;
                color: #92400e;
                padding: 8px 12px;
                border-radius: 4px;
                font-size: 0.875rem;
                z-index: 10001;
                box-shadow: 0 2px 8px rgba(0,0,0,0.1);
                display: flex;
                align-items: center;
                gap: 8px;
            `;
            document.body.appendChild(indicator);
        } else if (!show && indicator) {
            indicator.remove();
        }
    }

    retryConnection() {
        this.showOfflineIndicator(false);
        this.showNotification('Reconnecting...', 'info');
        this.loadDashboardData();
    }

    // Connection monitoring
    startConnectionMonitoring() {
        // Check connection status periodically
        setInterval(async () => {
            try {
                const response = await fetch('/api/metrics', { 
                    method: 'HEAD',
                    cache: 'no-cache'
                });
                if (response.ok) {
                    this.showOfflineIndicator(false);
                }
            } catch (error) {
                this.showOfflineIndicator(true);
            }
        }, 30000); // Check every 30 seconds
    }

    // Chat functionality
    async loadChatSection() {
        try {
            // Initialize chat interface
            this.initializeChatInterface();
            
            // Load available models
            await this.loadChatModels();
            
            // Setup chat event listeners
            this.setupChatEventListeners();
            
            // Initialize user chat functionality
            this.initializeUserChat();
            
        } catch (error) {
            console.error('Error loading chat section:', error);
            this.showNotification('Failed to load chat interface', 'error');
        }
    }

    initializeChatInterface() {
        // Initialize chat state
        this.currentChatSession = null;
        this.chatMessages = [];
        this.isTyping = false;
        
        // Hide suggestions initially if there are messages
        const messagesContainer = document.getElementById('chat-messages');
        const suggestions = document.getElementById('chat-suggestions');
        
        if (messagesContainer && suggestions) {
            const hasMessages = messagesContainer.children.length > 1; // more than welcome message
            suggestions.style.display = hasMessages ? 'none' : 'block';
        }
    }

    async loadChatModels() {
        try {
            const response = await fetch('/api/chat/models');
            const data = await response.json();
            
            const modelSelect = document.getElementById('ai-model-select');
            if (modelSelect && data.models) {
                modelSelect.innerHTML = '';
                data.models.forEach(model => {
                    const option = document.createElement('option');
                    option.value = model.id;
                    option.textContent = `${model.name} (${model.provider})`;
                    option.disabled = !model.available;
                    modelSelect.appendChild(option);
                });
                
                // Set default model
                if (data.default) {
                    modelSelect.value = data.default;
                }
            }
        } catch (error) {
            console.error('Error loading chat models:', error);
        }
    }

    setupChatEventListeners() {
        const chatInput = document.getElementById('chat-input');
        const sendBtn = document.getElementById('send-btn');
        const charCount = document.getElementById('char-count');

        if (chatInput) {
            // Auto-resize textarea
            chatInput.addEventListener('input', (e) => {
                this.updateCharCount();
                this.autoResizeTextarea(e.target);
            });

            // Handle Enter key
            chatInput.addEventListener('keydown', (e) => {
                if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
                    e.preventDefault();
                    this.sendMessage();
                }
            });
        }

        if (sendBtn) {
            sendBtn.addEventListener('click', () => this.sendMessage());
        }
    }

    updateCharCount() {
        const chatInput = document.getElementById('chat-input');
        const charCount = document.getElementById('char-count');
        
        if (chatInput && charCount) {
            const length = chatInput.value.length;
            const maxLength = 4000;
            charCount.textContent = `${length}/${maxLength}`;
            
            if (length > maxLength * 0.9) {
                charCount.style.color = '#ef4444';
            } else if (length > maxLength * 0.7) {
                charCount.style.color = '#f59e0b';
            } else {
                charCount.style.color = '#6b7280';
            }
        }
    }

    autoResizeTextarea(textarea) {
        textarea.style.height = 'auto';
        const newHeight = Math.min(textarea.scrollHeight, 120); // max 120px
        textarea.style.height = newHeight + 'px';
    }

    async sendMessage() {
        const chatInput = document.getElementById('chat-input');
        const sendBtn = document.getElementById('send-btn');
        const modelSelect = document.getElementById('ai-model-select');
        
        if (!chatInput || !chatInput.value.trim()) {
            return;
        }

        const message = chatInput.value.trim();
        const model = modelSelect ? modelSelect.value : 'gpt-3.5-turbo';

        // Disable input during sending
        chatInput.disabled = true;
        sendBtn.disabled = true;
        sendBtn.innerHTML = '<span class="send-icon">⏳</span>';

        try {
            // Add user message to chat
            this.addMessageToChat('user', message);
            
            // Clear input and hide suggestions
            chatInput.value = '';
            this.updateCharCount();
            this.autoResizeTextarea(chatInput);
            document.getElementById('chat-suggestions').style.display = 'none';
            
            // Show typing indicator
            this.showTypingIndicator(true);

            // Send to backend
            const response = await fetch('/api/chat/send', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    message: message,
                    model: model,
                    session_id: this.currentChatSession
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }

            const data = await response.json();
            
            // Store session ID for future messages
            this.currentChatSession = data.session_id;
            
            // Add AI response to chat
            this.addMessageToChat('assistant', data.message, {
                model: data.model,
                responseTime: data.response_time_ms,
                tokensUsed: data.tokens_used
            });

        } catch (error) {
            console.error('Error sending message:', error);
            this.addMessageToChat('system', 'Sorry, I encountered an error processing your message. Please try again.', {
                error: true
            });
        } finally {
            // Re-enable input
            chatInput.disabled = false;
            sendBtn.disabled = false;
            sendBtn.innerHTML = '<span class="send-icon">➤</span>';
            chatInput.focus();
            this.showTypingIndicator(false);
        }
    }

    addMessageToChat(role, content, metadata = {}) {
        const messagesContainer = document.getElementById('chat-messages');
        if (!messagesContainer) return;

        const messageDiv = document.createElement('div');
        messageDiv.className = `chat-message ${role}`;
        
        const timestamp = new Date();
        const timeString = timestamp.toLocaleTimeString([], { 
            hour: '2-digit', 
            minute: '2-digit' 
        });

        let avatarHtml = '';
        if (role === 'user') {
            avatarHtml = '<div class="user-avatar">👤</div>';
        } else if (role === 'assistant') {
            avatarHtml = '<div class="assistant-avatar">🤖</div>';
        } else if (role === 'system') {
            avatarHtml = '<div class="system-avatar">⚠️</div>';
        }

        let metadataHtml = '';
        if (metadata.model && metadata.responseTime) {
            metadataHtml = `<div class="message-metadata">
                <span>${metadata.model}</span>
                <span>${metadata.responseTime}ms</span>
                ${metadata.tokensUsed ? `<span>${metadata.tokensUsed} tokens</span>` : ''}
            </div>`;
        }

        messageDiv.innerHTML = `
            ${avatarHtml}
            <div class="message-content">
                <div class="message-bubble ${role} ${metadata.error ? 'error' : ''}">
                    ${this.formatMessageContent(content)}
                </div>
                <div class="message-time">${timeString}</div>
                ${metadataHtml}
            </div>
        `;

        messagesContainer.appendChild(messageDiv);
        this.scrollToBottom();
    }

    formatMessageContent(content) {
        // Convert markdown-like formatting to HTML
        content = content.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
        content = content.replace(/\*(.*?)\*/g, '<em>$1</em>');
        content = content.replace(/`(.*?)`/g, '<code>$1</code>');
        
        // Convert bullet points
        content = content.replace(/^• (.+)$/gm, '<li>$1</li>');
        content = content.replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>');
        
        // Convert line breaks
        content = content.replace(/\n/g, '<br>');
        
        return content;
    }

    showTypingIndicator(show) {
        const indicator = document.getElementById('typing-indicator');
        if (indicator) {
            indicator.style.display = show ? 'inline' : 'none';
        }

        if (show) {
            const messagesContainer = document.getElementById('chat-messages');
            const typingDiv = document.createElement('div');
            typingDiv.id = 'typing-message';
            typingDiv.className = 'chat-message assistant typing';
            typingDiv.innerHTML = `
                <div class="assistant-avatar">🤖</div>
                <div class="message-content">
                    <div class="typing-dots">
                        <span></span>
                        <span></span>
                        <span></span>
                    </div>
                </div>
            `;
            messagesContainer.appendChild(typingDiv);
            this.scrollToBottom();
        } else {
            const typingMessage = document.getElementById('typing-message');
            if (typingMessage) {
                typingMessage.remove();
            }
        }
    }

    scrollToBottom() {
        const messagesContainer = document.getElementById('chat-messages');
        if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
    }

    useSuggestion(suggestion) {
        const chatInput = document.getElementById('chat-input');
        if (chatInput) {
            chatInput.value = suggestion;
            chatInput.focus();
            this.updateCharCount();
            this.autoResizeTextarea(chatInput);
        }
    }

    clearChat() {
        if (confirm('Are you sure you want to clear the chat history?')) {
            const messagesContainer = document.getElementById('chat-messages');
            if (messagesContainer) {
                // Keep only the welcome message
                const welcomeMessage = messagesContainer.querySelector('.welcome-message');
                messagesContainer.innerHTML = '';
                if (welcomeMessage) {
                    messagesContainer.appendChild(welcomeMessage);
                }
            }
            
            // Show suggestions again
            document.getElementById('chat-suggestions').style.display = 'block';
            
            // Reset session
            this.currentChatSession = null;
            
            this.showNotification('Chat cleared', 'info');
        }
    }

    downloadChat() {
        const messagesContainer = document.getElementById('chat-messages');
        if (!messagesContainer) return;

        const messages = Array.from(messagesContainer.querySelectorAll('.chat-message')).map(msg => {
            const role = msg.classList.contains('user') ? 'User' : 
                        msg.classList.contains('assistant') ? 'Assistant' : 'System';
            const content = msg.querySelector('.message-bubble').textContent.trim();
            const time = msg.querySelector('.message-time')?.textContent || '';
            
            return `[${time}] ${role}: ${content}`;
        }).join('\n\n');

        const blob = new Blob([messages], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `chat-export-${new Date().toISOString().split('T')[0]}.txt`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        this.showNotification('Chat exported successfully', 'success');
    }

    // User-to-User Chat functionality
    async initializeUserChat() {
        try {
            this.currentUserChatMode = 'ai-chat'; // Default to AI chat
            this.currentUserChatRoom = null;
            this.userChatRooms = [];
            this.onlineUsers = [];
            
            // Setup user chat event listeners
            this.setupUserChatEventListeners();
            
            // Load initial data when user chat tab is first activated
            console.log('User chat initialized');
        } catch (error) {
            console.error('Error initializing user chat:', error);
        }
    }

    setupUserChatEventListeners() {
        // Chat tab switching
        const chatTabs = document.querySelectorAll('.chat-tab');
        chatTabs.forEach(tab => {
            tab.addEventListener('click', (e) => {
                const tabType = tab.dataset.tab;
                this.switchChatTab(tabType);
            });
        });
    }

    async switchChatTab(tabType) {
        // Update active tab
        document.querySelectorAll('.chat-tab').forEach(tab => {
            tab.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tabType}"]`).classList.add('active');
        
        // Update tab content
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.remove('active');
        });
        document.getElementById(`${tabType}-info`).classList.add('active');
        
        // Update chat mode
        this.currentUserChatMode = tabType;
        
        if (tabType === 'user-chat') {
            // Load user chat data
            await this.loadUserChatData();
            this.updateChatTitle('Team Chat');
        } else {
            // Switch back to AI chat
            this.updateChatTitle('AI Assistant');
            // Clear any user chat messages and show AI welcome
            this.resetToAIChat();
        }
    }

    async loadUserChatData() {
        try {
            // Load chat rooms
            await this.loadUserChatRooms();
            
            // Load online users
            await this.loadOnlineUsers();
            
            console.log('User chat data loaded');
        } catch (error) {
            console.error('Error loading user chat data:', error);
            this.showNotification('Failed to load user chat data', 'error');
        }
    }

    async loadUserChatRooms() {
        try {
            const response = await this.fetchWithErrorHandling('/api/user-chat/rooms');
            const data = await response.json();
            
            this.userChatRooms = data.rooms || [];
            this.renderChatRooms();
        } catch (error) {
            console.error('Error loading chat rooms:', error);
            this.renderChatRoomsError();
        }
    }

    renderChatRooms() {
        const roomsList = document.getElementById('chat-rooms-list');
        if (!roomsList) return;

        if (this.userChatRooms.length === 0) {
            roomsList.innerHTML = '<div class="loading-rooms">No chat rooms available</div>';
            return;
        }

        roomsList.innerHTML = this.userChatRooms.map(room => `
            <div class="chat-room-item ${room.id === this.currentUserChatRoom?.id ? 'active' : ''}" 
                 data-room-id="${room.id}" onclick="dashboard.selectChatRoom('${room.id}')">
                <div class="room-avatar">
                    ${room.type === 'group' ? '👥' : '👤'}
                </div>
                <div class="room-info">
                    <div class="room-name">${room.name}</div>
                    <div class="room-last-message">
                        ${room.last_message ? room.last_message.content : 'No messages yet'}
                    </div>
                </div>
                <div class="room-meta">
                    <div class="room-time">
                        ${room.last_message ? this.formatRelativeTime(new Date(room.updated_at)) : ''}
                    </div>
                    ${room.unread_count > 0 ? `<div class="room-unread">${room.unread_count}</div>` : ''}
                </div>
            </div>
        `).join('');
    }

    renderChatRoomsError() {
        const roomsList = document.getElementById('chat-rooms-list');
        if (roomsList) {
            roomsList.innerHTML = '<div class="loading-rooms">Failed to load chat rooms</div>';
        }
    }

    async loadOnlineUsers() {
        try {
            const response = await this.fetchWithErrorHandling('/api/user-chat/users');
            const data = await response.json();
            
            this.onlineUsers = data.users || [];
            this.renderOnlineUsers();
        } catch (error) {
            console.error('Error loading online users:', error);
            this.renderOnlineUsersError();
        }
    }

    renderOnlineUsers() {
        const usersList = document.getElementById('online-users-list');
        if (!usersList) return;

        if (this.onlineUsers.length === 0) {
            usersList.innerHTML = '<div class="loading-users">No other users online</div>';
            return;
        }

        usersList.innerHTML = this.onlineUsers.map(user => `
            <div class="user-item" onclick="dashboard.startDirectChat(${user.user_id})">
                <div class="user-avatar-small">
                    ${user.user_name.charAt(0).toUpperCase()}
                    ${user.is_online ? '<div class="online-indicator"></div>' : ''}
                </div>
                <div class="user-info-small">
                    <div class="user-name-small">${user.user_name}</div>
                    <div class="user-status">
                        ${user.is_online ? 'Online' : `Last seen ${this.formatRelativeTime(new Date(user.last_seen))}`}
                    </div>
                </div>
            </div>
        `).join('');
    }

    renderOnlineUsersError() {
        const usersList = document.getElementById('online-users-list');
        if (usersList) {
            usersList.innerHTML = '<div class="loading-users">Failed to load users</div>';
        }
    }

    async selectChatRoom(roomId) {
        try {
            // Find the room
            const room = this.userChatRooms.find(r => r.id === roomId);
            if (!room) return;

            // Update current room
            this.currentUserChatRoom = room;
            
            // Update UI
            this.updateChatTitle(room.name);
            this.renderChatRooms(); // Re-render to update active state
            
            // Load room messages
            await this.loadRoomMessages(roomId);
            
        } catch (error) {
            console.error('Error selecting chat room:', error);
            this.showNotification('Failed to load chat room', 'error');
        }
    }

    async loadRoomMessages(roomId) {
        try {
            const response = await this.fetchWithErrorHandling(`/api/user-chat/room/${roomId}/messages`);
            const data = await response.json();
            
            // Clear current messages
            this.clearChatMessages();
            
            // Add room messages
            if (data.messages && data.messages.length > 0) {
                data.messages.forEach(message => {
                    this.addUserMessageToChat(message);
                });
            } else {
                this.addSystemMessage('This is the beginning of your conversation.');
            }
            
        } catch (error) {
            console.error('Error loading room messages:', error);
            this.addSystemMessage('Failed to load message history.');
        }
    }

    addUserMessageToChat(message) {
        const messagesContainer = document.getElementById('chat-messages');
        if (!messagesContainer) return;

        const messageDiv = document.createElement('div');
        messageDiv.className = `chat-message user`;
        
        const isCurrentUser = true; // In real implementation, check against current user ID
        messageDiv.innerHTML = `
            <div class="${isCurrentUser ? 'user-avatar' : 'assistant-avatar'}">
                ${isCurrentUser ? '👤' : message.from_user_name?.charAt(0) || '👤'}
            </div>
            <div class="message-content">
                <div class="message-bubble ${isCurrentUser ? 'user' : 'assistant'}">
                    <p>${this.escapeHtml(message.content)}</p>
                </div>
                <div class="message-time">${this.formatRelativeTime(new Date(message.timestamp))}</div>
            </div>
        `;

        messagesContainer.appendChild(messageDiv);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    addSystemMessage(content) {
        const messagesContainer = document.getElementById('chat-messages');
        if (!messagesContainer) return;

        const messageDiv = document.createElement('div');
        messageDiv.className = 'chat-message system';
        messageDiv.innerHTML = `
            <div class="system-message">
                <p>${content}</p>
            </div>
        `;

        messagesContainer.appendChild(messageDiv);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    async startDirectChat(userId) {
        try {
            // Find or create direct chat room
            const directRoomId = `direct_${Math.min(1, userId)}_${Math.max(1, userId)}`; // Mock current user ID as 1
            
            // Check if room already exists
            let room = this.userChatRooms.find(r => r.id === directRoomId);
            
            if (!room) {
                // Create new direct chat room (in real implementation, call API)
                const user = this.onlineUsers.find(u => u.user_id === userId);
                room = {
                    id: directRoomId,
                    name: user.user_name,
                    type: 'direct',
                    participants: [user],
                    unread_count: 0,
                    updated_at: new Date().toISOString()
                };
                
                this.userChatRooms.unshift(room);
                this.renderChatRooms();
            }
            
            // Select the room
            await this.selectChatRoom(room.id);
            
        } catch (error) {
            console.error('Error starting direct chat:', error);
            this.showNotification('Failed to start direct chat', 'error');
        }
    }

    updateChatTitle(title) {
        const titleElement = document.getElementById('current-chat-title');
        if (titleElement) {
            titleElement.textContent = title;
        }
    }

    clearChatMessages() {
        const messagesContainer = document.getElementById('chat-messages');
        if (messagesContainer) {
            messagesContainer.innerHTML = '';
        }
    }

    resetToAIChat() {
        this.clearChatMessages();
        // Re-add the welcome message
        const messagesContainer = document.getElementById('chat-messages');
        if (messagesContainer) {
            messagesContainer.innerHTML = `
                <div class="welcome-message">
                    <div class="assistant-avatar">🤖</div>
                    <div class="message-content">
                        <div class="message-bubble assistant">
                            <p>Hello! I'm your AI assistant. I can help you with questions about your Goodoo system, LLM tools, programming, and general assistance.</p>
                            <p>How can I help you today?</p>
                        </div>
                        <div class="message-time">Just now</div>
                    </div>
                </div>
            `;
        }
        
        // Show suggestions again
        const suggestions = document.getElementById('chat-suggestions');
        if (suggestions) {
            suggestions.style.display = 'block';
        }
    }

    async createNewGroupChat() {
        const groupName = prompt('Enter a name for the new group chat:');
        if (!groupName || !groupName.trim()) return;

        try {
            const response = await this.fetchWithErrorHandling('/api/user-chat/room/create', {
                method: 'POST',
                body: JSON.stringify({
                    name: groupName.trim(),
                    participant_ids: [] // In real implementation, show user selection
                })
            });

            const data = await response.json();
            if (data.success) {
                this.showNotification('Group chat created successfully', 'success');
                await this.loadUserChatRooms(); // Reload rooms
            } else {
                throw new Error(data.error || 'Failed to create group chat');
            }
        } catch (error) {
            console.error('Error creating group chat:', error);
            this.showNotification('Failed to create group chat', 'error');
        }
    }

    formatRelativeTime(date) {
        const now = new Date();
        const diffMs = now - date;
        const diffSecs = Math.floor(diffMs / 1000);
        const diffMins = Math.floor(diffSecs / 60);
        const diffHours = Math.floor(diffMins / 60);
        const diffDays = Math.floor(diffHours / 24);

        if (diffSecs < 60) return 'Just now';
        if (diffMins < 60) return `${diffMins}m ago`;
        if (diffHours < 24) return `${diffHours}h ago`;
        if (diffDays < 7) return `${diffDays}d ago`;
        
        return date.toLocaleDateString();
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Enhanced error handling for API calls
    async fetchWithErrorHandling(url, options = {}) {
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        try {
            const response = await fetch(url, defaultOptions);
            
            if (!response.ok) {
                let errorMessage = `HTTP ${response.status}`;
                try {
                    const errorData = await response.json();
                    errorMessage = errorData.error || errorData.message || errorMessage;
                } catch (parseError) {
                    // Ignore JSON parse errors, use HTTP status
                }
                throw new Error(errorMessage);
            }
            
            return response;
        } catch (error) {
            if (error.name === 'TypeError' && error.message.includes('Failed to fetch')) {
                throw new Error('Network connection failed');
            }
            throw error;
        }
    }
}

// Global functions for template usage
function refreshData() {
    window.dashboard.refreshData();
}

function updateTimeRange() {
    window.dashboard.updateTimeRange();
}

function saveSettings() {
    window.dashboard.saveSettings();
}

// User chat global functions
function selectChatRoom(roomId) {
    if (window.dashboard) {
        window.dashboard.selectChatRoom(roomId);
    }
}

function startDirectChat(userId) {
    if (window.dashboard) {
        window.dashboard.startDirectChat(userId);
    }
}

function createNewGroupChat() {
    if (window.dashboard) {
        window.dashboard.createNewGroupChat();
    }
}

// Initialize dashboard when page loads
document.addEventListener('DOMContentLoaded', () => {
    window.dashboard = new GoodooDashboard();
    // Start connection monitoring
    window.dashboard.startConnectionMonitoring();
});

// Handle page visibility changes
document.addEventListener('visibilitychange', () => {
    if (document.hidden) {
        window.dashboard.stopAutoRefresh();
    } else {
        window.dashboard.startAutoRefresh();
    }
});