import { BrowserRouter as Router, Routes, Route, Navigate, useLocation } from 'react-router-dom'
import { AuthProvider, useAuth } from './hooks/useAuth'
import Login from './components/auth/Login'
import Layout from './components/layout/Layout'
import Dashboard from './components/dashboard/Dashboard'
import TaskList from './components/tasks/TaskList'
import TaskDetail from './components/tasks/TaskDetail'
import ProductList from './components/products/ProductList'
import ProductDetail from './components/products/ProductDetail'
import EmbroideryCrawlerConfig from './components/products/EmbroideryCrawlerConfig'
import ProxyList from './components/proxies/ProxyList'
import TokenList from './components/tokens/TokenList'
import TokenGenerator from './components/tokens/TokenGenerator'
import LiveLogs from './components/logs/LiveLogs'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, loading } = useAuth()
  const location = useLocation()

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 text-gray-600">
        Loading admin panel...
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />
  }

  return <>{children}</>
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <PrivateRoute>
            <Layout />
          </PrivateRoute>
        }
      >
        <Route index element={<Dashboard />} />
        <Route path="tasks" element={<TaskList />} />
        <Route path="tasks/:id" element={<TaskDetail />} />
        <Route path="products" element={<ProductList />} />
        <Route path="products/:id" element={<ProductDetail />} />
        <Route path="crawler/config" element={<EmbroideryCrawlerConfig />} />
        <Route path="proxies" element={<ProxyList />} />
        <Route path="tokens" element={<TokenList />} />
        <Route path="tokens/generate" element={<TokenGenerator />} />
        <Route path="logs" element={<LiveLogs />} />
      </Route>
    </Routes>
  )
}

function App() {
  return (
    <AuthProvider>
      <Router>
        <AppRoutes />
      </Router>
    </AuthProvider>
  )
}

export default App

