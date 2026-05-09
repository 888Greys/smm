import { Zap } from 'lucide-react'
import Link from 'next/link'

export default function Footer() {
  return (
    <footer className="border-t border-border-dim bg-surface py-12 px-4">
      <div className="max-w-6xl mx-auto">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-10 mb-10">
          {/* Brand */}
          <div>
            <div className="flex items-center gap-2 mb-4">
              <div className="w-8 h-8 rounded-lg bg-gradient-brand flex items-center justify-center">
                <Zap size={16} className="text-white" />
              </div>
              <span className="font-bold text-lg">
                Inn<span className="gradient-text">Bucks</span> SMM
              </span>
            </div>
            <p className="text-slate-500 text-sm leading-relaxed">
              The fastest way to grow your social media presence in Kenya. Real engagement, real results.
            </p>
          </div>

          {/* Quick links */}
          <div>
            <h4 className="text-sm font-semibold text-white mb-4">Quick Links</h4>
            <ul className="space-y-2 text-sm text-slate-500">
              <li><a href="#packages" className="hover:text-violet-300 transition-colors">Packages</a></li>
              <li><a href="#how-it-works" className="hover:text-violet-300 transition-colors">How It Works</a></li>
              <li>
                <a href="https://t.me/pompomputrin888pom_bot" target="_blank" rel="noopener noreferrer" className="hover:text-violet-300 transition-colors">
                  Telegram Bot
                </a>
              </li>
              <li>
                <a href="https://t.me/workratew" target="_blank" rel="noopener noreferrer" className="hover:text-violet-300 transition-colors">
                  Support
                </a>
              </li>
            </ul>
          </div>

          {/* Contact */}
          <div>
            <h4 className="text-sm font-semibold text-white mb-4">Support</h4>
            <ul className="space-y-2 text-sm text-slate-500">
              <li>Telegram: <a href="https://t.me/workratew" className="text-violet-400 hover:text-violet-300">@workratew</a></li>
              <li>Response time: Within 1 hour</li>
              <li className="text-xs text-slate-600 mt-3">
                Orders are non-refundable once placed. Ensure your profile is public before ordering.
              </li>
            </ul>
          </div>
        </div>

        <div className="border-t border-border-dim pt-6 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-slate-600">
          <p>© {new Date().getFullYear()} InnBucks SMM. All rights reserved.</p>
          <p>Powered by M-Pesa &amp; SMMWiz</p>
        </div>
      </div>
    </footer>
  )
}
