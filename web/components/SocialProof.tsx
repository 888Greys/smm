const proofs = [
  { platform: '🎵 TikTok',    pkg: 'Viral Starter',    time: '2 mins ago',   flag: '🇰🇪' },
  { platform: '📸 Instagram', pkg: 'Quick-Start',       time: '7 mins ago',   flag: '🇺🇬' },
  { platform: '🎵 TikTok',    pkg: 'TikTok Flex',       time: '12 mins ago',  flag: '🇰🇪' },
  { platform: '▶️ YouTube',   pkg: 'Kickstart',         time: '18 mins ago',  flag: '🇹🇿' },
  { platform: '📸 Instagram', pkg: 'Celebrity Pack',    time: '24 mins ago',  flag: '🇰🇪' },
  { platform: '🎵 TikTok',    pkg: 'TikTok Starter',   time: '31 mins ago',  flag: '🇺🇬' },
  { platform: '📸 Instagram', pkg: 'Business Boost',   time: '45 mins ago',  flag: '🇰🇪' },
  { platform: '▶️ YouTube',   pkg: 'Kickstart',         time: '1 hr ago',     flag: '🇹🇿' },
]

export default function SocialProof() {
  return (
    <section className="py-20 px-4 overflow-hidden">
      <div className="max-w-6xl mx-auto">
        <div className="text-center mb-12">
          <h2 className="text-3xl sm:text-4xl font-black mb-4">
            Live <span className="gradient-text">Activity Feed</span>
          </h2>
          <p className="text-slate-400">Recent orders delivered across East Africa</p>
        </div>

        {/* Scrolling ticker */}
        <div className="relative overflow-hidden">
          <div className="flex gap-4 animate-[scroll_25s_linear_infinite]">
            {[...proofs, ...proofs].map((p, i) => (
              <div
                key={i}
                className="flex-shrink-0 glass rounded-xl px-5 py-3.5 flex items-center gap-3 min-w-[220px]"
              >
                <div className="w-9 h-9 rounded-lg bg-green-500/10 border border-green-500/20 flex items-center justify-center text-green-400 text-lg">
                  ✓
                </div>
                <div>
                  <p className="text-sm font-semibold text-white">{p.platform} {p.pkg}</p>
                  <p className="text-xs text-slate-500">{p.flag} Delivered {p.time}</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Fade edges */}
        <div className="pointer-events-none absolute left-0 w-24 bg-gradient-to-r from-dark to-transparent h-20" />
        <div className="pointer-events-none absolute right-0 w-24 bg-gradient-to-l from-dark to-transparent h-20" />
      </div>

      <style jsx>{`
        @keyframes scroll {
          0% { transform: translateX(0); }
          100% { transform: translateX(-50%); }
        }
      `}</style>
    </section>
  )
}
