'use client'

export interface LeaderboardEntry {
  user_id: string
  email: string
  total_campaigns: number
  total_clicks: number
  total_conversions: number
  rejected_count: number
  score: number
}

interface LeaderboardTableProps {
  leaderboard: LeaderboardEntry[]
}

const getRankBadge = (index: number) => {
  if (index === 0) return '🥇'
  if (index === 1) return '🥈'
  if (index === 2) return '🥉'
  return `#${index + 1}`
}

export default function LeaderboardTable({ leaderboard }: LeaderboardTableProps) {
  return (
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
              <td className="px-6 py-4 whitespace-nowrap text-black text-sm font-medium">{entry.email}</td>
              <td className="px-6 py-4 whitespace-nowrap text-black text-sm">{entry.total_campaigns}</td>
              <td className="px-6 py-4 whitespace-nowrap text-black text-sm">{entry.total_clicks}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-green-600">{entry.total_conversions}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-red-600">{entry.rejected_count}</td>
              <td className="px-6 py-4 whitespace-nowrap text-sm font-bold text-primary-600">{entry.score}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}


