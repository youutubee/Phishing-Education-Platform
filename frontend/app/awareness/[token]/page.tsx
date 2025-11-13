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
    <div className="min-h-screen bg-gradient-to-br from-red-50 via-orange-50 to-yellow-50">
      <div className="max-w-4xl mx-auto px-4 py-12">
        <div className="bg-white rounded-lg shadow-xl p-8">
          <div className="text-center mb-8">
            <div className="text-6xl mb-4">⚠️</div>
            <h1 className="text-4xl font-bold text-red-600 mb-2">
              {content?.content?.title || 'You\'ve Been Phished! (Simulated)'}
            </h1>
            <p className="text-xl text-gray-600">
              {content?.content?.description || 'This was a safe, educational simulation designed to teach you about phishing attacks.'}
            </p>
          </div>

          <div className="prose max-w-none mb-8 space-y-6">
            <div className="bg-red-50 border-l-4 border-red-500 p-4 rounded">
              <h2 className="text-2xl font-semibold mb-3 text-red-700">What Just Happened?</h2>
              <p className="text-gray-700 mb-2">
                You just interacted with a <strong>simulated phishing attempt</strong>. This was created as part of an educational campaign
                to help you learn about phishing and social engineering attacks. <strong>No real data was captured or stored.</strong>
              </p>
              <p className="text-gray-700">
                The goal is to help you recognize these tactics in real-world scenarios and protect yourself from actual threats.
              </p>
            </div>

            <div>
              <h2 className="text-2xl font-semibold mb-4 text-gray-900">🔍 How to Recognize Phishing Attempts</h2>
              <div className="grid md:grid-cols-2 gap-4">
                <div className="bg-blue-50 p-4 rounded-lg">
                  <h3 className="font-semibold text-blue-900 mb-2">1. Check the Sender</h3>
                  <p className="text-gray-700 text-sm">
                    Verify the email address carefully. Phishers often use similar-looking domains (e.g., 
                    <code className="bg-gray-200 px-1 rounded">amaz0n.com</code> instead of <code className="bg-gray-200 px-1 rounded">amazon.com</code>).
                  </p>
                </div>
                <div className="bg-blue-50 p-4 rounded-lg">
                  <h3 className="font-semibold text-blue-900 mb-2">2. Inspect URLs</h3>
                  <p className="text-gray-700 text-sm">
                    Hover over links before clicking. Look for misspellings, suspicious domains, or HTTP instead of HTTPS.
                  </p>
                </div>
                <div className="bg-blue-50 p-4 rounded-lg">
                  <h3 className="font-semibold text-blue-900 mb-2">3. Look for Urgency</h3>
                  <p className="text-gray-700 text-sm">
                    Phishing emails often create a sense of urgency ("Act now!" or "Your account will be closed!") 
                    to make you act quickly without thinking.
                  </p>
                </div>
                <div className="bg-blue-50 p-4 rounded-lg">
                  <h3 className="font-semibold text-blue-900 mb-2">4. Check for Errors</h3>
                  <p className="text-gray-700 text-sm">
                    Poor grammar, spelling mistakes, and unprofessional formatting are common red flags in phishing emails.
                  </p>
                </div>
                <div className="bg-blue-50 p-4 rounded-lg">
                  <h3 className="font-semibold text-blue-900 mb-2">5. Verify Requests</h3>
                  <p className="text-gray-700 text-sm">
                    If an email asks for sensitive information, verify it through a separate, trusted channel 
                    (phone call, official website, etc.).
                  </p>
                </div>
                <div className="bg-blue-50 p-4 rounded-lg">
                  <h3 className="font-semibold text-blue-900 mb-2">6. Check Attachments</h3>
                  <p className="text-gray-700 text-sm">
                    Be cautious with unexpected attachments, especially .exe, .zip, or .scr files. 
                    They may contain malware.
                  </p>
                </div>
              </div>
            </div>

            <div className="bg-green-50 border-l-4 border-green-500 p-4 rounded">
              <h2 className="text-2xl font-semibold mb-4 text-green-700">✅ Best Practices for Cybersecurity</h2>
              <ul className="list-disc list-inside space-y-2 text-gray-700">
                <li><strong>Never enter credentials</strong> on suspicious pages or links from emails</li>
                <li><strong>Use two-factor authentication (2FA)</strong> whenever possible for added security</li>
                <li><strong>Keep your software and browsers updated</strong> to patch security vulnerabilities</li>
                <li><strong>Be cautious with email attachments</strong> - scan them before opening</li>
                <li><strong>Report suspicious emails</strong> to your IT department or email provider</li>
                <li><strong>Use strong, unique passwords</strong> for each account and consider a password manager</li>
                <li><strong>Verify the website's SSL certificate</strong> - look for the padlock icon in your browser</li>
                <li><strong>Don't trust caller ID</strong> - phone numbers can be spoofed in vishing attacks</li>
              </ul>
            </div>

            <div className="bg-yellow-50 border-l-4 border-yellow-500 p-4 rounded">
              <h2 className="text-2xl font-semibold mb-3 text-yellow-700">📚 Types of Social Engineering Attacks</h2>
              <div className="space-y-3 text-gray-700">
                <div>
                  <strong className="text-yellow-800">Phishing:</strong> Fraudulent emails pretending to be from legitimate sources
                </div>
                <div>
                  <strong className="text-yellow-800">Spear Phishing:</strong> Targeted phishing attacks using personal information
                </div>
                <div>
                  <strong className="text-yellow-800">Vishing:</strong> Voice phishing via phone calls
                </div>
                <div>
                  <strong className="text-yellow-800">Smishing:</strong> SMS/text message phishing
                </div>
                <div>
                  <strong className="text-yellow-800">Pretexting:</strong> Creating a false scenario to obtain information
                </div>
              </div>
            </div>

            <div className="bg-purple-50 border-l-4 border-purple-500 p-4 rounded">
              <h2 className="text-2xl font-semibold mb-3 text-purple-700">🛡️ What to Do If You Suspect Phishing</h2>
              <ol className="list-decimal list-inside space-y-2 text-gray-700">
                <li><strong>Don't click</strong> any links or download attachments</li>
                <li><strong>Don't reply</strong> to the email or provide any information</li>
                <li><strong>Report it</strong> to your IT security team or email provider</li>
                <li><strong>Delete the email</strong> after reporting</li>
                <li><strong>If you clicked a link,</strong> change your passwords immediately and monitor your accounts</li>
                <li><strong>If you entered credentials,</strong> enable 2FA and contact the service provider</li>
              </ol>
            </div>
          </div>

          <div className="text-center mt-8 pt-6 border-t">
            <p className="text-gray-600 mb-4">
              Remember: Legitimate organizations will never ask for sensitive information via email or suspicious links.
            </p>
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
