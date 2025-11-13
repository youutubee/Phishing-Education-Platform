'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/lib/auth'
import api from '@/lib/api'
import Layout from '@/components/Layout'

interface AuditLog {
  id: string
  admin_id: string
  admin_email: string
  action: string
  resource_type: string
  resource_id: string | null
  details: string
  created_at: string
}

export default function AuditLogsPage() {
  const { user, loading: authLoading } = useAuth()
  const router = useRouter()
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!authLoading && (!user || user.role !== 'admin')) {
      router.push('/dashboard')
    }
  }, [user, authLoading, router])

  useEffect(() => {
    if (user && user.role === 'admin') {
      fetchLogs()
    }
  }, [user])

  const fetchLogs = async () => {
    try {
      const response = await api.get('/api/admin/audit-logs')
      setLogs(response.data)
    } catch (error) {
      console.error('Failed to fetch audit logs:', error)
    } finally {
      setLoading(false)
    }
  }

  const getActionColor = (action: string) => {
    if (action.includes('approve')) return 'text-green-600'
    if (action.includes('reject')) return 'text-red-600'
    if (action.includes('delete')) return 'text-orange-600'
    return 'text-gray-600'
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
        <h1 className="text-3xl font-bold text-gray-900 mb-6">Audit Logs</h1>
        <p className="text-gray-600 mb-6">
          Complete history of all admin actions for accountability and security.
        </p>

        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date & Time</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Admin</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Action</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Resource</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Details</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {logs.map((log) => (
                <tr key={log.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(log.created_at).toLocaleString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-black font-medium">
                    {log.admin_email}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    <span className={`font-medium ${getActionColor(log.action)}`}>
                      {log.action.replace('_', ' ').toUpperCase()}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-black">
                    {log.resource_type}
                    {log.resource_id && (
                      <span className="text-gray-500 ml-1">({log.resource_id.substring(0, 8)}...)</span>
                    )}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600 max-w-md">
                    {log.details ? (
                      <pre className="whitespace-pre-wrap text-xs bg-gray-50 p-2 rounded">
                        {log.details}
                      </pre>
                    ) : (
                      <span className="text-gray-400">-</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {logs.length === 0 && (
          <div className="bg-white rounded-lg shadow p-12 text-center">
            <p className="text-gray-500">No audit logs available yet</p>
          </div>
        )}
      </div>
    </Layout>
  )
}

