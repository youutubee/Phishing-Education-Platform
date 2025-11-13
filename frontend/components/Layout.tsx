'use client'

import { useRouter } from 'next/navigation'
import { useAuth } from '@/lib/auth'
import Link from 'next/link'

export default function Layout({ children }: { children: React.ReactNode }) {
  const { user, logout } = useAuth()
  const router = useRouter()

  const handleLogout = () => {
    logout()
    router.push('/login')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex">
              <Link href="/dashboard" className="flex items-center px-2 py-2 text-xl font-bold text-primary-700">
                SEAP
              </Link>
              <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
                <Link href="/dashboard" className="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-900 hover:text-primary-600">
                  Dashboard
                </Link>
                {user?.role === 'admin' && (
                  <>
                    <Link href="/admin/campaigns" className="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-900 hover:text-primary-600">
                      Admin Panel
                    </Link>
                    <Link href="/admin/users" className="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-900 hover:text-primary-600">
                      Users
                    </Link>
                    <Link href="/admin/analytics" className="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-900 hover:text-primary-600">
                      Analytics
                    </Link>
                    <Link href="/admin/leaderboard" className="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-900 hover:text-primary-600">
                      Leaderboard
                    </Link>
                    <Link href="/admin/audit-logs" className="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-900 hover:text-primary-600">
                      Audit Logs
                    </Link>
                  </>
                )}
                <Link href="/campaigns" className="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-900 hover:text-primary-600">
                  My Campaigns
                </Link>
                <Link href="/analytics" className="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-900 hover:text-primary-600">
                  Analytics
                </Link>
                <Link href="/profile" className="inline-flex items-center px-1 pt-1 text-sm font-medium text-gray-900 hover:text-primary-600">
                  Profile
                </Link>
              </div>
            </div>
            <div className="flex items-center">
              <span className="text-sm text-black mr-4">{user ? `${user.email} (${user.role})` : ''}</span>
              <button
                onClick={handleLogout}
                className="bg-primary-600 text-white px-4 py-2 rounded-lg hover:bg-primary-700 transition-colors text-sm"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        {children}
      </main>
    </div>
  )
}

