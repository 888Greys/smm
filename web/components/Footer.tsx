import Link from 'next/link'
import { SiTelegram, SiTiktok, SiInstagram, SiYoutube } from 'react-icons/si'

function MPesaLogo() {
  return (
    <span className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-[#00A651]/10 border border-[#00A651]/20 text-[#00A651] text-xs font-black tracking-wide">
      <span className="w-2 h-2 rounded-full bg-[#00A651]" />
      M-PESA
    </span>
  )
}

export default function Footer() {
  return (
    <footer className="border-t border-white/5 bg-[#07070a] py-14 px-4">
      <div className="max-w-6xl mx-auto">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-10 mb-10">
          {/* Brand */}
          <div className="md:col-span-2">
            <div className="flex items-center gap-2.5 mb-4">
              <div className="w-8 h-8 rounded-lg bg-gradient-brand flex items-center justify-center shadow-lg shadow-violet-900/40">
                <svg width="15" height="15" viewBox="0 0 15 15" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M8.5 1L2 9h6l-1.5 5L14 6H8L8.5 1Z" fill="white" />
                </svg>
              </div>
              <span className="font-black text-lg text-white">
                Inn<span className="gradient-text">Bucks</span>
              </span>
            </div>
            <p className="text-slate-500 text-sm leading-relaxed max-w-xs mb-5">
              The fastest way to grow your social media in Kenya. Real followers, real engagement, safe delivery.
            </p>
            {/* Platform icons */}
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-black flex items-center justify-center border border-white/10">
                <SiTiktok size={14} className="text-white" />
              </div>
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-[#833ab4] via-[#fd1d1d] to-[#fcb045] flex items-center justify-center">
                <SiInstagram size={14} className="text-white" />
              </div>
              <div className="w-8 h-8 rounded-lg bg-[#FF0000] flex items-center justify-center">
                <SiYoutube size={14} className="text-white" />
              </div>
              <a
                href="https://t.me/pompomputrin888pom_bot"
                target="_blank"
                rel="noopener noreferrer"
                className="w-8 h-8 rounded-lg bg-[#2AABEE] flex items-center justify-center hover:opacity-80 transition-opacity"
              >
                <SiTelegram size={14} className="text-white" />
              </a>
            </div>
          </div>

          {/* Quick links */}
          <div>
            <h4 className="text-xs font-bold text-slate-400 uppercase tracking-widest mb-4">Navigation</h4>
            <ul className="space-y-2.5 text-sm text-slate-500">
              <li><a href="#packages" className="hover:text-violet-300 transition-colors">Packages</a></li>
              <li><a href="#how-it-works" className="hover:text-violet-300 transition-colors">How It Works</a></li>
              <li>
                <a href="https://t.me/pompomputrin888pom_bot" target="_blank" rel="noopener noreferrer" className="hover:text-violet-300 transition-colors">
                  Telegram Bot
                </a>
              </li>
            </ul>
          </div>

          {/* Support */}
          <div>
            <h4 className="text-xs font-bold text-slate-400 uppercase tracking-widest mb-4">Support</h4>
            <ul className="space-y-2.5 text-sm text-slate-500">
              <li>
                <a href="https://t.me/workratew" className="flex items-center gap-2 hover:text-violet-300 transition-colors group" target="_blank" rel="noopener noreferrer">
                  <SiTelegram size={13} className="text-[#2AABEE] group-hover:scale-110 transition-transform" />
                  @workratew
                </a>
              </li>
              <li className="text-slate-600">Response: within 1 hour</li>
              <li className="mt-4">
                <MPesaLogo />
              </li>
            </ul>
          </div>
        </div>

        <div className="border-t border-white/5 pt-6 flex flex-col sm:flex-row items-center justify-between gap-3">
          <p className="text-xs text-slate-600">© {new Date().getFullYear()} InnBucks. All rights reserved.</p>
          <p className="text-xs text-slate-700">Secure payments via M-Pesa STK Push</p>
        </div>
      </div>
    </footer>
  )
}
