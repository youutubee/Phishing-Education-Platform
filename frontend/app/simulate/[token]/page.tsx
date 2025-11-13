'use client'

import { useEffect, useMemo, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import toast from 'react-hot-toast'

export default function SimulatePage() {
  const params = useParams()
  const router = useRouter()
  const token = params.token as string
  const [campaign, setCampaign] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [formData, setFormData] = useState({
    email: '',
    password: '',
  })
  const apiBaseUrl = useMemo(() => {
    if (typeof window !== 'undefined') {
      const envUrl = process.env.NEXT_PUBLIC_API_URL
      if (envUrl && envUrl.length > 0) {
        return envUrl.replace(/\/$/, '')
      }
    }
    return 'http://localhost:8080'
  }, [])

  useEffect(() => {
    if (!token) return

    const controller = new AbortController()

    const fetchCampaign = async () => {
      try {
        const response = await fetch(`${apiBaseUrl}/api/simulate/${token}`, {
          signal: controller.signal,
        })

        if (!response.ok) {
          const data = await response.json().catch(() => ({}))
          const message = data?.error || 'Campaign not found'
          toast.error(message)
          router.push('/')
          return
        }

        const data = await response.json()
        setCampaign(data)
      } catch (error: any) {
        if (error?.name !== 'AbortError') {
          toast.error('Failed to load campaign')
          router.push('/')
        }
      } finally {
        setLoading(false)
      }
    }

    setLoading(true)
    fetchCampaign()

    return () => controller.abort()
  }, [apiBaseUrl, router, token])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!token) return

    try {
      const response = await fetch(`${apiBaseUrl}/api/simulate/${token}/submit`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData),
      })

      if (!response.ok) {
        const data = await response.json().catch(() => ({}))
        throw new Error(data?.error || 'Failed to submit')
      }

      router.push(`/awareness/${token}`)
    } catch (error: any) {
      toast.error(error?.message || 'Failed to submit')
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  if (!campaign) {
    return null
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        <h1 className="text-2xl font-bold mb-6 text-center">{campaign.title}</h1>
        
        <div className="mb-6 p-4 bg-gray-50 rounded-lg">
          <p className="text-sm text-gray-600 whitespace-pre-wrap">{campaign.landing_url || 'Please enter your credentials to continue'}</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Email
            </label>
            <input
              type="email"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              className="w-full px-4 py-2 border border-gray-300 text-black rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder="your@email.com"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Password
            </label>
            <input
              type="password"
              value={formData.password}
              onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              className="w-full px-4 py-2 border border-gray-300 text-black rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder="••••••••"
            />
          </div>
          <button
            type="submit"
            className="w-full bg-primary-600 text-white py-2 px-4 rounded-lg hover:bg-primary-700 transition-colors"
          >
            Submit
          </button>
        </form>
      </div>
    </div>
  )
}

