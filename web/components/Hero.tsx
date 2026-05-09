'use client'

import { useEffect, useState } from 'react'
import { ArrowRight, Shield, Zap, RefreshCw } from 'lucide-react'

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

export default function Hero() {
  return (
    <section className="relative min-h-screen flex items-center justify-center px-4 overflow-hidden pt-16">
      {/* Background glow */}
      <div className="absolute inset-0 bg-gradient-hero pointer-events-none" />
      <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] rounded-full bg-violet-600/10 blur-[120px] pointer-events-none" />

      {/* Floating orbs */}
      <div className="absolute top-32 right-16 w-3 h-3 rounded-full bg-violet-500 animate-float opacity-60" />
      <div className="absolute top-48 right-32 w-2 h-2 rounded-full bg-cyan-400 animate-float opacity-40" style={{ animationDelay: '1s' }} />
      <div className="absolute bottom-32 left-16 w-4 h-4 rounded-full bg-violet-400 animate-float opacity-30" style={{ animationDelay: '2s' }} />
      <div className="absolute bottom-48 left-32 w-2 h-2 rounded-full bg-cyan-500 animate-float opacity-50" style={{ animationDelay: '0.5s' }} />

      <div className="relative z-10 max-w-4xl mx-auto text-center">
        {/* Badge */}
        <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-violet-500/10 border border-violet-500/20 text-violet-300 text-sm font-medium mb-8">
          <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse-slow" />
          Live orders being delivered right now
        </div>

        {/* Heading */}
        <h1 className="text-5xl sm:text-6xl md:text-7xl font-black leading-[1.05] tracking-tight mb-6">
          Grow Your{' '}
          <span className="gradient-text">Social Media</span>
          <br />
          in Minutes
        </h1>

        <p className="text-slate-400 text-lg sm:text-xl max-w-2xl mx-auto mb-10 leading-relaxed">
          Real followers, likes &amp; views for TikTok, Instagram &amp; YouTube.
          Pay with <strong className="text-white">M-Pesa</strong>. Delivery starts instantly.
        </p>

        {/* CTA */}
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-16">
          <a
            href="#packages"
            className="flex items-center gap-2 px-8 py-4 rounded-xl font-bold text-white bg-gradient-brand glow-violet hover:opacity-90 transition-opacity text-lg"
          >
            View Packages
            <ArrowRight size={20} />
          </a>
          <a
            href="https://t.me/AaPomSMM"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 px-8 py-4 rounded-xl font-semibold text-slate-300 border border-border-dim hover:border-slate-500 hover:text-white transition-all text-lg"
          >
            Open Telegram Bot
          </a>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-3 gap-4 max-w-lg mx-auto">
          {[
            { label: 'Orders Delivered', value: 12000, suffix: '+' },
            { label: 'Delivery Rate', value: 99, suffix: '%' },
            { label: 'Happy Clients', value: 3400, suffix: '+' },
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
        <div className="flex items-center justify-center gap-6 mt-10 text-slate-500 text-sm">
          <div className="flex items-center gap-1.5">
            <Shield size={14} className="text-green-500" />
            Secure M-Pesa
          </div>
          <div className="flex items-center gap-1.5">
            <Zap size={14} className="text-yellow-500" />
            Instant start
          </div>
          <div className="flex items-center gap-1.5">
            <RefreshCw size={14} className="text-violet-400" />
            30-day refill
          </div>
        </div>
      </div>
    </section>
  )
}
