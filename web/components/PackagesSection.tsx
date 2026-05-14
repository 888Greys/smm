'use client'

import { useEffect, useRef, useState } from 'react'
import { Package } from '@/lib/types'
import { getPackages } from '@/lib/api'
import OrderModal from './OrderModal'
import { SiTiktok, SiInstagram, SiYoutube } from 'react-icons/si'
import { RefreshCw, Star, Zap } from 'lucide-react'

const platformConfig = {
  tiktok: {
    Icon: SiTiktok,
    bg: 'bg-black',
    border: 'border-[#69C9D0]/30',
    iconColor: 'text-white',
    label: 'TikTok',
    filterActive: 'bg-black text-white border-[#69C9D0]/50',
    filterDot: 'bg-[#69C9D0]',
  },
  instagram: {
    Icon: SiInstagram,
    bg: 'bg-gradient-to-br from-[#833ab4] via-[#fd1d1d] to-[#fcb045]',
    border: 'border-[#fd1d1d]/20',
    iconColor: 'text-white',
    label: 'Instagram',
    filterActive: 'bg-gradient-to-r from-[#833ab4] to-[#fd1d1d] text-white border-transparent',
    filterDot: 'bg-[#fd1d1d]',
  },
  youtube: {
    Icon: SiYoutube,
    bg: 'bg-[#FF0000]',
    border: 'border-[#FF0000]/20',
    iconColor: 'text-white',
    label: 'YouTube',
    filterActive: 'bg-[#FF0000] text-white border-transparent',
    filterDot: 'bg-[#FF0000]',
  },
}

const tier = (priceKES: number) => {
  if (priceKES <= 600)  return { label: 'Starter', color: 'text-slate-400 border-slate-700/50 bg-slate-900/50' }
  if (priceKES <= 1000) return { label: 'Growth',  color: 'text-violet-300 border-violet-700/50 bg-violet-950/50' }
  return                        { label: 'Power',   color: 'text-cyan-300 border-cyan-700/50 bg-cyan-950/50' }
}

function PackageCard({ pkg, onOrder }: { pkg: Package; onOrder: (p: Package) => void }) {
  const cfg = platformConfig[pkg.platform as keyof typeof platformConfig] || platformConfig.tiktok
  const { Icon } = cfg
  const t = tier(pkg.price_kes)
  const cardRef = useRef<HTMLDivElement>(null)

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const el = cardRef.current
    if (!el) return
    const rect = el.getBoundingClientRect()
    const x = (e.clientX - rect.left) / rect.width - 0.5
    const y = (e.clientY - rect.top) / rect.height - 0.5
    el.style.transform = `perspective(900px) rotateY(${x * 12}deg) rotateX(${-y * 12}deg) translateZ(6px)`
  }

  const handleMouseLeave = () => {
    if (cardRef.current) cardRef.current.style.transform = ''
  }

  return (
    <div
      ref={cardRef}
      onMouseMove={handleMouseMove}
      onMouseLeave={handleMouseLeave}
      className="relative glass rounded-2xl p-6 flex flex-col gap-4 hover:border-violet-500/40 transition-[border-color,box-shadow] duration-300 group hover:shadow-2xl hover:shadow-violet-900/20"
      style={{ transformStyle: 'preserve-3d', transition: 'transform 0.15s ease, border-color 0.3s, box-shadow 0.3s' }}
    >
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className={`w-11 h-11 rounded-xl ${cfg.bg} flex items-center justify-center shadow-lg flex-shrink-0`}>
          <Icon size={20} className={cfg.iconColor} />
        </div>
        <span className={`text-xs font-semibold px-2.5 py-1 rounded-full border ${t.color}`}>
          {t.label}
        </span>
      </div>

      {/* Name + description */}
      <div>
        <h3 className="font-bold text-white text-lg mb-1.5 group-hover:text-violet-200 transition-colors leading-snug">
          {pkg.name}
        </h3>
        <p className="text-slate-400 text-sm leading-relaxed">{pkg.description}</p>
      </div>

      {/* Feature chips */}
      <div className="flex flex-wrap gap-2">
        <span className="flex items-center gap-1.5 text-xs text-slate-400 bg-slate-800/60 rounded-full px-3 py-1 border border-white/5">
          <Zap size={10} className="text-yellow-400" /> Instant start
        </span>
        {pkg.price_kes >= 500 && (
          <span className="flex items-center gap-1.5 text-xs text-slate-400 bg-slate-800/60 rounded-full px-3 py-1 border border-white/5">
            <RefreshCw size={10} className="text-violet-400" /> 30-day refill
          </span>
        )}
        {pkg.price_kes >= 1500 && (
          <span className="flex items-center gap-1.5 text-xs text-slate-400 bg-slate-800/60 rounded-full px-3 py-1 border border-white/5">
            <Star size={10} className="text-cyan-400" /> Premium quality
          </span>
        )}
      </div>

      {/* Price + CTA */}
      <div className="mt-auto pt-4 flex items-center justify-between border-t border-white/5">
        <div>
          <p className="text-xs text-slate-500 mb-0.5">From</p>
          <span className="text-2xl font-black text-white tracking-tight">KES {pkg.price_kes.toLocaleString()}</span>
        </div>
        <button
          onClick={() => onOrder(pkg)}
          className="px-5 py-2.5 rounded-xl font-bold text-sm text-white bg-gradient-brand hover:opacity-90 active:scale-95 transition-all shadow-lg shadow-violet-900/30"
        >
          Buy Now
        </button>
      </div>
    </div>
  )
}

const filterLabels: Record<string, { icon: React.ComponentType<{ size: number; className?: string }> | null; label: string }> = {
  all:       { icon: null, label: 'All Platforms' },
  tiktok:    { icon: SiTiktok,    label: 'TikTok' },
  instagram: { icon: SiInstagram, label: 'Instagram' },
  youtube:   { icon: SiYoutube,   label: 'YouTube' },
}

export default function PackagesSection({ initialPackages }: { initialPackages: Package[] }) {
  const [packages, setPackages] = useState<Package[]>(initialPackages)
  const [selected, setSelected] = useState<Package | null>(null)
  const [filter, setFilter] = useState<string>('all')

  useEffect(() => {
    getPackages().then(setPackages).catch(() => {})
  }, [])

  const platforms = ['all', 'tiktok', 'instagram', 'youtube']
  const filtered = filter === 'all' ? packages : packages.filter(p => p.platform === filter)

  return (
    <section id="packages" className="py-24 px-4">
      <div className="max-w-6xl mx-auto">
        {/* Heading */}
        <div className="text-center mb-12">
          <h2 className="text-3xl sm:text-4xl font-black mb-4">
            Choose Your <span className="gradient-text">Growth Package</span>
          </h2>
          <p className="text-slate-400 max-w-xl mx-auto">
            Real accounts, fast delivery, backed by a 30-day refill guarantee. Pay securely with M-Pesa.
          </p>
        </div>

        {/* Platform filter */}
        <div className="flex flex-wrap gap-3 justify-center mb-10">
          {platforms.map(p => {
            const fl = filterLabels[p]
            const IconComp = fl.icon
            const isActive = filter === p
            const cfg = platformConfig[p as keyof typeof platformConfig]
            return (
              <button
                key={p}
                onClick={() => setFilter(p)}
                className={`flex items-center gap-2 px-5 py-2.5 rounded-full text-sm font-semibold transition-all border ${
                  isActive
                    ? (cfg ? cfg.filterActive : 'bg-violet-600 text-white border-violet-500')
                    : 'bg-surface text-slate-400 border-border-dim hover:border-slate-500 hover:text-white'
                }`}
              >
                {IconComp && <IconComp size={14} className={isActive ? 'opacity-100' : 'opacity-60'} />}
                {fl.label}
              </button>
            )
          })}
        </div>

        {/* Package grid */}
        {packages.length === 0 ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="glass rounded-2xl p-6 h-64 shimmer" />
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {filtered.map(pkg => (
              <PackageCard key={pkg.id} pkg={pkg} onOrder={setSelected} />
            ))}
          </div>
        )}

        {/* Telegram CTA */}
        <div className="mt-14 text-center">
          <p className="text-slate-500 text-sm">
            Prefer ordering via Telegram?{' '}
            <a
              href="https://t.me/pompomputrin888pom_bot"
              target="_blank"
              rel="noopener noreferrer"
              className="text-violet-400 hover:text-violet-300 font-medium transition-colors"
            >
              Open our bot →
            </a>
          </p>
        </div>
      </div>

      {selected && (
        <OrderModal pkg={selected} onClose={() => setSelected(null)} />
      )}
    </section>
  )
}
