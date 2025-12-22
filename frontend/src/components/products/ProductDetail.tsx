import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { productsAPI } from '../../services/api'

type ProductStatus = 'pending' | 'approved' | 'rejected'

interface Product {
  id: number
  elastic_id: string
  product_id?: string
  item_id?: string
  name?: string
  brand?: string
  catalog?: string
  artist?: string
  rating?: number
  list_price?: number
  sale_price?: number
  club_price?: number
  sale_rank?: number
  customer_interest_index?: number
  in_stock: boolean
  is_active: boolean
  is_buyable: boolean
  licensed: boolean
  is_applique: boolean
  is_cross_stitch: boolean
  is_pdf_available: boolean
  is_fsl: boolean
  is_heat_transfer: boolean
  is_design_used_in_project: boolean
  in_custom_pack: boolean
  definition_name?: string
  product_type?: string
  gtin?: string
  color_sequence?: string
  design_keywords?: string
  categories?: string
  categories_list?: string
  keywords?: string
  sales?: string
  sales_list?: string
  sale_end_date?: string
  year_created?: string
  applied_discount_id?: number
  is_multiple_variants_available: boolean
  variants?: string
  raw_data?: string
  created_at: string
  updated_at: string
  status?: ProductStatus
}

export default function ProductDetail() {
  const { id } = useParams<{ id: string }>()
  const [product, setProduct] = useState<Product | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [statusUpdating, setStatusUpdating] = useState(false)
  const [statusFeedback, setStatusFeedback] = useState<
    { type: 'success' | 'error'; message: string } | null
  >(null)
  const statusOptions: ProductStatus[] = ['pending', 'approved', 'rejected']

  useEffect(() => {
    if (id) {
      loadProduct()
    }
  }, [id])

  const loadProduct = async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await productsAPI.get(Number(id))
      setProduct(response.data)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load product')
    } finally {
      setLoading(false)
    }
  }

  const parseJSON = (str?: string) => {
    if (!str) return null
    try {
      return JSON.parse(str)
    } catch {
      return str
    }
  }

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '—'
    try {
      return new Date(dateStr).toLocaleString()
    } catch {
      return dateStr
    }
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

  const handleStatusUpdate = async (newStatus: ProductStatus) => {
    if (!product || product.status === newStatus) {
      return
    }

    setStatusUpdating(true)
    setStatusFeedback(null)
    try {
      const response = await productsAPI.updateStatus(product.id, newStatus)
      setProduct(response.data)
      setStatusFeedback({
        type: 'success',
        message: `Status updated to ${newStatus}`,
      })
    } catch (err: any) {
      setStatusFeedback({
        type: 'error',
        message: err.response?.data?.error || 'Failed to update status',
      })
    } finally {
      setStatusUpdating(false)
    }
  }

  if (loading) {
    return (
      <div className="py-12 text-center text-gray-500">Loading product details…</div>
    )
  }

  if (error) {
    return (
      <div className="space-y-4">
        <div className="rounded-lg bg-red-50 p-4 text-red-700 border border-red-100">
          {error}
        </div>
        <Link
          to="/products"
          className="inline-block rounded-lg border border-gray-200 px-4 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50"
        >
          ← Back to Products
        </Link>
      </div>
    )
  }

  if (!product) {
    return (
      <div className="space-y-4">
        <div className="rounded-lg bg-yellow-50 p-4 text-yellow-700 border border-yellow-100">
          Product not found
        </div>
        <Link
          to="/products"
          className="inline-block rounded-lg border border-gray-200 px-4 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-50"
        >
          ← Back to Products
        </Link>
      </div>
    )
  }

  const categoriesList = parseJSON(product.categories_list)
  const keywordsList = parseJSON(product.keywords)
  const salesList = parseJSON(product.sales_list)
  const variantsList = parseJSON(product.variants)
  const rawData = parseJSON(product.raw_data)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <Link
            to="/products"
            className="text-sm text-gray-500 hover:text-gray-700 mb-2 inline-block"
          >
            ← Back to Products
          </Link>
          <h1 className="text-3xl font-semibold text-gray-900">
            {product.name || product.catalog || `Product #${product.id}`}
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Elastic ID: <span className="font-mono">{product.elastic_id}</span>
          </p>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Basic Information */}
        <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Basic Information</h2>
          <dl className="space-y-3">
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Product ID
              </dt>
              <dd className="mt-1 font-mono text-sm text-gray-900">
                {product.product_id || '—'}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Item ID
              </dt>
              <dd className="mt-1 font-mono text-sm text-gray-900">
                {product.item_id || '—'}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Name
              </dt>
              <dd className="mt-1 text-sm text-gray-900">{product.name || '—'}</dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Brand
              </dt>
              <dd className="mt-1 text-sm text-gray-900">{product.brand || '—'}</dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Catalog
              </dt>
              <dd className="mt-1 text-sm text-gray-900">{product.catalog || '—'}</dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Artist
              </dt>
              <dd className="mt-1 text-sm text-gray-900">{product.artist || '—'}</dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Product Type
              </dt>
              <dd className="mt-1 text-sm text-gray-900">{product.product_type || '—'}</dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Definition Name
              </dt>
              <dd className="mt-1 text-sm text-gray-900">{product.definition_name || '—'}</dd>
            </div>
          </dl>
        </div>

        {/* Pricing & Status */}
        <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Pricing & Status</h2>
          <dl className="space-y-3">
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                List Price
              </dt>
              <dd className="mt-1 text-sm text-gray-900">
                {product.list_price ? `$${product.list_price.toFixed(2)}` : '—'}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Sale Price
              </dt>
              <dd className="mt-1 text-sm text-gray-900">
                {product.sale_price ? `$${product.sale_price.toFixed(2)}` : '—'}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Club Price
              </dt>
              <dd className="mt-1 text-sm text-gray-900">
                {product.club_price ? `$${product.club_price.toFixed(2)}` : '—'}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Rating
              </dt>
              <dd className="mt-1 text-sm text-gray-900">
                {product.rating ? product.rating.toFixed(1) : '—'}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Moderation Status
              </dt>
              <dd className="mt-1 flex flex-col gap-2 text-sm text-gray-900">
                {renderStatusBadge(product.status)}
                <select
                  value={product.status ?? 'pending'}
                  disabled={statusUpdating}
                  onChange={(e) => handleStatusUpdate(e.target.value as ProductStatus)}
                  className="w-full rounded-lg border border-gray-200 px-3 py-2 text-xs uppercase tracking-wide focus:outline-none focus:ring-2 focus:ring-indigo-200"
                >
                  {statusOptions.map((status) => (
                    <option key={status} value={status}>
                      {status.charAt(0).toUpperCase() + status.slice(1)}
                    </option>
                  ))}
                </select>
                {statusFeedback && (
                  <span
                    className={`text-xs ${
                      statusFeedback.type === 'success' ? 'text-green-600' : 'text-red-600'
                    }`}
                  >
                    {statusFeedback.message}
                  </span>
                )}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Sale Rank
              </dt>
              <dd className="mt-1 text-sm text-gray-900">
                {product.sale_rank ?? '—'}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Customer Interest Index
              </dt>
              <dd className="mt-1 text-sm text-gray-900">
                {product.customer_interest_index ?? '—'}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Stock Status
              </dt>
              <dd className="mt-1">
                <span
                  className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${
                    product.in_stock
                      ? 'bg-green-100 text-green-700'
                      : 'bg-red-100 text-red-700'
                  }`}
                >
                  {product.in_stock ? 'In Stock' : 'Out of Stock'}
                </span>
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Status Flags
              </dt>
              <dd className="mt-1 flex flex-wrap gap-2">
                {product.is_active && (
                  <span className="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold bg-green-100 text-green-700">
                    Active
                  </span>
                )}
                {product.is_buyable && (
                  <span className="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold bg-blue-100 text-blue-700">
                    Buyable
                  </span>
                )}
                {product.licensed && (
                  <span className="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold bg-purple-100 text-purple-700">
                    Licensed
                  </span>
                )}
              </dd>
            </div>
          </dl>
        </div>

        {/* Product Features */}
        <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Product Features</h2>
          <dl className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                  Applique
                </dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {product.is_applique ? 'Yes' : 'No'}
                </dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                  Cross Stitch
                </dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {product.is_cross_stitch ? 'Yes' : 'No'}
                </dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                  PDF Available
                </dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {product.is_pdf_available ? 'Yes' : 'No'}
                </dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                  FSL
                </dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {product.is_fsl ? 'Yes' : 'No'}
                </dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                  Heat Transfer
                </dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {product.is_heat_transfer ? 'Yes' : 'No'}
                </dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                  Used in Project
                </dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {product.is_design_used_in_project ? 'Yes' : 'No'}
                </dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                  In Custom Pack
                </dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {product.in_custom_pack ? 'Yes' : 'No'}
                </dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                  Multiple Variants
                </dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {product.is_multiple_variants_available ? 'Yes' : 'No'}
                </dd>
              </div>
            </div>
          </dl>
        </div>

        {/* Additional Information */}
        <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Additional Information</h2>
          <dl className="space-y-3">
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                GTIN
              </dt>
              <dd className="mt-1 font-mono text-sm text-gray-900">{product.gtin || '—'}</dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Color Sequence
              </dt>
              <dd className="mt-1 text-sm text-gray-900">{product.color_sequence || '—'}</dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Design Keywords
              </dt>
              <dd className="mt-1 text-sm text-gray-900">{product.design_keywords || '—'}</dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Categories
              </dt>
              <dd className="mt-1 text-sm text-gray-900">{product.categories || '—'}</dd>
            </div>
            {categoriesList && Array.isArray(categoriesList) && (
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                  Categories List
                </dt>
                <dd className="mt-1">
                  <ul className="list-disc list-inside text-sm text-gray-900 space-y-1">
                    {categoriesList.map((cat: string, idx: number) => (
                      <li key={idx}>{cat}</li>
                    ))}
                  </ul>
                </dd>
              </div>
            )}
            {keywordsList && Array.isArray(keywordsList) && (
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                  Keywords
                </dt>
                <dd className="mt-1">
                  <div className="flex flex-wrap gap-2">
                    {keywordsList.map((kw: string, idx: number) => (
                      <span
                        key={idx}
                        className="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-gray-100 text-gray-700"
                      >
                        {kw}
                      </span>
                    ))}
                  </div>
                </dd>
              </div>
            )}
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Sales
              </dt>
              <dd className="mt-1 text-sm text-gray-900">{product.sales || '—'}</dd>
            </div>
            {salesList && Array.isArray(salesList) && (
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                  Sales List
                </dt>
                <dd className="mt-1">
                  <ul className="list-disc list-inside text-sm text-gray-900 space-y-1">
                    {salesList.map((sale: string, idx: number) => (
                      <li key={idx}>{sale}</li>
                    ))}
                  </ul>
                </dd>
              </div>
            )}
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Sale End Date
              </dt>
              <dd className="mt-1 text-sm text-gray-900">
                {formatDate(product.sale_end_date)}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                Year Created
              </dt>
              <dd className="mt-1 text-sm text-gray-900">
                {formatDate(product.year_created)}
              </dd>
            </div>
            {product.applied_discount_id && (
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                  Applied Discount ID
                </dt>
                <dd className="mt-1 text-sm text-gray-900">{product.applied_discount_id}</dd>
              </div>
            )}
          </dl>
        </div>
      </div>

      {/* Variants */}
      {variantsList && Array.isArray(variantsList) && variantsList.length > 0 && (
        <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Variants</h2>
          <div className="space-y-2">
            {variantsList.map((variant: any, idx: number) => (
              <div
                key={idx}
                className="rounded-lg border border-gray-200 p-3 text-sm"
              >
                <pre className="whitespace-pre-wrap text-xs text-gray-700">
                  {JSON.stringify(variant, null, 2)}
                </pre>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Raw Data */}
      {rawData && (
        <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Raw Data (JSON)</h2>
          <div className="rounded-lg border border-gray-200 bg-gray-50 p-4 overflow-x-auto">
            <pre className="text-xs text-gray-700 whitespace-pre-wrap">
              {JSON.stringify(rawData, null, 2)}
            </pre>
          </div>
        </div>
      )}

      {/* Metadata */}
      <div className="rounded-2xl border border-gray-100 bg-white p-6 shadow-sm">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Metadata</h2>
        <dl className="space-y-3">
          <div>
            <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
              Created At
            </dt>
            <dd className="mt-1 text-sm text-gray-900">{formatDate(product.created_at)}</dd>
          </div>
          <div>
            <dt className="text-xs font-semibold uppercase tracking-wider text-gray-500">
              Updated At
            </dt>
            <dd className="mt-1 text-sm text-gray-900">{formatDate(product.updated_at)}</dd>
          </div>
        </dl>
      </div>
    </div>
  )
}
