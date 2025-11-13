'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import Layout from '@/components/Layout'
import { useAuth } from '@/lib/auth'
import api from '@/lib/api'
import LeaderboardTable, { LeaderboardEntry } from '@/components/LeaderboardTable'

export default function LeaderboardPage() {
  const { user, loading: authLoading } = useAuth()
  const router = useRouter()
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login')
    }
  }, [authLoading, router, user])

  useEffect(() => {
    if (user) {
      fetchLeaderboard()
    }
  }, [user])

  const fetchLeaderboard = async () => {
    try {
      const response = await api.get('/api/leaderboard')
      setLeaderboard(response.data)
    } catch (error) {
      console.error('Failed to fetch leaderboard:', error)
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

  return (
    <Layout>
      <div className="px-4 py-6">
        <h1 className="text-3xl font-bold text-gray-900 mb-6">Leaderboard</h1>
        <p className="text-gray-600 mb-6">
          Top campaign creators ranked by engagement, conversions, and campaign quality.
        </p>

        <LeaderboardTable leaderboard={leaderboard} />

        {leaderboard.length === 0 && (
          <div className="bg-white rounded-lg shadow p-12 text-center mt-6">
            <p className="text-gray-500">No leaderboard data available yet</p>
          </div>
        )}
      </div>
    </Layout>
  )
}


