'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import api from '@/lib/api'
import Link from 'next/link'

export default function AwarenessPage() {
  const params = useParams()
  const token = params.token as string
  const [content, setContent] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchContent()
  }, [token])

  const fetchContent = async () => {
    try {
      const response = await api.get(`/api/awareness/${token}`)
      setContent(response.data)
    } catch (error) {
      console.error('Failed to fetch awareness content:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-red-50 to-orange-50">
      <div className="max-w-4xl mx-auto px-4 py-12">
        <div className="bg-white rounded-lg shadow-xl p-8">
          <div className="text-center mb-8">
            <div className="text-6xl mb-4">⚠️</div>
            <h1 className="text-4xl font-bold text-red-600 mb-2">
              {content?.content?.title || 'You\'ve Been Phished! (Simulated)'}
            </h1>
            <p className="text-xl text-gray-600">
              {content?.content?.description || 'This was a safe, educational simulation'}
            </p>
          </div>

          <div className="prose max-w-none mb-8">
            <h2 className="text-2xl font-semibold mb-4">What Just Happened?</h2>
            <p className="text-gray-700 mb-4">
              You just interacted with a simulated phishing attempt. This was created as part of an educational campaign
              to help you learn about phishing and social engineering attacks. No real data was captured or stored.
            </p>

            <h2 className="text-2xl font-semibold mb-4">How to Recognize Phishing Attempts</h2>
            <ul className="list-disc list-inside space-y-2 text-gray-700">
              <li><strong>Check the sender:</strong> Verify the email address carefully. Phishers often use similar-looking domains.</li>
              <li><strong>Inspect URLs:</strong> Hover over links before clicking. Look for misspellings or suspicious domains.</li>
              <li><strong>Look for urgency:</strong> Phishing emails often create a sense of urgency to make you act quickly.</li>
              <li><strong>Check for errors:</strong> Poor grammar and spelling mistakes are common in phishing emails.</li>
              <li><strong>Verify requests:</strong> If an email asks for sensitive information, verify it through a separate channel.</li>
            </ul>

            <h2 className="text-2xl font-semibold mb-4 mt-8">Best Practices</h2>
            <div className="bg-blue-50 border-l-4 border-blue-500 p-4 mb-4">
              <ul className="list-disc list-inside space-y-2 text-gray-700">
                <li>Never enter credentials on suspicious pages</li>
                <li>Use two-factor authentication whenever possible</li>
                <li>Keep your software and browsers updated</li>
                <li>Be cautious with email attachments</li>
                <li>Report suspicious emails to your IT department</li>
              </ul>
            </div>
          </div>

          <div className="text-center">
            <Link
              href="/"
              className="inline-block bg-primary-600 text-white px-6 py-3 rounded-lg hover:bg-primary-700 transition-colors"
            >
              Return to Home
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}

