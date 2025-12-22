import axios from 'axios'
import type { InternalAxiosRequestConfig } from 'axios'
import { storage } from '../utils/storage'

declare module 'axios' {
  interface InternalAxiosRequestConfig {
    _retry?: boolean
    skipAuthRefresh?: boolean
  }

  interface AxiosRequestConfig {
    skipAuthRefresh?: boolean
  }
}

type AuthenticatedRequestConfig = InternalAxiosRequestConfig

const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add token
api.interceptors.request.use(
  (config) => {
    const token = storage.getAccessToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

let isRefreshing = false
let failedQueue: Array<{
  resolve: (value?: unknown) => void
  reject: (reason?: any) => void
}> = []

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error)
    } else {
      prom.resolve(token)
    }
  })
  failedQueue = []
}

// Response interceptor to handle 401 errors
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config as AuthenticatedRequestConfig

    if (error.response?.status !== 401 || originalRequest?.skipAuthRefresh) {
      if (error.response?.status === 401) {
        storage.clearSession()
        window.location.href = '/login'
      }
      return Promise.reject(error)
    }

    if (originalRequest?._retry) {
      storage.clearSession()
      window.location.href = '/login'
      return Promise.reject(error)
    }

    const refreshToken = storage.getRefreshToken()
    if (!refreshToken) {
      storage.clearSession()
      window.location.href = '/login'
      return Promise.reject(error)
    }

    originalRequest._retry = true

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        failedQueue.push({
          resolve: (token: unknown) => {
            if (!token || typeof token !== 'string') {
              reject(error)
              return
            }
            originalRequest.headers.Authorization = `Bearer ${token}`
            resolve(api(originalRequest))
          },
          reject,
        })
      })
    }

    isRefreshing = true

    try {
      const { data } = await authAPI.refreshToken(refreshToken)
      storage.setAccessToken(data.token)
      storage.setRefreshToken(data.refresh_token)
      api.defaults.headers.common.Authorization = `Bearer ${data.token}`
      processQueue(null, data.token)
      originalRequest.headers.Authorization = `Bearer ${data.token}`
      return api(originalRequest)
    } catch (refreshError) {
      processQueue(refreshError, null)
      storage.clearSession()
      window.location.href = '/login'
      return Promise.reject(refreshError)
    } finally {
      isRefreshing = false
    }
  }
)

export default api

// API methods
export const authAPI = {
  login: (username: string, password: string) =>
    api.post('/auth/login', { username, password }),

  refreshToken: (refreshToken: string) =>
    api.post(
      '/auth/refresh',
      { refresh_token: refreshToken },
      { skipAuthRefresh: true }
    ),

  logout: (refreshToken: string) =>
    api.post(
      '/auth/logout',
      { refresh_token: refreshToken },
      { skipAuthRefresh: true }
    ),

  getCurrentUser: () => api.get('/auth/me'),
  
  generateToken: (tokenName: string, expiresAt: string) =>
    api.post('/auth/tokens', { token_name: tokenName, expires_at: expiresAt }),
  
  listTokens: () => api.get('/auth/tokens'),
  
  deleteToken: (id: number) => api.delete(`/auth/tokens/${id}`),

  issueAdminToken: (username: string, password: string, tokenName?: string) =>
    api.post(
      '/auth/admin-token',
      { username, password, token_name: tokenName },
      { skipAuthRefresh: true }
    ),
}

export const tasksAPI = {
  list: (limit = 20, offset = 0) =>
    api.get('/tasks', { params: { limit, offset } }),
  
  get: (id: number) => api.get(`/tasks/${id}`),
  
  create: (data: any) => api.post('/tasks', data),
  
  update: (id: number, data: any) => api.put(`/tasks/${id}`, data),
  
  delete: (id: number) => api.delete(`/tasks/${id}`),
  
  start: (id: number) => api.post(`/tasks/${id}/start`),
  
  stop: (id: number) => api.post(`/tasks/${id}/stop`),
  
  pause: (id: number) => api.post(`/tasks/${id}/pause`),
  
  getStatus: (id: number) => api.get(`/tasks/${id}/status`),
  
  getResults: (id: number, limit = 20, offset = 0) =>
    api.get(`/tasks/${id}/results`, { params: { limit, offset } }),
}

export const productsAPI = {
  list: (params?: {
    limit?: number
    offset?: number
    brand?: string
    catalog?: string
    in_stock?: boolean
    search?: string
    status?: string
  }) => api.get('/products', { params }),
  
  get: (id: number) => api.get(`/products/${id}`),
  
  getByElasticID: (elasticID: string) => api.get(`/products/elastic/${elasticID}`),
  
  delete: (id: number) => api.delete(`/products/${id}`),
  
  getStats: () => api.get('/products/stats'),
  
  startCrawl: () => api.post('/products/crawl'),

  updateStatus: (id: number, status: 'pending' | 'approved' | 'rejected') =>
    api.patch(`/products/${id}/status`, { status }),
}

export const crawlerAPI = {
  getEmbroideryConfig: () => api.get('/products/crawl-config'),
  updateEmbroideryConfig: (payloadOverrides: Record<string, unknown>) =>
    api.put('/products/crawl-config', { payload_overrides: payloadOverrides }),
}

export const proxiesAPI = {
  list: () => api.get('/proxies'),

  create: (data: any) => api.post('/proxies', data),

  delete: (id: number) => api.delete(`/proxies/${id}`),

  test: (data: any) => api.post('/proxies/test', data),
}

export const statsAPI = {
  get: () => api.get('/stats'),
}

