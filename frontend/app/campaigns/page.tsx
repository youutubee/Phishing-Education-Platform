'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/lib/auth'
import api from '@/lib/api'
import Layout from '@/components/Layout'
import Link from 'next/link'
import toast from 'react-hot-toast'

interface Campaign {
  id: number
  title: string
  description: string
  status: string
  tracking_token: string
  created_at: string
  admin_comment?: string
}

export default function CampaignsPage() {
  const { user, loading: authLoading } = useAuth()
  const router = useRouter()
  const [campaigns, setCampaigns] = useState<Campaign[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login')
    }
  }, [user, authLoading, router])

  useEffect(() => {
    if (user) {
      fetchCampaigns()
    }
  }, [user])

  const fetchCampaigns = async () => {
    try {
      const response = await api.get('/api/user/campaigns')
      setCampaigns(response.data)
    } catch (error) {
      toast.error('Failed to fetch campaigns')
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this campaign?')) return

    try {
      await api.delete(`/api/user/campaigns/${id}`)
      toast.success('Campaign deleted')
      fetchCampaigns()
    } catch (error) {
      toast.error('Failed to delete campaign')
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
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold text-gray-900">My Campaigns</h1>
          <Link
            href="/campaigns/new"
            className="bg-primary-600 text-white px-6 py-2 rounded-lg hover:bg-primary-700 transition-colors"
          >
            Create Campaign
          </Link>
        </div>

        {campaigns.length === 0 ? (
          <div className="bg-white text-black rounded-lg shadow p-12 text-center">
            <p className="text-black mb-4">No campaigns yet</p>
            <Link
              href="/campaigns/new"
              className="text-primary-600 hover:text-primary-700 font-medium"
            >
              Create your first campaign
            </Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4">
            {campaigns.map((campaign) => (
              <div key={campaign.id} className="bg-white text-black rounded-lg shadow p-6">
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <h3 className="text-xl font-semibold">{campaign.title}</h3>
                      <span className={`px-2 py-1 text-black rounded text-sm font-medium ${getStatusColor(campaign.status)}`}>
                        {campaign.status}
                      </span>
                    </div>
                    {campaign.description && (
                      <p className="text-black mb-2">{campaign.description}</p>
                    )}
                    {campaign.admin_comment && (
                      <p className="text-sm text-black mb-2">
                        <strong>Admin comment:</strong> {campaign.admin_comment}
                      </p>
                    )}
                    <div className="flex items-center gap-4 text-sm text-black">
                      <span>Token: {campaign.tracking_token.substring(0, 20)}...</span>
                      <span>Created: {new Date(campaign.created_at).toLocaleDateString()}</span>
                    </div>
                    {campaign.status === 'approved' && (
                      <div className="mt-3">
                        <p className="text-sm text-black mb-1">Simulation Link:</p>
                        <code className="bg-gray-100 text-black px-2 py-1 rounded text-sm">
                          {typeof window !== 'undefined' && `${window.location.origin}/simulate/${campaign.tracking_token}`}
                        </code>
                      </div>
                    )}
                  </div>
                  <div className="flex gap-2">
                    <Link
                      href={`/campaigns/${campaign.id}/edit`}
                      className="text-primary-600 hover:text-primary-700 text-sm"
                    >
                      Edit
                    </Link>
                    <button
                      onClick={() => handleDelete(campaign.id)}
                      className="text-red-600 hover:text-red-700 text-sm"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </Layout>
  )
}

