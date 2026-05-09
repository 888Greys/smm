'use client'

import { useEffect, useState } from 'react'
import { ArrowRight, Shield, RefreshCw } from 'lucide-react'
import { SiTiktok, SiInstagram, SiYoutube, SiTelegram } from 'react-icons/si'

function Counter({ target, suffix = '' }: { target: number; suffix?: string }) {
  const [count, setCount] = useState(0)

  useEffect(() => {
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
  }, [target])

  return <span>{count.toLocaleString()}{suffix}</span>
}

function MPesaBadge() {
  return (
    <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#00A651]/10 border border-[#00A651]/25 text-[#00A651] text-xs font-black tracking-wide">
      <span className="w-1.5 h-1.5 rounded-full bg-[#00A651] animate-pulse" />
      M-PESA
    </span>
  )
}

const platforms = [
  { Icon: SiTiktok,    label: 'TikTok',    bg: 'bg-black',                                                               border: 'border-white/10' },
  { Icon: SiInstagram, label: 'Instagram', bg: 'bg-gradient-to-br from-[#833ab4] via-[#fd1d1d] to-[#fcb045]',           border: 'border-transparent' },
  { Icon: SiYoutube,   label: 'YouTube',   bg: 'bg-[#FF0000]',                                                           border: 'border-transparent' },
]

export default function Hero() {
  return (
    <section className="relative min-h-screen flex items-center justify-center px-4 overflow-hidden pt-16">
      {/* Background glows */}
      <div className="absolute inset-0 bg-gradient-hero pointer-events-none" />
      <div className="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] h-[700px] rounded-full bg-violet-600/8 blur-[140px] pointer-events-none" />
      <div className="absolute top-1/4 right-0 w-[400px] h-[400px] rounded-full bg-cyan-500/5 blur-[120px] pointer-events-none" />

      {/* Floating orbs */}
      <div className="absolute top-32 right-16 w-3 h-3 rounded-full bg-violet-500 animate-float opacity-50" />
      <div className="absolute top-52 right-36 w-2 h-2 rounded-full bg-cyan-400 animate-float opacity-35" style={{ animationDelay: '1s' }} />
      <div className="absolute bottom-36 left-16 w-4 h-4 rounded-full bg-violet-400 animate-float opacity-25" style={{ animationDelay: '2s' }} />
      <div className="absolute bottom-52 left-36 w-2 h-2 rounded-full bg-cyan-500 animate-float opacity-45" style={{ animationDelay: '0.5s' }} />

      <div className="relative z-10 max-w-4xl mx-auto text-center">
        {/* Live badge */}
        <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-violet-500/8 border border-violet-500/20 text-violet-300 text-sm font-medium mb-8">
          <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
          Live orders being delivered right now
        </div>

        {/* Platform icons row */}
        <div className="flex items-center justify-center gap-3 mb-8">
          {platforms.map(({ Icon, label, bg, border }) => (
            <div
              key={label}
              className={`flex items-center gap-2 px-4 py-2 rounded-full ${bg} border ${border} text-white text-sm font-semibold shadow-lg`}
            >
              <Icon size={15} className="text-white" />
              {label}
            </div>
          ))}
        </div>

        {/* Heading */}
        <h1 className="text-5xl sm:text-6xl md:text-7xl font-black leading-[1.05] tracking-tight mb-6">
          Grow Your{' '}
          <span className="gradient-text">Social Media</span>
          <br />
          in Minutes
        </h1>

        <p className="text-slate-400 text-lg sm:text-xl max-w-2xl mx-auto mb-10 leading-relaxed">
          Real followers, likes &amp; views. Pay with <MPesaBadge />. Delivery starts instantly.
        </p>

        {/* CTA */}
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-16">
          <a
            href="#packages"
            className="flex items-center gap-2 px-8 py-4 rounded-xl font-bold text-white bg-gradient-brand glow-violet hover:opacity-90 transition-opacity text-lg shadow-xl shadow-violet-900/30"
          >
            View Packages
            <ArrowRight size={20} />
          </a>
          <a
            href="https://t.me/pompomputrin888pom_bot"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 px-8 py-4 rounded-xl font-semibold text-slate-300 border border-white/10 hover:border-slate-500 hover:text-white transition-all text-lg"
          >
            <SiTelegram size={18} className="text-[#2AABEE]" />
            Telegram Bot
          </a>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-3 gap-4 max-w-lg mx-auto">
          {[
            { label: 'Orders Delivered', value: 12000, suffix: '+' },
            { label: 'Delivery Rate',    value: 99,    suffix: '%' },
            { label: 'Happy Clients',    value: 3400,  suffix: '+' },
          ].map((stat) => (
            <div key={stat.label} className="glass rounded-xl p-4 text-center">
              <div className="text-2xl font-black gradient-text mb-1">
                <Counter target={stat.value} suffix={stat.suffix} />
              </div>
              <div className="text-xs text-slate-500">{stat.label}</div>
            </div>
          ))}
        </div>

        {/* Trust badges */}
        <div className="flex flex-wrap items-center justify-center gap-5 mt-10 text-slate-500 text-sm">
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
