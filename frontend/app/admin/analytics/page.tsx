'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/lib/auth'
import api from '@/lib/api'
import Layout from '@/components/Layout'
import { Line, Bar, Doughnut } from 'react-chartjs-2'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend
)

export default function AdminAnalyticsPage() {
  const { user, loading: authLoading } = useAuth()
  const router = useRouter()
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!authLoading && (!user || user.role !== 'admin')) {
      router.push('/dashboard')
    }
  }, [user, authLoading, router])

  useEffect(() => {
    if (user && user.role === 'admin') {
      fetchAnalytics()
    }
  }, [user])

  const fetchAnalytics = async () => {
    try {
      setError(null)
      const response = await api.get('/api/admin/analytics')
      setData(response.data)
    } catch (error: any) {
      console.error('Failed to fetch analytics:', error)
      setError(error.response?.data?.error || 'Failed to load analytics data')
    } finally {
      setLoading(false)
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

  if (error) {
    return (
      <Layout>
        <div className="px-4 py-6">
          <div className="bg-red-50 border border-red-200 rounded-lg p-6">
            <h2 className="text-xl font-semibold text-red-800 mb-2">Error Loading Analytics</h2>
            <p className="text-red-600">{error}</p>
            <button
              onClick={fetchAnalytics}
              className="mt-4 bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors"
            >
              Retry
            </button>
          </div>
        </div>
      </Layout>
    )
  }

  if (!data) {
    return (
      <Layout>
        <div className="px-4 py-6">
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6">
            <p className="text-yellow-800">No analytics data available</p>
          </div>
        </div>
      </Layout>
    )
  }

  const timelineData = {
    labels: data.timeline?.map((t: any) => new Date(t.date).toLocaleDateString()) || [],
    datasets: [
      {
        label: 'Events',
        data: data.timeline?.map((t: any) => t.count) || [],
        borderColor: 'rgb(14, 165, 233)',
        backgroundColor: 'rgba(14, 165, 233, 0.1)',
      },
    ],
  }

  const distributionData = {
    labels: data.distribution?.map((d: any) => d.status) || [],
    datasets: [
      {
        data: data.distribution?.map((d: any) => d.count) || [],
        backgroundColor: [
          'rgba(34, 197, 94, 0.5)',
          'rgba(234, 179, 8, 0.5)',
          'rgba(239, 68, 68, 0.5)',
        ],
      },
    ],
  }

  return (
    <Layout>
      <div className="px-4 py-6">
        <h1 className="text-3xl font-bold text-gray-900 mb-6">Admin Analytics</h1>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Total Users</h3>
            <p className="text-3xl font-bold text-primary-600">{data.stats?.total_users || 0}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Total Campaigns</h3>
            <p className="text-3xl font-bold text-blue-600">{data.stats?.total_campaigns || 0}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Total Clicks</h3>
            <p className="text-3xl font-bold text-green-600">{data.stats?.total_clicks || 0}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Total Conversions</h3>
            <p className="text-3xl font-bold text-purple-600">{data.stats?.total_conversions || 0}</p>
          </div>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Total Events</h3>
            <p className="text-3xl font-bold text-indigo-600">{data.stats?.total_events || 0}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Avg Conversion Rate</h3>
            <p className="text-3xl font-bold text-orange-600">
              {data.stats?.average_conversion_rate ? `${data.stats.average_conversion_rate.toFixed(1)}%` : '0%'}
            </p>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Approved Campaigns</h3>
            <p className="text-3xl font-bold text-emerald-600">{data.stats?.approved_campaigns || 0}</p>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-4">Event Timeline (Last 30 Days)</h2>
            {data.timeline && data.timeline.length > 0 ? (
              <Line data={timelineData} />
            ) : (
              <p className="text-gray-500 text-center py-8">No data available</p>
            )}
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-4">Campaign Status Distribution</h2>
            {data.distribution && data.distribution.length > 0 ? (
              <Doughnut data={distributionData} />
            ) : (
              <p className="text-gray-500 text-center py-8">No data available</p>
            )}
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Platform Statistics</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <p className="text-sm text-gray-500">Approved Campaigns</p>
              <p className="text-2xl font-bold text-green-600">{data.stats?.approved_campaigns || 0}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500">Pending Campaigns</p>
              <p className="text-2xl font-bold text-yellow-600">{data.stats?.pending_campaigns || 0}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500">Rejected Campaigns</p>
              <p className="text-2xl font-bold text-red-600">{data.stats?.rejected_campaigns || 0}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500">Total Clicks</p>
              <p className="text-2xl font-bold text-blue-600">{data.stats?.total_clicks || 0}</p>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  )
}

