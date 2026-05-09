'use client'

import { useEffect, useState } from 'react'
import { getOrderStatus } from '@/lib/api'
import { OrderStatus } from '@/lib/types'
import { CheckCircle, Clock, Loader2, XCircle, RefreshCw } from 'lucide-react'
import Link from 'next/link'

const statusConfig = {
  pending:    { label: 'Awaiting Payment',    icon: Clock,         color: 'text-yellow-400',  bg: 'bg-yellow-500/10 border-yellow-500/30' },
  processing: { label: 'Delivering Now',      icon: Loader2,       color: 'text-violet-400',  bg: 'bg-violet-500/10 border-violet-500/30', spin: true },
  completed:  { label: 'Delivered!',          icon: CheckCircle,   color: 'text-green-400',   bg: 'bg-green-500/10 border-green-500/30' },
  failed:     { label: 'Failed',              icon: XCircle,       color: 'text-red-400',     bg: 'bg-red-500/10 border-red-500/30' },
  cancelled:  { label: 'Cancelled',           icon: XCircle,       color: 'text-slate-400',   bg: 'bg-slate-500/10 border-slate-500/30' },
}

export default function StatusTracker({ orderId }: { orderId: number }) {
  const [order, setOrder] = useState<OrderStatus | null>(null)
  const [error, setError] = useState(false)

  const fetchStatus = async () => {
    try {
      const data = await getOrderStatus(orderId)
      if (data.error) { setError(true); return }
      setOrder(data)
    } catch {
      setError(true)
    }
  }

  useEffect(() => {
    fetchStatus()
    // Poll every 15s for active orders
    const interval = setInterval(() => {
      if (order?.status === 'completed' || order?.status === 'failed' || order?.status === 'cancelled') {
        clearInterval(interval)
        return
      }
      fetchStatus()
    }, 15_000)
    return () => clearInterval(interval)
  }, [orderId, order?.status])

  if (error) {
    return (
      <div className="glass rounded-2xl p-8 max-w-md w-full text-center space-y-4">
        <XCircle size={40} className="text-red-400 mx-auto" />
        <p className="text-white font-bold">Order not found</p>
        <p className="text-slate-400 text-sm">Order #{orderId} could not be loaded.</p>
        <Link href="/" className="text-violet-400 hover:text-violet-300 text-sm">← Back to shop</Link>
      </div>
    )
  }

  if (!order) {
    return (
      <div className="glass rounded-2xl p-8 max-w-md w-full text-center">
        <Loader2 size={32} className="text-violet-400 animate-spin mx-auto mb-4" />
        <p className="text-slate-400 text-sm">Loading order status...</p>
      </div>
    )
  }

  const cfg = statusConfig[order.status] ?? statusConfig.pending
  const Icon = cfg.icon

  const steps = [
    { label: 'Order Created',    done: true },
    { label: 'Payment Confirmed', done: ['processing', 'completed'].includes(order.status) },
    { label: 'Delivering',       done: ['processing', 'completed'].includes(order.status) },
    { label: 'Complete',         done: order.status === 'completed' },
  ]

  return (
    <div className="glass rounded-2xl p-8 max-w-md w-full space-y-6">
      {/* Header */}
      <div>
        <p className="text-xs text-slate-500 mb-1">Order #{order.order_id}</p>
        <h2 className="text-xl font-bold text-white">{order.package_name}</h2>
        <p className="text-slate-400 text-sm mt-0.5">{order.description}</p>
        <p className="text-violet-300 font-bold mt-1">KES {order.price_kes?.toLocaleString()}</p>
      </div>

      {/* Status badge */}
      <div className={`flex items-center gap-3 rounded-xl border p-4 ${cfg.bg}`}>
        <Icon size={24} className={`${cfg.color} ${(cfg as any).spin ? 'animate-spin' : ''}`} />
        <div>
          <p className={`font-bold ${cfg.color}`}>{cfg.label}</p>
          <p className="text-xs text-slate-500">
            {order.status === 'processing' ? 'Followers are on their way to your profile' :
             order.status === 'completed'  ? 'Check your profile — delivery is done!' :
             order.status === 'pending'    ? 'Waiting for M-Pesa payment confirmation' :
             'Something went wrong — contact support'}
          </p>
        </div>
      </div>

      {/* Progress steps */}
      <div className="space-y-3">
        {steps.map((s, i) => (
          <div key={i} className="flex items-center gap-3">
            <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 ${
              s.done ? 'bg-violet-600 border-violet-600' : 'border-slate-600'
            }`}>
              {s.done && <div className="w-2 h-2 rounded-full bg-white" />}
            </div>
            <span className={`text-sm ${s.done ? 'text-white' : 'text-slate-600'}`}>{s.label}</span>
          </div>
        ))}
      </div>

      {/* Actions */}
      <div className="flex flex-col gap-2 pt-2 border-t border-border-dim">
        {order.status !== 'completed' && order.status !== 'cancelled' && order.status !== 'failed' && (
          <button
            onClick={fetchStatus}
            className="flex items-center justify-center gap-2 text-sm text-slate-400 hover:text-white transition-colors"
          >
            <RefreshCw size={14} /> Refresh status
          </button>
        )}
        <Link href="/" className="text-center text-sm text-violet-400 hover:text-violet-300 transition-colors">
          ← Order more packages
        </Link>
      </div>
    </div>
  )
}
