'use client'

import { useEffect, useState } from 'react'
import { Package } from '@/lib/types'
import { getPackages } from '@/lib/api'
import OrderModal from './OrderModal'
import { Music2, Camera, Youtube, Zap, RefreshCw, Star } from 'lucide-react'

const platformConfig = {
  tiktok:    { icon: Music2,  color: 'from-pink-500 to-violet-600',  label: 'TikTok'    },
  instagram: { icon: Camera,  color: 'from-orange-500 to-pink-600',  label: 'Instagram' },
  youtube:   { icon: Youtube, color: 'from-red-500 to-red-700',      label: 'YouTube'   },
}

const tier = (priceKES: number) => {
  if (priceKES <= 600)  return { label: 'Entry',  color: 'text-slate-400 border-slate-700' }
  if (priceKES <= 1000) return { label: 'Growth', color: 'text-violet-400 border-violet-700' }
  return                        { label: 'Power',  color: 'text-cyan-400 border-cyan-700' }
}

function PackageCard({ pkg, onOrder }: { pkg: Package; onOrder: (p: Package) => void }) {
  const cfg = platformConfig[pkg.platform] || platformConfig.tiktok
  const Icon = cfg.icon
  const t = tier(pkg.price_kes)

  return (
    <div className="glass rounded-2xl p-6 flex flex-col gap-4 hover:border-violet-500/30 transition-all duration-300 group hover:glow-violet-sm">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className={`w-10 h-10 rounded-xl bg-gradient-to-br ${cfg.color} flex items-center justify-center shadow-lg`}>
          <Icon size={18} className="text-white" />
        </div>
        <span className={`text-xs font-semibold px-2.5 py-1 rounded-full border ${t.color}`}>
          {t.label}
        </span>
      </div>

      {/* Name + description */}
      <div>
        <h3 className="font-bold text-white text-lg mb-1 group-hover:text-violet-200 transition-colors">
          {pkg.name}
        </h3>
        <p className="text-slate-400 text-sm leading-relaxed">{pkg.description}</p>
      </div>

      {/* Features */}
      <div className="flex flex-wrap gap-2">
        <span className="flex items-center gap-1 text-xs text-slate-400 bg-surface-2 rounded-full px-2.5 py-1">
          <Zap size={10} className="text-yellow-400" /> Instant start
        </span>
        {pkg.price_kes >= 500 && (
          <span className="flex items-center gap-1 text-xs text-slate-400 bg-surface-2 rounded-full px-2.5 py-1">
            <RefreshCw size={10} className="text-violet-400" /> 30-day refill
          </span>
        )}
        {pkg.price_kes >= 1500 && (
          <span className="flex items-center gap-1 text-xs text-slate-400 bg-surface-2 rounded-full px-2.5 py-1">
            <Star size={10} className="text-cyan-400" /> Premium quality
          </span>
        )}
      </div>

      {/* Price + CTA */}
      <div className="mt-auto pt-2 flex items-center justify-between border-t border-border-dim">
        <div>
          <span className="text-2xl font-black text-white">KES {pkg.price_kes.toLocaleString()}</span>
        </div>
        <button
          onClick={() => onOrder(pkg)}
          className="px-5 py-2.5 rounded-xl font-semibold text-sm text-white bg-gradient-brand hover:opacity-90 transition-opacity"
        >
          Buy Now
        </button>
      </div>
    </div>
  )
}

export default function PackagesSection({ initialPackages }: { initialPackages: Package[] }) {
  const [packages, setPackages] = useState<Package[]>(initialPackages)
  const [selected, setSelected] = useState<Package | null>(null)
  const [filter, setFilter] = useState<string>('all')

  // Hydrate on client if server-side fetch failed
  useEffect(() => {
    if (packages.length === 0) {
      getPackages().then(setPackages).catch(() => {})
    }
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
            All packages use real accounts. Fast delivery guaranteed. Pay securely with M-Pesa STK push.
          </p>
        </div>

        {/* Platform filter */}
        <div className="flex flex-wrap gap-2 justify-center mb-10">
          {platforms.map(p => (
            <button
              key={p}
              onClick={() => setFilter(p)}
              className={`px-5 py-2 rounded-full text-sm font-semibold capitalize transition-all ${
                filter === p
                  ? 'bg-violet-600 text-white'
                  : 'bg-surface text-slate-400 border border-border-dim hover:border-slate-500'
              }`}
            >
              {p === 'all' ? 'All Platforms' : p === 'tiktok' ? 'TikTok' : p === 'instagram' ? 'Instagram' : 'YouTube'}
            </button>
          ))}
        </div>

        {/* Grid */}
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
        <div className="mt-12 text-center">
          <p className="text-slate-500 text-sm">
            Prefer ordering via Telegram?{' '}
            <a
              href="https://t.me/AaPomSMM"
              target="_blank"
              rel="noopener noreferrer"
              className="text-violet-400 hover:text-violet-300 font-medium"
            >
              Open our bot →
            </a>
          </p>
        </div>
      </div>

      {/* Order modal */}
      {selected && (
        <OrderModal pkg={selected} onClose={() => setSelected(null)} />
      )}
    </section>
  )
}
