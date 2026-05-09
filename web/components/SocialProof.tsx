'use client'

import { SiTiktok, SiInstagram, SiYoutube } from 'react-icons/si'

const proofs = [
  { platform: 'tiktok',    pkg: 'Viral Starter',    time: '2 mins ago',   flag: '🇰🇪' },
  { platform: 'instagram', pkg: 'Quick-Start',       time: '7 mins ago',   flag: '🇺🇬' },
  { platform: 'tiktok',    pkg: 'TikTok Starter',    time: '12 mins ago',  flag: '🇰🇪' },
  { platform: 'youtube',   pkg: 'Kickstart',         time: '18 mins ago',  flag: '🇹🇿' },
  { platform: 'instagram', pkg: 'Celebrity Pack',    time: '24 mins ago',  flag: '🇰🇪' },
  { platform: 'tiktok',    pkg: 'Viral Creator',     time: '31 mins ago',  flag: '🇺🇬' },
  { platform: 'instagram', pkg: 'Business Boost',    time: '45 mins ago',  flag: '🇰🇪' },
  { platform: 'youtube',   pkg: 'Kickstart',         time: '1 hr ago',     flag: '🇹🇿' },
]

const platformIcon = {
  tiktok:    { Icon: SiTiktok,    bg: 'bg-black',                                                  border: 'border-white/10',  iconColor: 'text-white',  label: 'TikTok'    },
  instagram: { Icon: SiInstagram, bg: 'bg-gradient-to-br from-[#833ab4] via-[#fd1d1d] to-[#fcb045]', border: 'border-transparent', iconColor: 'text-white', label: 'Instagram' },
  youtube:   { Icon: SiYoutube,   bg: 'bg-[#FF0000]',                                              border: 'border-transparent', iconColor: 'text-white',  label: 'YouTube'   },
}

export default function SocialProof() {
  return (
    <section className="py-20 px-4 overflow-hidden">
      <div className="max-w-6xl mx-auto">
        <div className="text-center mb-12">
          <h2 className="text-3xl sm:text-4xl font-black mb-4">
            Live <span className="gradient-text">Activity Feed</span>
          </h2>
          <p className="text-slate-400">Real orders delivered across East Africa</p>
        </div>

        <div className="relative overflow-hidden">
          {/* Fade edges */}
          <div className="pointer-events-none absolute left-0 top-0 bottom-0 w-20 bg-gradient-to-r from-dark to-transparent z-10" />
          <div className="pointer-events-none absolute right-0 top-0 bottom-0 w-20 bg-gradient-to-l from-dark to-transparent z-10" />

          <div className="flex gap-4" style={{ animation: 'scrollTicker 28s linear infinite' }}>
            {[...proofs, ...proofs].map((p, i) => {
              const cfg = platformIcon[p.platform as keyof typeof platformIcon]
              const { Icon } = cfg
              return (
                <div
                  key={i}
                  className="flex-shrink-0 glass rounded-xl px-5 py-3.5 flex items-center gap-3.5 min-w-[240px]"
                >
                  {/* Platform icon */}
                  <div className={`w-10 h-10 rounded-xl ${cfg.bg} border ${cfg.border} flex items-center justify-center flex-shrink-0 shadow-sm`}>
                    <Icon size={16} className={cfg.iconColor} />
                  </div>
                  {/* Check badge */}
                  <div className="w-5 h-5 rounded-full bg-green-500/15 border border-green-500/30 flex items-center justify-center flex-shrink-0">
                    <svg width="8" height="8" viewBox="0 0 8 8" fill="none">
                      <path d="M1 4l2 2 4-4" stroke="#4ade80" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                  </div>
                  <div className="min-w-0">
                    <p className="text-sm font-semibold text-white truncate">{cfg.label} · {p.pkg}</p>
                    <p className="text-xs text-slate-500">{p.flag} Delivered {p.time}</p>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      </div>

      <style jsx>{`
        @keyframes scrollTicker {
          0%   { transform: translateX(0); }
          100% { transform: translateX(-50%); }
        }
      `}</style>
    </section>
  )
}
