'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/lib/auth'
import api from '@/lib/api'
import Layout from '@/components/Layout'

interface LeaderboardEntry {
  user_id: number
  email: string
  total_campaigns: number
  total_clicks: number
  total_conversions: number
  rejected_count: number
  score: number
}

export default function LeaderboardPage() {
  const { user, loading: authLoading } = useAuth()
  const router = useRouter()
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!authLoading && (!user || user.role !== 'admin')) {
      router.push('/dashboard')
    }
  }, [user, authLoading, router])

  useEffect(() => {
    if (user && user.role === 'admin') {
      fetchLeaderboard()
    }
  }, [user])

  const fetchLeaderboard = async () => {
    try {
      const response = await api.get('/api/admin/leaderboard')
      setLeaderboard(response.data)
    } catch (error) {
      console.error('Failed to fetch leaderboard:', error)
    } finally {
      setLoading(false)
    }
  }

  const getRankBadge = (index: number) => {
    if (index === 0) return '🥇'
    if (index === 1) return '🥈'
    if (index === 2) return '🥉'
    return `#${index + 1}`
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
        <h1 className="text-3xl font-bold text-gray-900 mb-6">Leaderboard</h1>
        <p className="text-gray-600 mb-6">
          Top campaign creators ranked by engagement, conversions, and campaign quality.
        </p>

        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Rank</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">User</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Campaigns</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Clicks</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Conversions</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Rejections</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Score</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {leaderboard.map((entry, index) => (
                <tr key={entry.user_id} className={index < 3 ? 'bg-yellow-50' : ''}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-bold">
                    {getRankBadge(index)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">{entry.email}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">{entry.total_campaigns}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">{entry.total_clicks}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-green-600">{entry.total_conversions}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-red-600">{entry.rejected_count}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-bold text-primary-600">{entry.score}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {leaderboard.length === 0 && (
          <div className="bg-white rounded-lg shadow p-12 text-center">
            <p className="text-gray-500">No leaderboard data available yet</p>
          </div>
        )}
      </div>
    </Layout>
  )
}

