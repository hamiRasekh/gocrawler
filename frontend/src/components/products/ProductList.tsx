import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { productsAPI } from '../../services/api'

type ProductStatus = 'pending' | 'approved' | 'rejected'

interface Product {
  id: number
  name?: string
  brand?: string
  catalog?: string
  list_price?: number
  sale_price?: number
  in_stock: boolean
  rating?: number
  status?: ProductStatus
}

interface ProductStats {
  total?: number
  total_products?: number
  in_stock?: number
  brands_count?: number
  status_breakdown?: Record<string, number>
}

export default function ProductList() {
  const [products, setProducts] = useState<Product[]>([])
  const [stats, setStats] = useState<ProductStats>({})
  const [loading, setLoading] = useState(true)
  const [filters, setFilters] = useState({ search: '', brand: '', inStock: 'all', status: 'pending' })
  const [limit, setLimit] = useState(25)
  const [offset, setOffset] = useState(0)
  const [total, setTotal] = useState(0)
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; message: string } | null>(
    null
  )
  const [crawling, setCrawling] = useState(false)

  useEffect(() => {
    loadProducts()
    loadStats()
  }, [limit, offset])

  const loadProducts = async () => {
    setLoading(true)
    try {
      const response = await productsAPI.list({
        limit,
        offset,
        search: filters.search || undefined,
        brand: filters.brand || undefined,
        in_stock:
          filters.inStock === 'all' ? undefined : filters.inStock === 'true' ? true : false,
        status: filters.status === 'all' ? undefined : filters.status,
      })
      setProducts(response.data.products || [])
      setTotal(response.data.total || 0)
    } catch (error) {
      console.error('Failed to load products', error)
      setFeedback({ type: 'error', message: 'Failed to load products' })
    } finally {
      setLoading(false)
    }
  }

  const loadStats = async () => {
    try {
      const response = await productsAPI.getStats()
      setStats(response.data || {})
    } catch (error) {
      console.error('Failed to load product stats', error)
    }
  }

  const handleSearch = () => {
    setOffset(0)
    loadProducts()
  }

  const totalPages = useMemo(() => Math.max(1, Math.ceil(total / limit)), [total, limit])
  const currentPage = Math.floor(offset / limit) + 1

  const changePage = (page: number) => {
    const clamped = Math.min(Math.max(page, 1), totalPages)
    setOffset((clamped - 1) * limit)
  }

  const renderStatusBadge = (status?: ProductStatus) => {
    const normalized = status ?? 'pending'
    const labels: Record<ProductStatus, string> = {
      pending: 'Pending',
      approved: 'Approved',
      rejected: 'Rejected',
    }
    const tones: Record<ProductStatus, string> = {
      pending: 'bg-amber-100 text-amber-800',
      approved: 'bg-emerald-100 text-emerald-700',
      rejected: 'bg-rose-100 text-rose-700',
    }

    return (
      <span
        className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${
          tones[normalized]
        }`}
      >
        {labels[normalized]}
      </span>
    )
  }

  const handleStartCrawl = async () => {
    setCrawling(true)
    try {
      await productsAPI.startCrawl()
      setFeedback({ type: 'success', message: 'Embroidery crawl kicked off' })
    } catch (error) {
      console.error('Failed to start crawl', error)
      setFeedback({ type: 'error', message: 'Failed to trigger crawl' })
    } finally {
      setCrawling(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h1 className="text-3xl font-semibold text-gray-900">Product Catalog</h1>
          <p className="text-sm text-gray-500">
            Browse and validate designs pulled from embroiderydesigns.com
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <button
            onClick={handleStartCrawl}
            disabled={crawling}
            className="rounded-lg border border-gray-200 px-3 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50 disabled:opacity-50"
          >
            {crawling ? 'Starting crawl…' : 'Start embroidery crawl'}
          </button>
        </div>
      </div>

      {feedback && (
        <div
          className={`rounded-lg p-4 text-sm ${
            feedback.type === 'success'
              ? 'bg-green-50 text-green-700 border border-green-100'
              : 'bg-red-50 text-red-700 border border-red-100'
          }`}
        >
          <div className="flex items-center justify-between">
            <span>{feedback.message}</span>
            <button className="text-xs" onClick={() => setFeedback(null)}>
              Dismiss
            </button>
          </div>
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
        <ProductSummary
          label="Tracked products"
          value={stats.total_products ?? stats.total ?? total}
        />
        <ProductSummary
          label="Pending review"
          value={stats.status_breakdown?.pending ?? 0}
          tone="amber"
        />
        <ProductSummary label="In stock" value={stats.in_stock ?? 0} tone="green" />
        <ProductSummary label="Brands" value={stats.brands_count ?? 0} tone="purple" />
        <ProductSummary label="Page size" value={limit} tone="slate" />
      </div>

      <div className="rounded-2xl border border-gray-100 bg-white p-4 shadow-sm">
        <div className="grid gap-4 md:grid-cols-5">
          <div>
            <label className="text-xs font-semibold uppercase tracking-wider text-gray-500">
              Search
            </label>
            <input
              type="search"
              placeholder="Name, catalog, elastic id…"
              value={filters.search}
              onChange={(e) => setFilters((prev) => ({ ...prev, search: e.target.value }))}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
            />
          </div>
          <div>
            <label className="text-xs font-semibold uppercase tracking-wider text-gray-500">
              Brand
            </label>
            <input
              type="text"
              placeholder="Brand filter"
              value={filters.brand}
              onChange={(e) => setFilters((prev) => ({ ...prev, brand: e.target.value }))}
              className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
            />
          </div>
          <div>
            <label className="text-xs font-semibold uppercase tracking-wider text-gray-500">
              Stock
            </label>
            <select
              value={filters.inStock}
              onChange={(e) => setFilters((prev) => ({ ...prev, inStock: e.target.value }))}
              className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
            >
              <option value="all">All</option>
              <option value="true">In stock</option>
              <option value="false">Out of stock</option>
            </select>
          </div>
          <div>
            <label className="text-xs font-semibold uppercase tracking-wider text-gray-500">
              Status
            </label>
            <select
              value={filters.status}
              onChange={(e) => setFilters((prev) => ({ ...prev, status: e.target.value }))}
              className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-200"
            >
              <option value="all">All</option>
              <option value="pending">Pending</option>
              <option value="approved">Approved</option>
              <option value="rejected">Rejected</option>
            </select>
          </div>
          <div className="flex items-end gap-2">
            <button
              onClick={handleSearch}
              className="w-full rounded-lg bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
            >
              Apply filters
            </button>
            <button
              onClick={() => {
                setFilters({ search: '', brand: '', inStock: 'all', status: 'pending' })
                setOffset(0)
                loadProducts()
              }}
              className="rounded-lg border border-gray-200 px-3 py-2 text-xs font-semibold text-gray-600 hover:bg-gray-50"
            >
              Reset
            </button>
          </div>
        </div>
      </div>

      <div className="rounded-2xl border border-gray-100 bg-white shadow-sm">
        {loading ? (
          <div className="py-12 text-center text-gray-500">Loading products…</div>
        ) : products.length === 0 ? (
          <div className="py-12 text-center text-gray-500">No products match the filters.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-100 text-sm">
              <thead className="bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                <tr>
                  <th className="px-4 py-3">ID</th>
                  <th className="px-4 py-3">Name</th>
                  <th className="px-4 py-3">Brand</th>
                  <th className="px-4 py-3">Price</th>
                  <th className="px-4 py-3">Rating</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Stock</th>
                  <th className="px-4 py-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 text-gray-700">
                {products.map((product) => (
                  <tr key={product.id}>
                    <td className="px-4 py-3 font-mono text-xs text-gray-500">{product.id}</td>
                    <td className="px-4 py-3 font-medium text-gray-900">
                      {product.name || product.catalog || 'Unnamed'}
                    </td>
                    <td className="px-4 py-3">{product.brand || '—'}</td>
                    <td className="px-4 py-3">
                      {product.sale_price
                        ? `$${product.sale_price.toFixed(2)}`
                        : product.list_price
                        ? `$${product.list_price.toFixed(2)}`
                        : '—'}
                    </td>
                    <td className="px-4 py-3">{product.rating ? product.rating.toFixed(1) : '—'}</td>
                    <td className="px-4 py-3">{renderStatusBadge(product.status)}</td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${
                          product.in_stock ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
                        }`}
                      >
                        {product.in_stock ? 'In stock' : 'Out of stock'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <Link
                        to={`/products/${product.id}`}
                        className="rounded-lg border border-gray-200 px-3 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50"
                      >
                        View
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        {products.length > 0 && (
          <div className="flex flex-col items-center justify-between gap-3 border-t border-gray-100 px-4 py-3 text-sm text-gray-600 md:flex-row">
            <div>
              Showing {offset + 1}–
              {Math.min(offset + limit, total)} of {total}
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => changePage(currentPage - 1)}
                className="rounded-lg border border-gray-200 px-3 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50"
                disabled={currentPage === 1}
              >
                Prev
              </button>
              <span className="text-xs text-gray-500">
                Page {currentPage} / {totalPages}
              </span>
              <button
                onClick={() => changePage(currentPage + 1)}
                className="rounded-lg border border-gray-200 px-3 py-1 text-xs font-semibold text-gray-700 hover:bg-gray-50"
                disabled={currentPage === totalPages}
              >
                Next
              </button>
              <select
                value={limit}
                onChange={(e) => {
                  setLimit(Number(e.target.value))
                  setOffset(0)
                }}
                className="rounded-lg border border-gray-200 px-2 py-1 text-xs focus:outline-none"
              >
                {[25, 50, 100].map((size) => (
                  <option key={size} value={size}>
                    {size}/page
                  </option>
                ))}
              </select>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function ProductSummary({
  label,
  value,
  tone = 'gray',
}: {
  label: string
  value: number
  tone?: 'gray' | 'green' | 'purple' | 'slate' | 'amber'
}) {
  const toneStyles: Record<string, string> = {
    gray: 'bg-white border-gray-100 text-gray-900',
    green: 'bg-green-50 border-green-100 text-green-800',
    purple: 'bg-purple-50 border-purple-100 text-purple-800',
    slate: 'bg-slate-50 border-slate-100 text-slate-800',
    amber: 'bg-amber-50 border-amber-100 text-amber-800',
  }

  return (
    <div className={`rounded-2xl border p-4 shadow-sm ${toneStyles[tone]}`}>
      <p className="text-xs font-semibold uppercase tracking-wider text-gray-500">{label}</p>
      <p className="mt-2 text-2xl font-semibold">{value}</p>
    </div>
  )
}
