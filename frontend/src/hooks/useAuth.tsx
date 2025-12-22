import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { authAPI } from '../services/api'
import { storage } from '../utils/storage'

interface User {
  id: number
  username: string
}

interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  loading: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let isMounted = true

    const bootstrap = async () => {
      try {
        const cachedUser = storage.getUser()
        if (cachedUser && isMounted) {
          setUser(cachedUser)
        }

        const accessToken = storage.getAccessToken()
        const refreshToken = storage.getRefreshToken()

        if (accessToken) {
          const response = await authAPI.getCurrentUser()
          if (!isMounted) return
          setUser(response.data)
          storage.setUser(response.data)
          return
        }

        if (refreshToken) {
          const { data } = await authAPI.refreshToken(refreshToken)
          storage.setAccessToken(data.token)
          storage.setRefreshToken(data.refresh_token)
          const response = await authAPI.getCurrentUser()
          if (!isMounted) return
          setUser(response.data)
          storage.setUser(response.data)
          return
        }
      } catch {
        storage.clearSession()
        if (isMounted) {
          setUser(null)
        }
      } finally {
        if (isMounted) {
          setLoading(false)
        }
      }
    }

    bootstrap()

    return () => {
      isMounted = false
    }
  }, [])

  const login = async (username: string, password: string) => {
    const response = await authAPI.login(username, password)
    storage.setAccessToken(response.data.token)
    storage.setRefreshToken(response.data.refresh_token)
    storage.setUser(response.data.user)
    setUser(response.data.user)
  }

  const logout = async () => {
    const refreshToken = storage.getRefreshToken()
    if (refreshToken) {
      try {
        await authAPI.logout(refreshToken)
      } catch {
        // Ignore logout errors
      }
    }
    storage.clearSession()
    setUser(null)
    return
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        login,
        logout,
        loading,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

