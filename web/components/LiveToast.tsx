'use client'

import { useEffect, useState, useCallback } from 'react'
import { SiTiktok, SiInstagram, SiYoutube } from 'react-icons/si'

const pool = [
  { platform: 'tiktok',    pkg: 'TikTok Quick-Start',    flag: '🇰🇪', city: 'Nairobi'      },
  { platform: 'instagram', pkg: 'IG Business Boost',      flag: '🇰🇪', city: 'Mombasa'      },
  { platform: 'tiktok',    pkg: 'Viral Creator Combo',    flag: '🇰🇪', city: 'Nairobi'      },
  { platform: 'youtube',   pkg: 'YouTube Kickstart',      flag: '🇹🇿', city: 'Dar es Salaam'},
  { platform: 'instagram', pkg: 'IG Celebrity Pack',      flag: '🇰🇪', city: 'Kisumu'       },
  { platform: 'tiktok',    pkg: 'TikTok Viral Starter',   flag: '🇺🇬', city: 'Kampala'      },
  { platform: 'instagram', pkg: 'Follower Booster',       flag: '🇰🇪', city: 'Nakuru'       },
  { platform: 'tiktok',    pkg: 'TikTok Starter',         flag: '🇺🇬', city: 'Entebbe'      },
  { platform: 'instagram', pkg: 'IG Quick-Start',         flag: '🇰🇪', city: 'Eldoret'      },
  { platform: 'youtube',   pkg: 'YouTube Kickstart',      flag: '🇹🇿', city: 'Arusha'       },
]

const platformIcon = {
  tiktok:    { Icon: SiTiktok,    bg: 'bg-black',                                                         color: 'text-white' },
  instagram: { Icon: SiInstagram, bg: 'bg-gradient-to-br from-[#833ab4] via-[#fd1d1d] to-[#fcb045]',     color: 'text-white' },
  youtube:   { Icon: SiYoutube,   bg: 'bg-[#FF0000]',                                                     color: 'text-white' },
}

interface Toast {
  id: number
  data: typeof pool[0]
  visible: boolean
}

let toastCounter = 0

export default function LiveToast() {
  const [toast, setToast] = useState<Toast | null>(null)

  const showNext = useCallback(() => {
    const data = pool[Math.floor(Math.random() * pool.length)]
    const id = ++toastCounter
    setToast({ id, data, visible: true })

    // Start hide animation after 3.5s
    setTimeout(() => {
      setToast(t => t?.id === id ? { ...t, visible: false } : t)
    }, 3500)

    // Remove from DOM after animation
    setTimeout(() => {
      setToast(t => t?.id === id ? null : t)
    }, 4200)
  }, [])

  useEffect(() => {
    // First toast after 4s, then every 9-14s
    const first = setTimeout(showNext, 4000)
    let interval: ReturnType<typeof setInterval>

    const schedule = () => {
      const delay = 9000 + Math.random() * 5000
      interval = setInterval(showNext, delay)
    }

    const scheduleTimer = setTimeout(schedule, 4000)

    return () => {
      clearTimeout(first)
      clearTimeout(scheduleTimer)
      clearInterval(interval)
    }
  }, [showNext])

  if (!toast) return null

  const cfg = platformIcon[toast.data.platform as keyof typeof platformIcon]
  const { Icon } = cfg

  return (
    <div
      className="fixed bottom-6 left-6 z-50 pointer-events-none"
      style={{
        animation: toast.visible
          ? 'toastIn 0.5s cubic-bezier(0.16,1,0.3,1) forwards'
          : 'toastOut 0.4s ease-in forwards',
      }}
    >
      <div className="flex items-center gap-3 glass rounded-2xl px-4 py-3 shadow-2xl shadow-black/40 border border-white/10 max-w-[280px]">
        {/* Platform icon */}
        <div className={`w-10 h-10 rounded-xl ${cfg.bg} flex items-center justify-center flex-shrink-0 shadow-md`}>
          <Icon size={16} className={cfg.color} />
        </div>

        {/* Text */}
        <div className="min-w-0">
          <div className="flex items-center gap-1.5 mb-0.5">
            <span className="w-2 h-2 rounded-full bg-green-400 flex-shrink-0" />
            <p className="text-xs font-bold text-white truncate">{toast.data.pkg}</p>
          </div>
          <p className="text-xs text-slate-400">{toast.data.flag} {toast.data.city} · just now</p>
        </div>
      </div>
    </div>
  )
}
