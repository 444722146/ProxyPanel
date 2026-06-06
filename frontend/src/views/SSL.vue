<template>
  <div class="ssl-view">
    <!-- SSL证书列表 -->
    <el-card class="ssl-card">
      <template #header>
        <div class="card-header">
          <span>SSL证书管理</span>
        </div>
      </template>
      
      <el-table :data="proxiesWithSSL" v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="代理名称" width="160" />
        <el-table-column prop="domain" label="域名" width="200" />
        <el-table-column label="证书类型" width="130">
          <template #default="{ row }">
            <el-tag v-if="row.ssl_type === 'letsencrypt'" type="success" size="small">
              Let's Encrypt
            </el-tag>
            <el-tag v-else-if="row.ssl_type === 'zerossl'" type="warning" size="small">
              ZeroSSL
            </el-tag>
            <el-tag v-else-if="row.ssl_type === 'selfsigned'" type="info" size="small">
              自签名
            </el-tag>
            <el-tag v-else-if="row.ssl_enabled" type="" size="small">
              手动上传
            </el-tag>
            <el-tag v-else type="danger" size="small">未配置</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="到期时间" width="170">
          <template #default="{ row }">
            <template v-if="row.ssl_expires_at">
              <span :class="{ 'text-danger': isExpiringSoon(row.ssl_expires_at) }">
                {{ formatDate(row.ssl_expires_at) }}
              </span>
              <el-tag
                v-if="isExpiringSoon(row.ssl_expires_at)"
                type="danger"
                size="small"
                style="margin-left: 4px;"
              >即将过期</el-tag>
            </template>
            <span v-else-if="row.ssl_enabled" style="color: #999">-</span>
            <span v-else style="color: #999">-</span>
          </template>
        </el-table-column>
        <el-table-column label="自动续签" width="100">
          <template #default="{ row }">
            <el-switch
              v-if="row.ssl_type === 'letsencrypt' || row.ssl_type === 'zerossl'"
              v-model="row.ssl_auto_renew"
              @change="handleToggleAutoRenew(row)"
              active-color="#67c23a"
              size="small"
            />
            <span v-else style="color: #999">-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="380" fixed="right">
          <template #default="{ row }">
            <el-button
              size="small"
              type="success"
              @click="showFreeCertDialog(row)"
              v-if="(!row.ssl_enabled || row.ssl_type === 'selfsigned') && !isIPDomain(row.domain)"
            >
              <el-icon><Key /></el-icon>
              申请证书
            </el-button>
            <el-button
              size="small"
              type="warning"
              @click="handleRenew(row)"
              v-if="row.ssl_type === 'letsencrypt' || row.ssl_type === 'zerossl'"
              :loading="renewingIds.has(row.id)"
            >
              <el-icon><Refresh /></el-icon>
              续签
            </el-button>
            <el-button
              size="small"
              type="primary"
              @click="showUploadDialog(row)"
              v-if="!row.ssl_enabled"
            >
              <el-icon><Upload /></el-icon>
              上传
            </el-button>
            <el-button
              size="small"
              type="info"
              @click="handleGenerate(row)"
              v-if="!row.ssl_enabled"
            >
              <el-icon><MagicStick /></el-icon>
              自签名
            </el-button>
            <el-button
              size="small"
              type="danger"
              @click="handleRemove(row)"
              v-if="row.ssl_enabled"
            >
              <el-icon><Delete /></el-icon>
              移除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 申请证书对话框 -->
    <el-dialog
      v-model="freeCertDialogVisible"
      title="申请SSL证书"
      width="520px"
      :close-on-click-modal="false"
    >
      <div class="free-cert-info">
        <el-alert type="info" :closable="false" style="margin-bottom: 20px;">
          <template #title>
            <div>为域名 <strong>{{ freeCertDomain }}</strong> 申请SSL证书</div>
          </template>
        </el-alert>

        <el-form label-width="100px">
          <el-form-item label="证书机构">
            <el-radio-group v-model="freeCertCA">
              <el-radio value="zerossl">
                <span>ZeroSSL</span>
                <el-tag size="small" type="success" style="margin-left: 6px;">推荐</el-tag>
              </el-radio>
              <el-radio value="letsencrypt">Let's Encrypt</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="验证方式">
            <el-tag type="success" size="small">DNS 验证</el-tag>
            <div style="margin-top: 6px; color: #909399; font-size: 12px;">
              证书申请通过DNS来验证域名的所有权，您需要添加一条TXT或CNAME记录来完成域名所有权验证
            </div>
          </el-form-item>

          <el-form-item label="自动续签">
            <el-switch v-model="freeCertAutoRenew" active-color="#67c23a" />
            <span style="margin-left: 8px; color: #909399; font-size: 12px;">
              到期前30天自动续签
            </span>
          </el-form-item>
        </el-form>

        <el-alert type="warning" :closable="false" style="margin-top: 10px;">
          <template #title>
            <div>
              <p>1. 请确保域名已解析到本服务器</p>
              <p>2. 申请过程需要您到域名DNS服务商添加TXT记录进行验证</p>
              <p>3. 申请过程可能需要 30-60 秒</p>
            </div>
          </template>
        </el-alert>
      </div>

      <template #footer>
        <el-button @click="freeCertDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          @click="handleRequestFreeCert"
          :loading="requesting"
        >
          申请证书
        </el-button>
      </template>
    </el-dialog>

    <!-- 上传证书对话框 -->
    <el-dialog
      v-model="uploadDialogVisible"
      title="上传SSL证书"
      width="500px"
    >
      <el-form label-width="100px">
        <el-form-item label="域名">
          <el-input :value="uploadingDomain" disabled />
        </el-form-item>
        
        <el-form-item label="证书文件">
          <el-upload
            ref="certUploadRef"
            :auto-upload="false"
            :limit="1"
            accept=".crt,.pem,.cer"
          >
            <el-button type="primary">
              <el-icon><Upload /></el-icon>
              选择证书文件
            </el-button>
            <template #tip>
              <div class="el-upload__tip">
                支持 .crt, .pem, .cer 格式
              </div>
            </template>
          </el-upload>
        </el-form-item>
        
        <el-form-item label="私钥文件">
          <el-upload
            ref="keyUploadRef"
            :auto-upload="false"
            :limit="1"
            accept=".key,.pem"
          >
            <el-button type="primary">
              <el-icon><Upload /></el-icon>
              选择私钥文件
            </el-button>
            <template #tip>
              <div class="el-upload__tip">
                支持 .key, .pem 格式
              </div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleUpload" :loading="uploading">
          上传并配置
        </el-button>
      </template>
    </el-dialog>

    <!-- 证书信息提示 -->
    <el-card class="info-card">
      <template #header>
        <span>SSL证书说明</span>
      </template>
      <el-alert type="info" :closable="false">
        <template #title>
          <div>
            <p>1. <strong>申请证书</strong>：通过 Let's Encrypt / ZeroSSL 自动申请，默认开启自动续签</p>
            <p>2. <strong>上传证书</strong>：上传您从证书颁发机构获取的正式SSL证书</p>
            <p>3. <strong>自签名证书</strong>：快速生成测试证书（仅用于开发测试，浏览器会提示不安全）</p>
            <p>4. IP地址类型域名不支持证书申请，请使用上传证书或自签名证书</p>
          </div>
        </template>
      </el-alert>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getProxyRules,
  uploadSSLCertificate,
  generateSelfSignedCertificate,
  removeSSLCertificate,
  requestFreeCert,
  renewCertificate,
  toggleAutoRenew
} from '@/api/proxy'

const loading = ref(false)
const uploading = ref(false)
const requesting = ref(false)
const proxies = ref([])
const uploadDialogVisible = ref(false)
const uploadingId = ref(null)
const uploadingDomain = ref('')
const certUploadRef = ref(null)
const keyUploadRef = ref(null)

// 证书申请相关
const freeCertDialogVisible = ref(false)
const freeCertId = ref(null)
const freeCertDomain = ref('')
const freeCertCA = ref('letsencrypt')
const freeCertAutoRenew = ref(true)
const freeCertIsIP = ref(false)
const renewingIds = ref(new Set())

const proxiesWithSSL = computed(() => {
  return proxies.value.map(p => ({
    ...p,
    ssl_cert: p.ssl_cert || '未配置',
    ssl_key: p.ssl_key || '未配置'
  }))
})

const isExpiringSoon = (dateStr) => {
  if (!dateStr) return false
  const d = new Date(dateStr)
  const now = new Date()
  const daysLeft = (d - now) / (1000 * 60 * 60 * 24)
  return daysLeft < 30
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

const isIPDomain = (domain) => {
  const parts = domain.split('.')
  if (parts.length === 4) {
    return parts.every(p => /^\d+$/.test(p))
  }
  return domain.includes(':')
}

const loadProxies = async () => {
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

// 证书申请
const showFreeCertDialog = (row) => {
  freeCertId.value = row.id
  freeCertDomain.value = row.domain
  freeCertCA.value = 'zerossl'
  freeCertAutoRenew.value = true
  freeCertIsIP.value = isIPDomain(row.domain)
  freeCertDialogVisible.value = true
}

const handleRequestFreeCert = async () => {
  requesting.value = true
  try {
    await requestFreeCert(freeCertId.value, freeCertCA.value)
    ElMessage.success('SSL证书申请成功！')
    freeCertDialogVisible.value = false
    loadProxies()
  } catch (error) {
    ElMessage.error('申请失败: ' + error.message)
  } finally {
    requesting.value = false
  }
}

// 续签
const handleRenew = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定续签域名 "${row.domain}" 的SSL证书吗？`,
      '续签证书',
      {
        confirmButtonText: '续签',
        cancelButtonText: '取消',
        type: 'info'
      }
    )
    
    renewingIds.value.add(row.id)
    renewingIds.value = new Set(renewingIds.value)
    
    await renewCertificate(row.id)
    ElMessage.success('证书续签成功')
    loadProxies()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('续签失败: ' + error.message)
    }
  } finally {
    renewingIds.value.delete(row.id)
    renewingIds.value = new Set(renewingIds.value)
  }
}

// 自动续签开关
const handleToggleAutoRenew = async (row) => {
  try {
    await toggleAutoRenew(row.id)
    ElMessage.success(row.ssl_auto_renew ? '已开启自动续签' : '已关闭自动续签')
  } catch (error) {
    row.ssl_auto_renew = !row.ssl_auto_renew
    ElMessage.error('操作失败: ' + error.message)
  }
}

// 上传证书
const showUploadDialog = (row) => {
  uploadingId.value = row.id
  uploadingDomain.value = row.domain
  uploadDialogVisible.value = true
}

const handleUpload = async () => {
  const certFiles = certUploadRef.value?.uploadFiles || []
  const keyFiles = keyUploadRef.value?.uploadFiles || []
  
  if (certFiles.length === 0) {
    ElMessage.warning('请选择证书文件')
    return
  }
  
  if (keyFiles.length === 0) {
    ElMessage.warning('请选择私钥文件')
    return
  }
  
  uploading.value = true
  try {
    const certFile = certFiles[0].raw
    const keyFile = keyFiles[0].raw
    
    await uploadSSLCertificate(uploadingId.value, certFile, keyFile)
    ElMessage.success('SSL证书上传成功')
    uploadDialogVisible.value = false
    loadProxies()
  } catch (error) {
    ElMessage.error('上传失败: ' + error.message)
  } finally {
    uploading.value = false
  }
}

// 自签名
const handleGenerate = async (row) => {
  try {
    await ElMessageBox.confirm(
      `将为域名 "${row.domain}" 生成自签名测试证书。\n注意：此证书仅用于测试，浏览器会提示不安全。`,
      '生成自签名证书',
      {
        confirmButtonText: '生成',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    loading.value = true
    await generateSelfSignedCertificate(row.id)
    ElMessage.success('自签名证书已生成')
    loadProxies()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('生成失败: ' + error.message)
    }
  } finally {
    loading.value = false
  }
}

// 移除
const handleRemove = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定移除域名 "${row.domain}" 的SSL证书吗？此操作将禁用HTTPS。`,
      '移除证书',
      {
        confirmButtonText: '移除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    loading.value = true
    await removeSSLCertificate(row.id)
    ElMessage.success('SSL证书已移除')
    loadProxies()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('移除失败: ' + error.message)
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadProxies()
})
</script>

<style scoped>
.ssl-view {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.ssl-card {
  flex: 1;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-card {
  margin-top: 0;
}

.el-alert p {
  margin: 5px 0;
  line-height: 1.6;
}

.text-danger {
  color: #f56c6c;
  font-weight: bold;
}

.free-cert-info {
  padding: 0 10px;
}
</style>
