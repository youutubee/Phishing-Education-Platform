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

export default function AnalyticsPage() {
  const { user, loading: authLoading } = useAuth()
  const router = useRouter()
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login')
    }
  }, [user, authLoading, router])

  useEffect(() => {
    if (user) {
      fetchAnalytics()
    }
  }, [user])

  const fetchAnalytics = async () => {
    try {
      const response = await api.get('/api/user/analytics')
      setData(response.data)
    } catch (error) {
      console.error('Failed to fetch analytics:', error)
    } finally {
      setLoading(false)
    }
  }

  if (authLoading || loading || !data) {
    return (
      <Layout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
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

  const campaignData = {
    labels: data.campaigns?.map((c: any) => c.title) || [],
    datasets: [
      {
        label: 'Clicks',
        data: data.campaigns?.map((c: any) => c.clicks) || [],
        backgroundColor: 'rgba(14, 165, 233, 0.5)',
      },
      {
        label: 'Conversions',
        data: data.campaigns?.map((c: any) => c.awareness_views) || [],
        backgroundColor: 'rgba(34, 197, 94, 0.5)',
      },
    ],
  }

  return (
    <Layout>
      <div className="px-4 py-6">
        <h1 className="text-3xl font-bold text-gray-900 mb-6">Analytics</h1>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Total Campaigns</h3>
            <p className="text-3xl font-bold text-primary-600">{data.stats?.total_campaigns || 0}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Total Clicks</h3>
            <p className="text-3xl font-bold text-blue-600">{data.stats?.total_clicks || 0}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Total Conversions</h3>
            <p className="text-3xl font-bold text-green-600">{data.stats?.total_awareness_views || 0}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Conversion Rate</h3>
            <p className="text-3xl font-bold text-purple-600">
              {data.stats?.conversion_rate ? `${data.stats.conversion_rate.toFixed(1)}%` : '0%'}
            </p>
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
            <h2 className="text-xl font-semibold mb-4">Campaign Performance</h2>
            {data.campaigns && data.campaigns.length > 0 ? (
              <Bar data={campaignData} />
            ) : (
              <p className="text-gray-500 text-center py-8">No data available</p>
            )}
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Campaign Details</h2>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Campaign</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Clicks</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Submissions</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Awareness Views</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {data.campaigns?.map((campaign: any) => (
                  <tr key={campaign.id}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">{campaign.title}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <span className={`px-2 py-1 rounded ${
                        campaign.status === 'approved' ? 'bg-green-100 text-green-800' :
                        campaign.status === 'rejected' ? 'bg-red-100 text-red-800' :
                        'bg-yellow-100 text-yellow-800'
                      }`}>
                        {campaign.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">{campaign.clicks}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">{campaign.submissions}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">{campaign.awareness_views}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </Layout>
  )
}

