'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/lib/auth'
import api from '@/lib/api'
import Layout from '@/components/Layout'
import toast from 'react-hot-toast'

export default function NewCampaignPage() {
  const { user, loading: authLoading } = useAuth()
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    email_text: '',
    landing_page_url: '',
    expiry_date: '',
  })

  if (authLoading) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      </Layout>
    )
  }

  if (!user) {
    router.push('/login')
    return null
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      const payload = {
        ...formData,
        expiry_date: formData.expiry_date ? new Date(formData.expiry_date).toISOString() : null,
      }
      await api.post('/api/user/campaigns', payload)
      toast.success('Campaign created successfully!')
      router.push('/campaigns')
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to create campaign')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Layout>
      <div className="px-4 py-6 max-w-3xl">
        <h1 className="text-3xl font-bold text-gray-900 mb-6">Create New Campaign</h1>

        <form onSubmit={handleSubmit} className="bg-white rounded-lg shadow p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Title *
            </label>
            <input
              type="text"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              required
              className="w-full px-4 py-2 border border-gray-300 text-black rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder="Phishing Campaign Title"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Description
            </label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              rows={3}
              className="w-full px-4 py-2 border border-gray-300 text-black rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder="Campaign description"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Email Text (Simulated Email Content) *
            </label>
            <textarea
              value={formData.email_text}
              onChange={(e) => setFormData({ ...formData, email_text: e.target.value })}
              required
              rows={6}
              className="w-full px-4 py-2 border text-black border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent font-mono text-sm"
              placeholder="Enter the email content that will be shown in the simulation..."
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Landing Page URL
            </label>
            <input
              type="url"
              value={formData.landing_page_url}
              onChange={(e) => setFormData({ ...formData, landing_page_url: e.target.value })}
              className="w-full px-4 py-2 border border-gray-300 text-black rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder="https://example.com/phishing-page"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Expiry Date
            </label>
            <input
              type="datetime-local"
              value={formData.expiry_date}
              onChange={(e) => setFormData({ ...formData, expiry_date: e.target.value })}
              className="w-full px-4 py-2 border border-gray-300 text-black rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
          </div>

          <div className="flex gap-4">
            <button
              type="submit"
              disabled={loading}
              className="bg-primary-600 text-white px-6 py-2 rounded-lg hover:bg-primary-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Creating...' : 'Create Campaign'}
            </button>
            <button
              type="button"
              onClick={() => router.back()}
              className="bg-gray-200 text-gray-800 px-6 py-2 rounded-lg hover:bg-gray-300 transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </Layout>
  )
}

