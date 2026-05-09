'use client'

import { useEffect, useRef, useState } from 'react'
import { ArrowRight, Shield, RefreshCw } from 'lucide-react'
import { SiTiktok, SiInstagram, SiYoutube, SiTelegram } from 'react-icons/si'

// ── Typewriter ───────────────────────────────────────────────────────────────

const PLATFORMS = ['TikTok', 'Instagram', 'YouTube']

function Typewriter() {
  const [idx, setIdx] = useState(0)
  const [displayed, setDisplayed] = useState('')
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    const target = PLATFORMS[idx]
    const speed = deleting ? 70 : 110

    if (!deleting && displayed === target) {
      const t = setTimeout(() => setDeleting(true), 1800)
      return () => clearTimeout(t)
    }
    if (deleting && displayed === '') {
      setDeleting(false)
      setIdx(i => (i + 1) % PLATFORMS.length)
      return
    }
    const t = setTimeout(() => {
      setDisplayed(prev =>
        deleting ? prev.slice(0, -1) : target.slice(0, prev.length + 1)
      )
    }, speed)
    return () => clearTimeout(t)
  }, [displayed, deleting, idx])

  return (
    <span className="gradient-text">
      {displayed}
      <span className="inline-block w-[3px] h-[0.85em] bg-violet-400 ml-1 align-middle animate-[blink_1s_step-end_infinite]" />
    </span>
  )
}

// ── Scroll-triggered counter ─────────────────────────────────────────────────

function Counter({ target, suffix = '', trigger }: { target: number; suffix?: string; trigger: boolean }) {
  const [count, setCount] = useState(0)

  useEffect(() => {
    if (!trigger) return
    const duration = 2000
    const steps = 60
    const increment = target / steps
    let current = 0
    const timer = setInterval(() => {
      current += increment
      if (current >= target) {
        setCount(target)
        clearInterval(timer)
      } else {
        setCount(Math.floor(current))
      }
    }, duration / steps)
    return () => clearInterval(timer)
  }, [trigger, target])

  return <span>{count.toLocaleString()}{suffix}</span>
}

// ── M-Pesa badge ─────────────────────────────────────────────────────────────

function MPesaBadge() {
  return (
    <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#00A651]/10 border border-[#00A651]/25 text-[#00A651] text-xs font-black tracking-wide align-middle">
      <span className="w-1.5 h-1.5 rounded-full bg-[#00A651] animate-pulse" />
      M-PESA
    </span>
  )
}

// ── Platform pills ────────────────────────────────────────────────────────────

const platforms = [
  { Icon: SiTiktok,    label: 'TikTok',    bg: 'bg-black',                                                               border: 'border-white/10' },
  { Icon: SiInstagram, label: 'Instagram', bg: 'bg-gradient-to-br from-[#833ab4] via-[#fd1d1d] to-[#fcb045]',           border: 'border-transparent' },
  { Icon: SiYoutube,   label: 'YouTube',   bg: 'bg-[#FF0000]',                                                           border: 'border-transparent' },
]

// ── Hero ─────────────────────────────────────────────────────────────────────

export default function Hero() {
  const statsRef = useRef<HTMLDivElement>(null)
  const [statsVisible, setStatsVisible] = useState(false)

  useEffect(() => {
    const el = statsRef.current
    if (!el) return
    const observer = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) { setStatsVisible(true); observer.unobserve(el) } },
      { threshold: 0.3 }
    )
    observer.observe(el)
    return () => observer.disconnect()
  }, [])

  return (
    <section className="relative min-h-screen flex items-center justify-center px-4 overflow-hidden pt-16">
      {/* Animated mesh background */}
      <div className="absolute inset-0 bg-gradient-hero pointer-events-none" />
      <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] rounded-full pointer-events-none"
        style={{ background: 'radial-gradient(circle, rgba(124,58,237,0.12) 0%, transparent 65%)', animation: 'meshMove1 18s ease-in-out infinite' }} />
      <div className="absolute top-3/4 left-1/4 w-[500px] h-[500px] rounded-full pointer-events-none"
        style={{ background: 'radial-gradient(circle, rgba(6,182,212,0.07) 0%, transparent 65%)', animation: 'meshMove2 22s ease-in-out infinite' }} />
      <div className="absolute top-1/2 right-1/4 w-[400px] h-[400px] rounded-full pointer-events-none"
        style={{ background: 'radial-gradient(circle, rgba(167,139,250,0.06) 0%, transparent 65%)', animation: 'meshMove3 26s ease-in-out infinite' }} />

      {/* Floating orbs */}
      <div className="absolute top-32 right-16 w-3 h-3 rounded-full bg-violet-500 animate-float opacity-50" />
      <div className="absolute top-52 right-36 w-2 h-2 rounded-full bg-cyan-400 animate-float opacity-35" style={{ animationDelay: '1s' }} />
      <div className="absolute bottom-36 left-16 w-4 h-4 rounded-full bg-violet-400 animate-float opacity-25" style={{ animationDelay: '2s' }} />
      <div className="absolute bottom-52 left-36 w-2 h-2 rounded-full bg-cyan-500 animate-float opacity-45" style={{ animationDelay: '0.5s' }} />

      <div className="relative z-10 max-w-4xl mx-auto text-center">
        {/* Live badge */}
        <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-violet-500/8 border border-violet-500/20 text-violet-300 text-sm font-medium mb-8 animate-fade-up">
          <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
          Live orders being delivered right now
        </div>

        {/* Platform pills */}
        <div className="flex items-center justify-center gap-3 mb-8 animate-fade-up" style={{ animationDelay: '0.1s' }}>
          {platforms.map(({ Icon, label, bg, border }) => (
            <div key={label} className={`flex items-center gap-2 px-4 py-2 rounded-full ${bg} border ${border} text-white text-sm font-semibold shadow-lg`}>
              <Icon size={15} className="text-white" />
              {label}
            </div>
          ))}
        </div>

        {/* Heading with typewriter */}
        <h1 className="text-5xl sm:text-6xl md:text-7xl font-black leading-[1.05] tracking-tight mb-6 animate-fade-up" style={{ animationDelay: '0.2s' }}>
          Grow Your <Typewriter />
          <br />
          <span className="text-white">in Minutes</span>
        </h1>

        <p className="text-slate-400 text-lg sm:text-xl max-w-2xl mx-auto mb-10 leading-relaxed animate-fade-up" style={{ animationDelay: '0.3s' }}>
          Real followers, likes &amp; views. Pay with <MPesaBadge />. Delivery starts instantly.
        </p>

        {/* CTA */}
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-16 animate-fade-up" style={{ animationDelay: '0.4s' }}>
          <a
            href="#packages"
            className="flex items-center gap-2 px-8 py-4 rounded-xl font-bold text-white bg-gradient-brand glow-violet hover:opacity-90 hover:scale-105 transition-all text-lg shadow-xl shadow-violet-900/30"
          >
            View Packages
            <ArrowRight size={20} />
          </a>
          <a
            href="https://t.me/pompomputrin888pom_bot"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 px-8 py-4 rounded-xl font-semibold text-slate-300 border border-white/10 hover:border-slate-500 hover:text-white hover:scale-105 transition-all text-lg"
          >
            <SiTelegram size={18} className="text-[#2AABEE]" />
            Telegram Bot
          </a>
        </div>

        {/* Stats — count up only when scrolled into view */}
        <div ref={statsRef} className="grid grid-cols-3 gap-4 max-w-lg mx-auto animate-fade-up" style={{ animationDelay: '0.5s' }}>
          {[
            { label: 'Orders Delivered', value: 12000, suffix: '+' },
            { label: 'Delivery Rate',    value: 99,    suffix: '%' },
            { label: 'Happy Clients',    value: 3400,  suffix: '+' },
          ].map((stat) => (
            <div key={stat.label} className="glass rounded-xl p-4 text-center hover:border-violet-500/30 transition-colors">
              <div className="text-2xl font-black gradient-text mb-1">
                <Counter target={stat.value} suffix={stat.suffix} trigger={statsVisible} />
              </div>
              <div className="text-xs text-slate-500">{stat.label}</div>
            </div>
          ))}
        </div>

        {/* Trust badges */}
        <div className="flex flex-wrap items-center justify-center gap-5 mt-10 text-slate-500 text-sm animate-fade-up" style={{ animationDelay: '0.6s' }}>
          <div className="flex items-center gap-2">
            <Shield size={14} className="text-green-500" />
            Secure M-Pesa
          </div>
          <span className="text-slate-700">·</span>
          <div className="flex items-center gap-2">
            <RefreshCw size={14} className="text-violet-400" />
            30-day refill
          </div>
          <span className="text-slate-700">·</span>
          <div className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
            24/7 support
          </div>
        </div>
      </div>
    </section>
  )
}
