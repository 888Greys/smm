'use client'

import { useState, useEffect } from 'react'
import { X, Link2, Phone, CheckCircle, AlertCircle, Loader2, ExternalLink } from 'lucide-react'
import { Package } from '@/lib/types'
import { createOrder, getOrderStatus } from '@/lib/api'
import { useRouter } from 'next/navigation'

type Step = 'link' | 'phone' | 'waiting' | 'success' | 'error'

interface Props {
  pkg: Package
  onClose: () => void
}

export default function OrderModal({ pkg, onClose }: Props) {
  const router = useRouter()
  const [step, setStep] = useState<Step>('link')
  const [profileLink, setProfileLink] = useState('')
  const [phone, setPhone] = useState('')
  const [referralCode, setReferralCode] = useState('')
  const [orderId, setOrderId] = useState<number | null>(null)
  const [errorMsg, setErrorMsg] = useState('')
  const [loading, setLoading] = useState(false)

  // Poll order status when waiting
  useEffect(() => {
    if (step !== 'waiting' || !orderId) return

    const interval = setInterval(async () => {
      try {
        const status = await getOrderStatus(orderId)
        if (status.status === 'processing' || status.status === 'completed') {
          setStep('success')
          clearInterval(interval)
          // Fire confetti
          import('canvas-confetti').then(({ default: confetti }) => {
            confetti({ particleCount: 140, spread: 80, origin: { y: 0.55 }, colors: ['#7c3aed', '#06b6d4', '#a78bfa', '#38bdf8', '#ffffff'] })
          })
        } else if (status.status === 'failed' || status.status === 'cancelled') {
          setErrorMsg('Payment was not completed. Please try again.')
          setStep('error')
          clearInterval(interval)
        }
      } catch {}
    }, 5000)

    // Timeout after 5 minutes
    const timeout = setTimeout(() => {
      clearInterval(interval)
      if (step === 'waiting') {
        setErrorMsg('Payment confirmation timed out. If you paid, contact support.')
        setStep('error')
      }
    }, 300_000)

    return () => { clearInterval(interval); clearTimeout(timeout) }
  }, [step, orderId])

  // Close on Escape
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  const handleSubmitLink = () => {
    if (!profileLink.startsWith('http')) {
      setErrorMsg('Please paste a valid URL starting with https://')
      return
    }
    setErrorMsg('')
    setStep('phone')
  }

  const handleSubmitPhone = async () => {
    if (phone.length < 9) {
      setErrorMsg('Enter a valid Safaricom number (e.g. 0712345678)')
      return
    }
    setLoading(true)
    setErrorMsg('')

    try {
      const result = await createOrder({
        package_id: pkg.id,
        profile_link: profileLink,
        phone,
        referral_code: referralCode || undefined,
      })

      if (result.error) {
        setErrorMsg(result.error)
        setLoading(false)
        return
      }

      setOrderId(result.order_id)
      setStep('waiting')
    } catch (e) {
      setErrorMsg('Network error. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" onClick={onClose} />

      {/* Modal */}
      <div className="relative w-full max-w-md glass rounded-2xl p-6 shadow-2xl animate-fade-up">
        {/* Close */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-slate-500 hover:text-white transition-colors"
        >
          <X size={20} />
        </button>

        {/* Package summary */}
        <div className="mb-6 pb-5 border-b border-border-dim">
          <p className="text-xs text-violet-400 font-semibold uppercase tracking-wider mb-1">Selected Package</p>
          <h2 className="text-xl font-bold text-white">{pkg.name}</h2>
          <p className="text-slate-400 text-sm mt-0.5">{pkg.description}</p>
          <div className="mt-3 flex items-center justify-between">
            <span className="text-2xl font-black text-white">KES {pkg.price_kes.toLocaleString()}</span>
            <span className="text-xs text-green-400 bg-green-400/10 border border-green-400/20 px-2.5 py-1 rounded-full">
              Instant delivery
            </span>
          </div>
        </div>

        {/* Step: Link */}
        {step === 'link' && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">
                Your Profile / Post Link
              </label>
              <div className="relative">
                <Link2 size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                <input
                  type="url"
                  placeholder="https://tiktok.com/@yourprofile"
                  value={profileLink}
                  onChange={e => setProfileLink(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && handleSubmitLink()}
                  className="w-full bg-surface-2 border border-border-dim rounded-xl pl-9 pr-4 py-3 text-white placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors text-sm"
                  autoFocus
                />
              </div>
              <p className="text-xs text-slate-500 mt-1.5">Make sure your profile is public</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">
                Referral Code <span className="text-slate-600">(optional)</span>
              </label>
              <input
                type="text"
                placeholder="Enter code if you have one"
                value={referralCode}
                onChange={e => setReferralCode(e.target.value.toUpperCase())}
                className="w-full bg-surface-2 border border-border-dim rounded-xl px-4 py-3 text-white placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors text-sm"
              />
            </div>

            {errorMsg && <p className="text-red-400 text-sm">{errorMsg}</p>}

            <button
              onClick={handleSubmitLink}
              className="w-full py-3.5 rounded-xl font-bold text-white bg-gradient-brand hover:opacity-90 transition-opacity"
            >
              Continue →
            </button>
          </div>
        )}

        {/* Step: Phone */}
        {step === 'phone' && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">
                M-Pesa Phone Number
              </label>
              <div className="relative">
                <Phone size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                <input
                  type="tel"
                  placeholder="0712 345 678"
                  value={phone}
                  onChange={e => setPhone(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && !loading && handleSubmitPhone()}
                  className="w-full bg-surface-2 border border-border-dim rounded-xl pl-9 pr-4 py-3 text-white placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors text-sm"
                  autoFocus
                />
              </div>
              <p className="text-xs text-slate-500 mt-1.5">
                You&apos;ll receive an M-Pesa prompt on this number
              </p>
            </div>

            {errorMsg && <p className="text-red-400 text-sm">{errorMsg}</p>}

            <button
              onClick={handleSubmitPhone}
              disabled={loading}
              className="w-full py-3.5 rounded-xl font-bold text-white bg-gradient-brand hover:opacity-90 disabled:opacity-50 transition-all flex items-center justify-center gap-2"
            >
              {loading ? (
                <><Loader2 size={18} className="animate-spin" /> Sending M-Pesa request...</>
              ) : (
                <>Pay KES {pkg.price_kes.toLocaleString()} with M-Pesa</>
              )}
            </button>

            <button
              onClick={() => { setStep('link'); setErrorMsg('') }}
              className="w-full text-sm text-slate-500 hover:text-slate-300 transition-colors"
            >
              ← Back
            </button>
          </div>
        )}

        {/* Step: Waiting */}
        {step === 'waiting' && (
          <div className="text-center py-4 space-y-5">
            <div className="w-16 h-16 mx-auto rounded-full bg-violet-500/10 border border-violet-500/30 flex items-center justify-center">
              <Loader2 size={28} className="text-violet-400 animate-spin" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-white mb-2">Check Your Phone</h3>
              <p className="text-slate-400 text-sm leading-relaxed">
                An M-Pesa prompt has been sent to <strong className="text-white">{phone}</strong>.
                Enter your PIN to complete the payment.
              </p>
            </div>
            <div className="glass rounded-xl p-4 text-sm text-slate-400">
              Order #{orderId} · Waiting for confirmation...
            </div>
            <p className="text-xs text-slate-600">
              This page updates automatically. Do not close it.
            </p>
          </div>
        )}

        {/* Step: Success */}
        {step === 'success' && (
          <div className="text-center py-4 space-y-5">
            <div className="w-16 h-16 mx-auto rounded-full bg-green-500/10 border border-green-500/30 flex items-center justify-center">
              <CheckCircle size={28} className="text-green-400" />
            </div>
            <div>
              <h3 className="text-xl font-bold text-white mb-2">Payment Confirmed! 🎉</h3>
              <p className="text-slate-400 text-sm leading-relaxed">
                Your order is being processed. Followers will start arriving within minutes.
              </p>
            </div>
            <div className="flex flex-col gap-2">
              <button
                onClick={() => router.push(`/status/${orderId}`)}
                className="flex items-center justify-center gap-2 w-full py-3 rounded-xl font-semibold text-white bg-gradient-brand hover:opacity-90 transition-opacity"
              >
                <ExternalLink size={16} />
                Track Your Order
              </button>
              <button
                onClick={onClose}
                className="w-full py-3 text-sm text-slate-400 hover:text-white transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        )}

        {/* Step: Error */}
        {step === 'error' && (
          <div className="text-center py-4 space-y-5">
            <div className="w-16 h-16 mx-auto rounded-full bg-red-500/10 border border-red-500/30 flex items-center justify-center">
              <AlertCircle size={28} className="text-red-400" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-white mb-2">Something Went Wrong</h3>
              <p className="text-slate-400 text-sm">{errorMsg}</p>
            </div>
            <div className="flex flex-col gap-2">
              <button
                onClick={() => { setStep('phone'); setErrorMsg('') }}
                className="w-full py-3 rounded-xl font-semibold text-white bg-gradient-brand hover:opacity-90 transition-opacity"
              >
                Try Again
              </button>
              <a
                href="https://t.me/workratew"
                target="_blank"
                rel="noopener noreferrer"
                className="w-full py-3 text-sm text-slate-400 hover:text-violet-300 transition-colors"
              >
                Contact Support →
              </a>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
