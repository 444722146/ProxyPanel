<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon :size="40" color="#67c23a">
              <SuccessFilled />
            </el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.total }}</div>
              <div class="stat-label">总代理数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon :size="40" color="#409eff">
              <CircleCheckFilled />
            </el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.enabled }}</div>
              <div class="stat-label">已启用</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon :size="40" color="#e6a23c">
              <WarningFilled />
            </el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.disabled }}</div>
              <div class="stat-label">已禁用</div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon :size="40" color="#f56c6c">
              <Key />
            </el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.ssl }}</div>
              <div class="stat-label">SSL证书</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 快速操作 -->
    <el-card class="action-card">
      <template #header>
        <div class="card-header">
          <span>快速操作</span>
        </div>
      </template>
      <el-row :gutter="20">
        <el-col :span="6">
          <el-button type="primary" size="large" @click="$router.push('/proxy')">
            <el-icon><Plus /></el-icon>
            新增代理
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button type="success" size="large" @click="handleSync">
            <el-icon><Refresh /></el-icon>
            同步配置
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button type="warning" size="large" @click="$router.push('/logs')">
            <el-icon><Document /></el-icon>
            查看日志
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button type="info" size="large" @click="$router.push('/ssl')">
            <el-icon><Key /></el-icon>
            SSL管理
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 最近代理 -->
    <el-card class="recent-card">
      <template #header>
        <div class="card-header">
          <span>最近添加的代理</span>
          <el-button type="text" @click="$router.push('/proxy')">查看全部</el-button>
        </div>
      </template>
      <el-table :data="recentProxies" style="width: 100%">
        <el-table-column prop="name" label="名称" width="180" />
        <el-table-column prop="domain" label="域名" width="180" />
        <el-table-column label="端口" width="80">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.port || (row.ssl_enabled ? 443 : 80) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="target_url" label="目标地址" />
        <el-table-column prop="enabled" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'danger'">
              {{ row.enabled ? '已启用' : '已禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ssl_enabled" label="SSL" width="80">
          <template #default="{ row }">
            <el-tag :type="row.ssl_enabled ? 'success' : 'info'" size="small">
              {{ row.ssl_enabled ? '已配置' : '未配置' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getProxyRules, syncProxyConfigs } from '@/api/proxy'

const proxies = ref([])
const syncing = ref(false)

const stats = computed(() => ({
  total: proxies.value.length,
  enabled: proxies.value.filter(p => p.enabled).length,
  disabled: proxies.value.filter(p => !p.enabled).length,
  ssl: proxies.value.filter(p => p.ssl_enabled).length
}))

const recentProxies = computed(() => {
  return proxies.value.slice(0, 5)
})

const loadData = async () => {
  try {
    const res = await getProxyRules()
    proxies.value = res.data || []
  } catch (error) {
    ElMessage.error('加载代理数据失败: ' + error.message)
  }
}

const handleSync = async () => {
  syncing.value = true
  try {
    await syncProxyConfigs()
    ElMessage.success('配置已同步')
  } catch (error) {
    ElMessage.error('同步失败: ' + error.message)
  } finally {
    syncing.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.stats-row {
  margin-bottom: 0;
}

.stat-card {
  height: 120px;
}

.stat-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.stat-info {
  text-align: right;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.action-card {
  margin-top: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.recent-card {
  flex: 1;
}
</style>