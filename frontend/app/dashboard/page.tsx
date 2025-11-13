'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/lib/auth'
import api from '@/lib/api'
import Layout from '@/components/Layout'
import Link from 'next/link'
import toast from 'react-hot-toast'

interface Campaign {
  id: string
  title: string
  status: string
  tracking_token: string
}

export default function DashboardPage() {
  const { user, loading: authLoading } = useAuth()
  const router = useRouter()
  const [stats, setStats] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [campaigns, setCampaigns] = useState<Campaign[]>([])
  const [shareEmail, setShareEmail] = useState('')
  const [selectedCampaignId, setSelectedCampaignId] = useState('')
  const [sharing, setSharing] = useState(false)

  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login')
    }
  }, [user, authLoading, router])

  useEffect(() => {
    if (user) {
      fetchStats()
      fetchCampaigns()
    }
  }, [user])

  const fetchStats = async () => {
    try {
      const response = await api.get('/api/user/analytics')
      setStats(response.data.stats)
    } catch (error) {
      console.error('Failed to fetch stats:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchCampaigns = async () => {
    try {
      const response = await api.get('/api/user/campaigns')
      setCampaigns(response.data || [])
    } catch (error) {
      console.error('Failed to fetch campaigns:', error)
    }
  }

  const handleShareSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedCampaignId || !shareEmail) {
      toast.error('Please select a campaign and enter an email address')
      return
    }

    // Basic email validation
    if (!shareEmail.includes('@') || !shareEmail.includes('.')) {
      toast.error('Please enter a valid email address')
      return
    }

    setSharing(true)
    try {
      const response = await api.post(`/api/user/campaigns/${selectedCampaignId}/share`, {
        email: shareEmail,
      })
      toast.success(`Campaign link sent to ${shareEmail}`)
      setShareEmail('')
      setSelectedCampaignId('')
    } catch (error: any) {
      console.error('Share campaign error:', error)
      const errorMessage = error.response?.data?.error || error.message || 'Failed to send campaign link'
      toast.error(errorMessage)
      
      // Show more detailed error if available
      if (error.response?.status === 503 || errorMessage.includes('Email service')) {
        toast.error('Email service is not configured. Please contact administrator.')
      }
    } finally {
      setSharing(false)
    }
  }

  const approvedCampaigns = campaigns.filter(c => c.status === 'approved')

  if (authLoading || loading) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      </Layout>
    )
  }

  return (
    <Layout>
      <div className="px-4 py-6">
        <h1 className="text-3xl font-bold text-gray-900 mb-6">Dashboard</h1>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Total Campaigns</h3>
            <p className="text-3xl font-bold text-primary-600">{stats?.total_campaigns || 0}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Approved</h3>
            <p className="text-3xl font-bold text-green-600">{stats?.approved_campaigns || 0}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Total Clicks</h3>
            <p className="text-3xl font-bold text-blue-600">{stats?.total_clicks || 0}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Conversion Rate</h3>
            <p className="text-3xl font-bold text-purple-600">
              {stats?.conversion_rate ? `${stats.conversion_rate.toFixed(1)}%` : '0%'}
            </p>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-semibold">Quick Actions</h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Link
                href="/campaigns/new"
                className="bg-primary-600 text-white px-6 py-3 rounded-lg hover:bg-primary-700 transition-colors text-center"
              >
                Create New Campaign
              </Link>
              <Link
                href="/campaigns"
                className="bg-gray-200 text-gray-800 px-6 py-3 rounded-lg hover:bg-gray-300 transition-colors text-center"
              >
                View My Campaigns
              </Link>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-4">Share Campaign</h2>
            <p className="text-sm text-gray-600 mb-4">
              Send a campaign link to someone via email. Only approved campaigns can be shared.
            </p>
            <form onSubmit={handleShareSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Select Campaign
                </label>
                <select
                  value={selectedCampaignId}
                  onChange={(e) => setSelectedCampaignId(e.target.value)}
                  required
                  className="w-full px-4 py-2 border border-gray-300 text-black rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                >
                  <option value="">-- Select a campaign --</option>
                  {approvedCampaigns.map((campaign) => (
                    <option key={campaign.id} value={campaign.id}>
                      {campaign.title}
                    </option>
                  ))}
                </select>
                {approvedCampaigns.length === 0 && (
                  <p className="text-sm text-gray-500 mt-2">
                    No approved campaigns available. Create and get a campaign approved first.
                  </p>
                )}
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Recipient Email Address
                </label>
                <input
                  type="email"
                  value={shareEmail}
                  onChange={(e) => setShareEmail(e.target.value)}
                  required
                  placeholder="recipient@example.com"
                  className="w-full px-4 py-2 border border-gray-300 text-black rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                />
              </div>
              <button
                type="submit"
                disabled={sharing || approvedCampaigns.length === 0}
                className="w-full bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {sharing ? 'Sending...' : 'Send Campaign Link'}
              </button>
            </form>
          </div>
        </div>
      </div>
    </Layout>
  )
}

