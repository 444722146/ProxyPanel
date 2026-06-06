<template>
  <div class="app-container">
    <el-container>
      <!-- 侧边栏 -->
      <el-aside width="220px" class="sidebar">
        <div class="logo">
          <el-icon :size="32" color="#409eff">
            <Connection />
          </el-icon>
          <h2>ProxyPanel</h2>
        </div>
        
        <el-menu
          :default-active="activeMenu"
          router
          class="sidebar-menu"
        >
          <el-menu-item index="/">
            <el-icon><Dashboard /></el-icon>
            <span>仪表盘</span>
          </el-menu-item>
          
          <el-menu-item index="/proxy">
            <el-icon><Server /></el-icon>
            <span>代理管理</span>
          </el-menu-item>
          
          <el-menu-item index="/logs">
            <el-icon><Document /></el-icon>
            <span>日志查看</span>
          </el-menu-item>
          
          <el-menu-item index="/ssl">
            <el-icon><Key /></el-icon>
            <span>SSL证书</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      
      <!-- 主内容区 -->
      <el-container>
        <el-header class="header">
          <div class="header-content">
            <h1>{{ pageTitle }}</h1>
            <div class="header-actions">
              <el-button type="primary" @click="testNginx" :loading="testing">
                <el-icon><Check /></el-icon>
                测试Nginx
              </el-button>
              <el-button type="success" @click="syncConfigs" :loading="syncing">
                <el-icon><Refresh /></el-icon>
                同步配置
              </el-button>
            </div>
          </div>
        </el-header>
        
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { testNginxConfig, syncProxyConfigs } from '@/api/proxy'

const route = useRoute()
const testing = ref(false)
const syncing = ref(false)

const activeMenu = computed(() => {
  return route.path
})

const pageTitle = computed(() => {
  const titles = {
    '/': '仪表盘',
    '/proxy': '代理管理',
    '/logs': '日志查看',
    '/ssl': 'SSL证书管理'
  }
  return titles[route.path] || 'ProxyPanel'
})

const testNginx = async () => {
  testing.value = true
  try {
    const res = await testNginxConfig()
    if (res.data.success) {
      ElMessage.success('Nginx配置正常')
    } else {
      ElMessage.error('Nginx配置有误，请检查')
    }
  } catch (error) {
    ElMessage.error('测试失败: ' + error.message)
  } finally {
    testing.value = false
  }
}

const syncConfigs = async () => {
  syncing.value = true
  try {
    await syncProxyConfigs()
    ElMessage.success('配置已同步到Nginx')
  } catch (error) {
    ElMessage.error('同步失败: ' + error.message)
  } finally {
    syncing.value = false
  }
}
</script>

<style scoped>
.app-container {
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

.sidebar {
  background: #f5f7fa;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 20px;
  border-bottom: 1px solid #e4e7ed;
}

.logo h2 {
  margin-left: 12px;
  color: #303133;
  font-size: 18px;
}

.sidebar-menu {
  border-right: none;
  flex: 1;
}

.header {
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 20px;
  display: flex;
  align-items: center;
}

.header-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-content h1 {
  font-size: 20px;
  color: #303133;
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.main-content {
  background: #f5f7fa;
  padding: 20px;
  overflow-y: auto;
}
</style>