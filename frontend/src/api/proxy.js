import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 120000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 响应拦截器
api.interceptors.response.use(
  response => {
    const res = response.data
    if (res.code !== 200) {
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res
  },
  error => {
    const message = error.response?.data?.message || error.message || '请求失败'
    return Promise.reject({ ...error, message })
  }
)

// 代理规则API
export const getProxyRules = () => api.get('/proxy')
export const getProxyRule = (id) => api.get(`/proxy/${id}`)
export const createProxyRule = (data) => api.post('/proxy', data)
export const updateProxyRule = (id, data) => api.put(`/proxy/${id}`, data)
export const deleteProxyRule = (id) => api.delete(`/proxy/${id}`)
export const toggleProxyRule = (id) => api.post(`/proxy/${id}/toggle`)

// Nginx管理API
export const testNginxConfig = () => api.post('/nginx/test')
export const syncProxyConfigs = () => api.post('/nginx/sync')

// 日志API
export const getAccessLog = (domain, lines = 100) => api.get(`/log/access`, { params: { domain, lines } })
export const getErrorLog = (domain, lines = 100) => api.get(`/log/error`, { params: { domain, lines } })
export const getGeneralAccessLog = (lines = 100) => api.get(`/log/access/general`, { params: { lines } })
export const getGeneralErrorLog = (lines = 100) => api.get(`/log/error/general`, { params: { lines } })
export const clearLog = (type, domain) => api.delete(`/log/clear`, { params: { type, domain } })
export const searchLog = (keyword, domain, type, lines = 100) => api.get(`/log/search`, { params: { keyword, domain, type, lines } })

// SSL证书API
export const uploadSSLCertificate = (id, certFile, keyFile) => {
  const formData = new FormData()
  formData.append('cert', certFile)
  formData.append('key', keyFile)
  return api.post(`/ssl/${id}/upload`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const generateSelfSignedCertificate = (id) => api.post(`/ssl/${id}/generate`)
export const removeSSLCertificate = (id) => api.delete(`/ssl/${id}`)

// 免费证书API
export const requestFreeCert = (id, ca = 'letsencrypt') => api.post(`/ssl/${id}/request-free`, { ca })
export const renewCertificate = (id) => api.post(`/ssl/${id}/renew`)
export const getCertStatus = (id) => api.get(`/ssl/${id}/status`)
export const toggleAutoRenew = (id) => api.post(`/ssl/${id}/toggle-auto-renew`)

export default api