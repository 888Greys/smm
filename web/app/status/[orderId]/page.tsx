import StatusTracker from '@/components/StatusTracker'
import Navbar from '@/components/Navbar'
import Link from 'next/link'

export default function StatusPage({ params }: { params: { orderId: string } }) {
  const orderId = parseInt(params.orderId)

  if (isNaN(orderId)) {
    return (
      <main className="min-h-screen bg-dark flex flex-col">
        <Navbar />
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center">
            <p className="text-slate-400 mb-4">Invalid order ID.</p>
            <Link href="/" className="text-violet-400 hover:text-violet-300">← Back home</Link>
          </div>
        </div>
      </main>
    )
  }

  return (
    <main className="min-h-screen bg-dark flex flex-col">
      <Navbar />
      <div className="flex-1 flex items-center justify-center px-4 py-20">
        <StatusTracker orderId={orderId} />
      </div>
    </main>
  )
}
