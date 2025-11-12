'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/lib/auth'
import api from '@/lib/api'
import Layout from '@/components/Layout'
import toast from 'react-hot-toast'

interface Campaign {
  id: number
  title: string
  description: string
  status: string
  user_email: string
  created_at: string
  admin_comment?: string
}

export default function AdminCampaignsPage() {
  const { user, loading: authLoading } = useAuth()
  const router = useRouter()
  const [campaigns, setCampaigns] = useState<Campaign[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedCampaign, setSelectedCampaign] = useState<number | null>(null)
  const [comment, setComment] = useState('')
  const [action, setAction] = useState<'approve' | 'reject' | null>(null)

  useEffect(() => {
    if (!authLoading && (!user || user.role !== 'admin')) {
      router.push('/dashboard')
    }
  }, [user, authLoading, router])

  useEffect(() => {
    if (user && user.role === 'admin') {
      fetchCampaigns()
    }
  }, [user])

  const fetchCampaigns = async () => {
    try {
      const response = await api.get('/api/admin/campaigns')
      setCampaigns(response.data)
    } catch (error) {
      toast.error('Failed to fetch campaigns')
    } finally {
      setLoading(false)
    }
  }

  const handleApprove = async (id: number) => {
    setSelectedCampaign(id)
    setAction('approve')
    setComment('')
  }

  const handleReject = async (id: number) => {
    setSelectedCampaign(id)
    setAction('reject')
    setComment('')
  }

  const submitAction = async () => {
    if (!selectedCampaign || !action) return

    try {
      if (action === 'approve') {
        await api.post(`/api/admin/campaigns/${selectedCampaign}/approve`, { comment })
      } else {
        if (!comment.trim()) {
          toast.error('Comment is required for rejection')
          return
        }
        await api.post(`/api/admin/campaigns/${selectedCampaign}/reject`, { comment })
      }
      toast.success(`Campaign ${action === 'approve' ? 'approved' : 'rejected'} successfully`)
      setSelectedCampaign(null)
      setAction(null)
      setComment('')
      fetchCampaigns()
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update campaign')
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'approved':
        return 'bg-green-100 text-green-800'
      case 'rejected':
        return 'bg-red-100 text-red-800'
      default:
        return 'bg-yellow-100 text-yellow-800'
    }
  }

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
        <h1 className="text-3xl font-bold text-gray-900 mb-6">Admin - Campaign Management</h1>

        {selectedCampaign && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
            <div className="bg-white rounded-lg shadow-xl p-6 max-w-md w-full">
              <h2 className="text-xl font-semibold mb-4">
                {action === 'approve' ? 'Approve Campaign' : 'Reject Campaign'}
              </h2>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Comment {action === 'reject' && '*'}
                </label>
                <textarea
                  value={comment}
                  onChange={(e) => setComment(e.target.value)}
                  rows={4}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  placeholder={action === 'approve' ? 'Optional comment...' : 'Reason for rejection (required)'}
                />
              </div>
              <div className="flex gap-4">
                <button
                  onClick={submitAction}
                  className={`flex-1 px-4 py-2 rounded-lg text-white transition-colors ${
                    action === 'approve'
                      ? 'bg-green-600 hover:bg-green-700'
                      : 'bg-red-600 hover:bg-red-700'
                  }`}
                >
                  {action === 'approve' ? 'Approve' : 'Reject'}
                </button>
                <button
                  onClick={() => {
                    setSelectedCampaign(null)
                    setAction(null)
                    setComment('')
                  }}
                  className="flex-1 px-4 py-2 rounded-lg bg-gray-200 text-gray-800 hover:bg-gray-300 transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 gap-4">
          {campaigns.map((campaign) => (
            <div key={campaign.id} className="bg-white rounded-lg shadow p-6">
              <div className="flex justify-between items-start">
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-2">
                    <h3 className="text-xl font-semibold">{campaign.title}</h3>
                    <span className={`px-2 py-1 rounded text-sm font-medium ${getStatusColor(campaign.status)}`}>
                      {campaign.status}
                    </span>
                  </div>
                  {campaign.description && (
                    <p className="text-gray-600 mb-2">{campaign.description}</p>
                  )}
                  {campaign.admin_comment && (
                    <p className="text-sm text-gray-500 mb-2">
                      <strong>Admin comment:</strong> {campaign.admin_comment}
                    </p>
                  )}
                  <div className="text-sm text-gray-500">
                    <p>Created by: {campaign.user_email}</p>
                    <p>Created: {new Date(campaign.created_at).toLocaleDateString()}</p>
                  </div>
                </div>
                {campaign.status === 'pending' && (
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleApprove(campaign.id)}
                      className="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition-colors text-sm"
                    >
                      Approve
                    </button>
                    <button
                      onClick={() => handleReject(campaign.id)}
                      className="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors text-sm"
                    >
                      Reject
                    </button>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    </Layout>
  )
}

