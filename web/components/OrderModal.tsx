'use client'

import { useState, useEffect } from 'react'
import { X, Phone, CheckCircle, AlertCircle, Loader2, ExternalLink, User, ChevronRight } from 'lucide-react'
import { SiTiktok, SiInstagram, SiYoutube } from 'react-icons/si'
import { Package, ProfileInfo } from '@/lib/types'
import { createOrder, getOrderStatus, lookupProfile } from '@/lib/api'

type Step = 'username' | 'verify' | 'phone' | 'waiting' | 'success' | 'error'

interface Props {
  pkg: Package
  onClose: () => void
}

const platformIcon = {
  tiktok:    { Icon: SiTiktok,    bg: 'bg-black',                                                       label: 'TikTok',    placeholder: 'e.g. yourhandle' },
  instagram: { Icon: SiInstagram, bg: 'bg-gradient-to-br from-[#833ab4] via-[#fd1d1d] to-[#fcb045]',   label: 'Instagram', placeholder: 'e.g. yourusername' },
  youtube:   { Icon: SiYoutube,   bg: 'bg-[#FF0000]',                                                   label: 'YouTube',   placeholder: 'e.g. yourchannelname' },
}

function PlatformIcon({ platform }: { platform: string }) {
  const cfg = platformIcon[platform as keyof typeof platformIcon] || platformIcon.tiktok
  const { Icon, bg } = cfg
  return (
    <div className={`w-8 h-8 rounded-lg ${bg} flex items-center justify-center flex-shrink-0`}>
      <Icon size={14} className="text-white" />
    </div>
  )
}

export default function OrderModal({ pkg, onClose }: Props) {
  const [step, setStep] = useState<Step>('username')
  const [username, setUsername] = useState('')
  const [profile, setProfile] = useState<ProfileInfo | null>(null)
  const [phone, setPhone] = useState('')
  const [referralCode, setReferralCode] = useState('')
  const [orderId, setOrderId] = useState<number | null>(null)
  const [errorMsg, setErrorMsg] = useState('')
  const [loading, setLoading] = useState(false)

  const cfg = platformIcon[pkg.platform as keyof typeof platformIcon] || platformIcon.tiktok

  // Poll order status when waiting
  useEffect(() => {
    if (step !== 'waiting' || !orderId) return
    const interval = setInterval(async () => {
      try {
        const status = await getOrderStatus(orderId)
        if (status.status === 'processing' || status.status === 'completed') {
          setStep('success')
          clearInterval(interval)
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

  // ── Handlers ──────────────────────────────────────────────────────────────

  const handleUsernameSubmit = async () => {
    const clean = username.trim().replace(/^@/, '')
    if (!clean || clean.length < 2) {
      setErrorMsg('Enter a valid username')
      return
    }
    setLoading(true)
    setErrorMsg('')
    try {
      const info = await lookupProfile(pkg.platform, clean)
      setProfile(info)
      setStep('verify')
    } catch {
      // Network error — still allow proceeding with a basic profile
      setProfile({
        platform: pkg.platform,
        username: clean,
        name: clean,
        bio: '',
        followers: '',
        profile_url: buildProfileURL(pkg.platform, clean),
        avatar_url: '',
        found: false,
      })
      setStep('verify')
    } finally {
      setLoading(false)
    }
  }

  const handlePhoneSubmit = async () => {
    if (phone.length < 9) {
      setErrorMsg('Enter a valid Safaricom number (e.g. 0712345678)')
      return
    }
    if (!profile) return
    setLoading(true)
    setErrorMsg('')
    try {
      const result = await createOrder({
        package_id: pkg.id,
        profile_link: profile.profile_url,
        phone,
        referral_code: referralCode || undefined,
      })
      if (result.error) { setErrorMsg(result.error); setLoading(false); return }
      setOrderId(result.order_id)
      setStep('waiting')
    } catch {
      setErrorMsg('Network error. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  // ── Render ─────────────────────────────────────────────────────────────────

  return (
    <div className="fixed inset-0 z-50 flex items-end sm:items-center justify-center p-0 sm:p-4">
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" onClick={onClose} />

      <div className="relative w-full sm:max-w-md glass sm:rounded-2xl rounded-t-2xl p-5 sm:p-6 shadow-2xl animate-fade-up max-h-[92vh] overflow-y-auto">
        {/* Close */}
        <button onClick={onClose} className="absolute top-4 right-4 text-slate-500 hover:text-white transition-colors z-10">
          <X size={20} />
        </button>

        {/* Package summary */}
        <div className="mb-5 pb-4 border-b border-white/5">
          <div className="flex items-center gap-2 mb-2">
            <PlatformIcon platform={pkg.platform} />
            <div>
              <p className="text-xs text-violet-400 font-semibold uppercase tracking-wider">Selected Package</p>
              <h2 className="text-lg font-bold text-white leading-tight">{pkg.name}</h2>
            </div>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-xl font-black text-white">KES {pkg.price_kes.toLocaleString()}</span>
            <span className="text-xs text-green-400 bg-green-400/10 border border-green-400/20 px-2.5 py-1 rounded-full">Instant delivery</span>
          </div>
        </div>

        {/* ── Step: Username ───────────────────────────────────────────── */}
        {step === 'username' && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-semibold text-slate-200 mb-2">
                Your {cfg.label} Username
              </label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 font-bold text-sm">@</span>
                <input
                  type="text"
                  placeholder={cfg.placeholder}
                  value={username}
                  onChange={e => setUsername(e.target.value.replace(/^@/, ''))}
                  onKeyDown={e => e.key === 'Enter' && !loading && handleUsernameSubmit()}
                  className="w-full bg-surface-2 border border-white/8 rounded-xl pl-8 pr-4 py-3 text-white placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors text-sm"
                  autoFocus
                />
              </div>
              <p className="text-xs text-slate-500 mt-1.5">Just the username — no need to paste a link</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-400 mb-2">
                Referral Code <span className="text-slate-600">(optional)</span>
              </label>
              <input
                type="text"
                placeholder="Enter code if you have one"
                value={referralCode}
                onChange={e => setReferralCode(e.target.value.toUpperCase())}
                className="w-full bg-surface-2 border border-white/8 rounded-xl px-4 py-3 text-white placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors text-sm"
              />
            </div>

            {errorMsg && <p className="text-red-400 text-sm">{errorMsg}</p>}

            <button
              onClick={handleUsernameSubmit}
              disabled={loading}
              className="w-full py-3.5 rounded-xl font-bold text-white bg-gradient-brand hover:opacity-90 disabled:opacity-50 transition-all flex items-center justify-center gap-2"
            >
              {loading ? <><Loader2 size={18} className="animate-spin" /> Looking up account…</> : <>Find Account <ChevronRight size={18} /></>}
            </button>
          </div>
        )}

        {/* ── Step: Verify ─────────────────────────────────────────────── */}
        {step === 'verify' && profile && (
          <div className="space-y-4">
            {profile.found ? (
              <>
                <div className="text-center mb-2">
                  <span className="inline-flex items-center gap-1.5 text-xs text-green-400 bg-green-400/10 border border-green-400/20 px-3 py-1 rounded-full font-semibold">
                    <span className="w-1.5 h-1.5 rounded-full bg-green-400" />
                    Account Found
                  </span>
                </div>

                {/* Profile card */}
                <div className="glass rounded-xl p-4 flex items-start gap-3 border border-white/8">
                  {profile.avatar_url ? (
                    <img
                      src={profile.avatar_url}
                      alt={profile.name}
                      className="w-14 h-14 rounded-full object-cover flex-shrink-0 border-2 border-white/10"
                      onError={e => { (e.target as HTMLImageElement).style.display = 'none' }}
                    />
                  ) : (
                    <div className="w-14 h-14 rounded-full bg-violet-900/50 border border-violet-500/20 flex items-center justify-center flex-shrink-0">
                      <User size={24} className="text-violet-400" />
                    </div>
                  )}
                  <div className="min-w-0 flex-1">
                    <p className="font-bold text-white text-base truncate">{profile.name}</p>
                    <p className="text-slate-400 text-xs">@{profile.username}</p>
                    {profile.followers && (
                      <p className="text-violet-300 text-xs font-semibold mt-1">{profile.followers}</p>
                    )}
                    {profile.bio && (
                      <p className="text-slate-500 text-xs mt-1.5 leading-relaxed line-clamp-2">{profile.bio}</p>
                    )}
                  </div>
                </div>

                <p className="text-center text-sm text-slate-300 font-medium">Is this your account?</p>

                <button
                  onClick={() => { setErrorMsg(''); setStep('phone') }}
                  className="w-full py-3.5 rounded-xl font-bold text-white bg-gradient-brand hover:opacity-90 transition-all flex items-center justify-center gap-2"
                >
                  <CheckCircle size={18} /> Yes, this is me — Continue
                </button>
                <button
                  onClick={() => { setStep('username'); setProfile(null); setErrorMsg('') }}
                  className="w-full py-2 text-sm text-slate-500 hover:text-slate-300 transition-colors"
                >
                  ← Try a different username
                </button>
              </>
            ) : (
              /* Account not found / private / lookup failed */
              <>
                <div className="glass rounded-xl p-4 border border-yellow-500/20 bg-yellow-500/5">
                  <p className="text-yellow-300 text-sm font-semibold mb-1">⚠️ Could not verify account</p>
                  <p className="text-slate-400 text-xs leading-relaxed">
                    Your profile may be private or the platform is temporarily blocking previews.
                    Your order will be placed at: <span className="text-violet-300 break-all">{profile.profile_url}</span>
                  </p>
                </div>
                <p className="text-center text-sm text-slate-400">Make sure your profile is <strong className="text-white">public</strong> before continuing.</p>
                <button
                  onClick={() => { setErrorMsg(''); setStep('phone') }}
                  className="w-full py-3.5 rounded-xl font-bold text-white bg-gradient-brand hover:opacity-90 transition-all"
                >
                  Continue Anyway
                </button>
                <button
                  onClick={() => { setStep('username'); setProfile(null); setErrorMsg('') }}
                  className="w-full py-2 text-sm text-slate-500 hover:text-slate-300 transition-colors"
                >
                  ← Try a different username
                </button>
              </>
            )}
          </div>
        )}

        {/* ── Step: Phone ──────────────────────────────────────────────── */}
        {step === 'phone' && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-semibold text-slate-200 mb-2">M-Pesa Phone Number</label>
              <div className="relative">
                <Phone size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                <input
                  type="tel"
                  placeholder="0712 345 678"
                  value={phone}
                  onChange={e => setPhone(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && !loading && handlePhoneSubmit()}
                  className="w-full bg-surface-2 border border-white/8 rounded-xl pl-9 pr-4 py-3 text-white placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors text-sm"
                  autoFocus
                />
              </div>
              <p className="text-xs text-slate-500 mt-1.5">You'll receive an M-Pesa STK push on this number</p>
            </div>

            {errorMsg && <p className="text-red-400 text-sm">{errorMsg}</p>}

            <button
              onClick={handlePhoneSubmit}
              disabled={loading}
              className="w-full py-3.5 rounded-xl font-bold text-white bg-gradient-brand hover:opacity-90 disabled:opacity-50 transition-all flex items-center justify-center gap-2"
            >
              {loading
                ? <><Loader2 size={18} className="animate-spin" /> Sending M-Pesa request…</>
                : <>Pay KES {pkg.price_kes.toLocaleString()} with M-Pesa</>}
            </button>
            <button
              onClick={() => { setStep('verify'); setErrorMsg('') }}
              className="w-full py-2 text-sm text-slate-500 hover:text-slate-300 transition-colors"
            >
              ← Back
            </button>
          </div>
        )}

        {/* ── Step: Waiting ─────────────────────────────────────────────── */}
        {step === 'waiting' && (
          <div className="text-center py-4 space-y-5">
            <div className="w-16 h-16 mx-auto rounded-full bg-violet-500/10 border border-violet-500/30 flex items-center justify-center">
              <Loader2 size={28} className="text-violet-400 animate-spin" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-white mb-2">Check Your Phone</h3>
              <p className="text-slate-400 text-sm leading-relaxed">
                An M-Pesa prompt has been sent to <strong className="text-white">{phone}</strong>.<br />
                Enter your PIN to complete the payment.
              </p>
            </div>
            <div className="glass rounded-xl p-4 text-sm text-slate-400">
              Order #{orderId} · Waiting for confirmation…
            </div>
            <p className="text-xs text-slate-600">This page updates automatically. Do not close it.</p>
          </div>
        )}

        {/* ── Step: Success ─────────────────────────────────────────────── */}
        {step === 'success' && (
          <div className="text-center py-4 space-y-5">
            <div className="w-16 h-16 mx-auto rounded-full bg-green-500/10 border border-green-500/30 flex items-center justify-center">
              <CheckCircle size={28} className="text-green-400" />
            </div>
            <div>
              <h3 className="text-xl font-bold text-white mb-2">Payment Confirmed! 🎉</h3>
              <p className="text-slate-400 text-sm leading-relaxed">
                Your order is processing. Followers will start arriving within minutes.
              </p>
            </div>
            <button onClick={onClose} className="w-full py-3 text-sm text-slate-400 hover:text-white transition-colors">
              Close
            </button>
          </div>
        )}

        {/* ── Step: Error ───────────────────────────────────────────────── */}
        {step === 'error' && (
          <div className="text-center py-4 space-y-5">
            <div className="w-16 h-16 mx-auto rounded-full bg-red-500/10 border border-red-500/30 flex items-center justify-center">
              <AlertCircle size={28} className="text-red-400" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-white mb-2">Something Went Wrong</h3>
              <p className="text-slate-400 text-sm">{errorMsg}</p>
            </div>
            <button
              onClick={() => { setStep('phone'); setErrorMsg('') }}
              className="w-full py-3 rounded-xl font-semibold text-white bg-gradient-brand hover:opacity-90 transition-opacity"
            >
              Try Again
            </button>
            <a href="https://t.me/workratew" target="_blank" rel="noopener noreferrer"
              className="block w-full py-2 text-sm text-slate-500 hover:text-violet-300 transition-colors">
              Contact Support →
            </a>
          </div>
        )}
      </div>
    </div>
  )
}

function buildProfileURL(platform: string, username: string): string {
  const clean = username.replace(/^@/, '')
  switch (platform) {
    case 'tiktok':    return `https://www.tiktok.com/@${clean}`
    case 'instagram': return `https://www.instagram.com/${clean}/`
    case 'youtube':   return `https://www.youtube.com/@${clean}`
    default:          return `https://www.tiktok.com/@${clean}`
  }
}
