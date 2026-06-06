<template>
  <div class="proxy-list">
    <!-- 工具栏 -->
    <el-card class="toolbar-card">
      <el-row :gutter="20">
        <el-col :span="16">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索代理（名称、域名、目标地址）"
            clearable
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="8" style="text-align: right;">
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            新增代理
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 代理列表 -->
    <el-card class="list-card">
      <el-table
        :data="filteredProxies"
        v-loading="loading"
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" width="160" />
        <el-table-column prop="domain" label="访问域名" width="180" />
        <el-table-column prop="port" label="端口" width="80">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.port || (row.ssl_enabled ? 443 : 80) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="target_url" label="目标地址" min-width="200" show-overflow-tooltip />
        <el-table-column label="访问地址" width="220">
          <template #default="{ row }">
            <el-link
              v-if="row.access_url"
              :href="row.access_url"
              target="_blank"
              type="primary"
              :underline="false"
            >
              <el-icon style="margin-right: 4px"><Link /></el-icon>
              {{ row.access_url }}
            </el-link>
            <span v-else style="color: #999">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="状态" width="90">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="handleToggle(row)"
              active-color="#67c23a"
              inactive-color="#dcdfe6"
            />
          </template>
        </el-table-column>
        <el-table-column prop="ssl_enabled" label="SSL" width="80">
          <template #default="{ row }">
            <el-tag :type="row.ssl_enabled ? 'success' : 'info'" size="small">
              {{ row.ssl_enabled ? '已配置' : '未配置' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="showEditDialog(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button size="small" type="danger" link @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="680px"
      :close-on-click-modal="false"
    >
      <el-form
        :model="formData"
        :rules="formRules"
        ref="formRef"
        label-width="100px"
        class="proxy-form"
      >
        <!-- 基础信息 -->
        <div class="form-section">
          <div class="section-title">
            <el-icon><InfoFilled /></el-icon>
            <span>基础信息</span>
          </div>

          <el-form-item label="代理名称" prop="name">
            <el-input v-model="formData.name" placeholder="如：API Gateway 01" />
          </el-form-item>

          <el-row :gutter="16">
            <el-col :span="16">
              <el-form-item label="访问域名" prop="domain">
                <el-input v-model="formData.domain" placeholder="如：api.example.com" />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="监听端口" prop="port">
                <el-input-number
                  v-model="formData.port"
                  :min="0"
                  :max="65535"
                  :controls="false"
                  placeholder="默认"
                  style="width: 100%;"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item label="目标地址" prop="target_url">
            <el-input v-model="formData.target_url" placeholder="如：http://192.168.1.100:8080" />
          </el-form-item>
        </div>

        <!-- 高级配置 -->
        <div class="form-section">
          <div class="section-title">
            <el-icon><Setting /></el-icon>
            <span>高级配置</span>
          </div>

          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="伪装IP" prop="fake_ip">
                <el-input v-model="formData.fake_ip" placeholder="默认：127.0.0.1" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="认证Token" prop="token">
                <el-input v-model="formData.token" placeholder="可选" show-password />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item label="IP白名单" prop="whitelist">
            <el-input
              v-model="formData.whitelist"
              type="textarea"
              :rows="2"
              placeholder="多个IP用逗号分隔，如：192.168.1.1,10.0.0.1"
            />
          </el-form-item>

          <el-form-item label="启用状态" prop="enabled">
            <el-switch
              v-model="formData.enabled"
              active-text="启用"
              inactive-text="禁用"
            />
          </el-form-item>
        </div>

        <!-- SSL证书配置 -->
        <div class="form-section">
          <div class="section-title">
            <el-icon><Lock /></el-icon>
            <span>SSL证书配置（可选）</span>
          </div>

          <el-form-item label="启用SSL" prop="ssl_enabled">
            <el-switch
              v-model="formData.ssl_enabled"
              active-text="开启"
              inactive-text="关闭"
            />
          </el-form-item>

          <template v-if="formData.ssl_enabled">
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="证书路径" prop="ssl_cert">
                  <el-input v-model="formData.ssl_cert" placeholder="证书文件路径" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="私钥路径" prop="ssl_key">
                  <el-input v-model="formData.ssl_key" placeholder="私钥文件路径" />
                </el-form-item>
              </el-col>
            </el-row>
          </template>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ isEdit ? '更新' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getProxyRules,
  createProxyRule,
  updateProxyRule,
  deleteProxyRule,
  toggleProxyRule
} from '@/api/proxy'

const loading = ref(false)
const submitting = ref(false)
const proxies = ref([])
const searchKeyword = ref('')
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref(null)
const formRef = ref(null)

const dialogTitle = computed(() => isEdit.value ? '编辑代理' : '新增代理')

const formData = ref({
  name: '',
  domain: '',
  port: 0,
  target_url: '',
  fake_ip: '127.0.0.1',
  token: '',
  whitelist: '',
  enabled: true,
  ssl_enabled: false,
  ssl_cert: '',
  ssl_key: ''
})

const formRules = {
  name: [{ required: true, message: '请输入代理名称', trigger: 'blur' }],
  domain: [{ required: true, message: '请输入域名', trigger: 'blur' }],
  target_url: [{ required: true, message: '请输入目标地址', trigger: 'blur' }]
}

const filteredProxies = computed(() => {
  if (!searchKeyword.value) return proxies.value
  const keyword = searchKeyword.value.toLowerCase()
  return proxies.value.filter(p => 
    p.name.toLowerCase().includes(keyword) ||
    p.domain.toLowerCase().includes(keyword) ||
    p.target_url.toLowerCase().includes(keyword)
  )
})

const loadData = async () => {
  loading.value = true
  try {
    const res = await getProxyRules()
    proxies.value = res.data || []
  } catch (error) {
    ElMessage.error('加载代理列表失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  // 搜索逻辑已通过computed实现
}

const resetForm = () => {
  formData.value = {
    name: '',
    domain: '',
    port: 0,
    target_url: '',
    fake_ip: '127.0.0.1',
    token: '',
    whitelist: '',
    enabled: true,
    ssl_enabled: false,
    ssl_cert: '',
    ssl_key: ''
  }
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

const showCreateDialog = () => {
  resetForm()
  isEdit.value = false
  editingId.value = null
  dialogVisible.value = true
}

const showEditDialog = (row) => {
  isEdit.value = true
  editingId.value = row.id
  formData.value = {
    name: row.name,
    domain: row.domain,
    port: row.port || 0,
    target_url: row.target_url,
    fake_ip: row.fake_ip || '127.0.0.1',
    token: row.token || '',
    whitelist: row.whitelist || '',
    enabled: row.enabled,
    ssl_enabled: row.ssl_enabled,
    ssl_cert: row.ssl_cert || '',
    ssl_key: row.ssl_key || ''
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
  } catch (error) {
    return
  }

  submitting.value = true
  try {
    if (isEdit.value) {
      await updateProxyRule(editingId.value, formData.value)
      ElMessage.success('代理规则已更新')
    } else {
      await createProxyRule(formData.value)
      ElMessage.success('代理规则已创建')
    }
    dialogVisible.value = false
    loadData()
  } catch (error) {
    const msg = error.response?.data?.message || error.message || '未知错误'
    ElMessage.error((isEdit.value ? '更新失败: ' : '创建失败: ') + msg)
  } finally {
    submitting.value = false
  }
}

const handleToggle = async (row) => {
  try {
    await toggleProxyRule(row.id)
    ElMessage.success(row.enabled ? '代理已启用' : '代理已禁用')
  } catch (error) {
    row.enabled = !row.enabled // 恢复状态
    ElMessage.error('切换状态失败: ' + error.message)
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定删除代理 "${row.name}" 吗？此操作将同时删除对应的Nginx配置。`,
      '删除确认',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await deleteProxyRule(row.id)
    ElMessage.success('代理规则已删除')
    loadData()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + error.message)
    }
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.proxy-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.toolbar-card {
  margin-bottom: 0;
}

.list-card {
  flex: 1;
}

.proxy-form {
  padding: 0 10px;
}

.form-section {
  margin-bottom: 16px;
}

.form-section:last-child {
  margin-bottom: 0;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
  padding-bottom: 10px;
  border-bottom: 1px solid #ebeef5;
}

.section-title .el-icon {
  font-size: 16px;
  color: #409eff;
}

:deep(.el-form-item) {
  margin-bottom: 18px;
}

:deep(.el-form-item:last-child) {
  margin-bottom: 0;
}
</style>