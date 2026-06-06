<template>
  <div class="logs-view">
    <!-- 日志类型选择 -->
    <el-card class="filter-card">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-select v-model="logType" placeholder="日志类型">
            <el-option label="访问日志" value="access" />
            <el-option label="错误日志" value="error" />
          </el-select>
        </el-col>
        
        <el-col :span="6">
          <el-select v-model="selectedDomain" placeholder="选择域名" clearable>
            <el-option label="全部域名" value="" />
            <el-option
              v-for="proxy in proxies"
              :key="proxy.id"
              :label="proxy.domain"
              :value="proxy.domain"
            />
          </el-select>
        </el-col>
        
        <el-col :span="6">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索关键词"
            clearable
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        
        <el-col :span="6">
          <el-button type="primary" @click="loadLogs" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新日志
          </el-button>
          <el-button type="danger" @click="handleClear" :loading="clearing">
            <el-icon><Delete /></el-icon>
            清空
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 日志内容 -->
    <el-card class="log-card">
      <template #header>
        <div class="card-header">
          <span>日志内容（最近 {{ lineCount }} 行）</span>
          <div class="header-right">
            <el-input-number
              v-model="lineCount"
              :min="10"
              :max="500"
              :step="50"
              size="small"
              @change="loadLogs"
            />
          </div>
        </div>
      </template>
      
      <div class="log-container" v-loading="loading">
        <pre class="log-content">{{ logContent }}</pre>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getProxyRules,
  getAccessLog,
  getErrorLog,
  getGeneralAccessLog,
  getGeneralErrorLog,
  clearLog,
  searchLog
} from '@/api/proxy'

const loading = ref(false)
const clearing = ref(false)
const proxies = ref([])
const logType = ref('access')
const selectedDomain = ref('')
const searchKeyword = ref('')
const lineCount = ref(100)
const logContent = ref('')

const loadProxies = async () => {
  try {
    const res = await getProxyRules()
    proxies.value = res.data || []
  } catch (error) {
    ElMessage.error('加载代理列表失败: ' + error.message)
  }
}

const loadLogs = async () => {
  loading.value = true
  try {
    let res
    
    // 根据日志类型和域名选择不同的API
    if (logType.value === 'access') {
      if (selectedDomain.value) {
        res = await getAccessLog(selectedDomain.value, lineCount.value)
      } else {
        res = await getGeneralAccessLog(lineCount.value)
      }
    } else {
      if (selectedDomain.value) {
        res = await getErrorLog(selectedDomain.value, lineCount.value)
      } else {
        res = await getGeneralErrorLog(lineCount.value)
      }
    }
    
    // 处理日志数据
    const logs = res.data.logs || []
    if (logs.length === 0) {
      logContent.value = '暂无日志内容'
    } else {
      logContent.value = logs.join('\n')
    }
  } catch (error) {
    ElMessage.error('加载日志失败: ' + error.message)
    logContent.value = '加载日志失败: ' + error.message
  } finally {
    loading.value = false
  }
}

const handleSearch = async () => {
  if (!searchKeyword.value) {
    loadLogs()
    return
  }
  
  loading.value = true
  try {
    const res = await searchLog(
      searchKeyword.value,
      selectedDomain.value,
      logType.value,
      lineCount.value
    )
    
    const logs = res.data.logs || []
    if (logs.length === 0) {
      logContent.value = '未找到匹配的日志'
    } else {
      logContent.value = logs.join('\n')
    }
  } catch (error) {
    ElMessage.error('搜索失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const handleClear = async () => {
  try {
    await ElMessageBox.confirm(
      '确定清空日志吗？此操作不可恢复。',
      '清空确认',
      {
        confirmButtonText: '清空',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    clearing.value = true
    await clearLog(logType.value, selectedDomain.value)
    ElMessage.success('日志已清空')
    logContent.value = '日志已清空'
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('清空失败: ' + error.message)
    }
  } finally {
    clearing.value = false
  }
}

onMounted(() => {
  loadProxies()
  loadLogs()
})
</script>

<style scoped>
.logs-view {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.filter-card {
  margin-bottom: 0;
}

.log-card {
  flex: 1;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
}

.log-container {
  background: #1e1e1e;
  border-radius: 4px;
  padding: 16px;
  min-height: 500px;
  max-height: 600px;
  overflow-y: auto;
}

.log-content {
  color: #d4d4d4;
  font-family: 'Consolas', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-wrap: break-word;
  margin: 0;
}
</style>