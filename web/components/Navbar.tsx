'use client'

import Link from 'next/link'
import { useState } from 'react'
import { Menu, X } from 'lucide-react'

export default function Navbar() {
  const [open, setOpen] = useState(false)

  return (
    <nav className="fixed top-0 left-0 right-0 z-50 glass border-b border-white/5">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
        {/* Logo */}
        <Link href="/" className="flex items-center">
          <span className="font-black text-xl tracking-tight text-white">
            Vector<span className="gradient-text">Boost</span>
          </span>
        </Link>

        {/* Desktop nav */}
        <div className="hidden md:flex items-center gap-8">
          <a href="#packages" className="text-sm text-slate-400 hover:text-white transition-colors">Packages</a>
          <a href="#how-it-works" className="text-sm text-slate-400 hover:text-white transition-colors">How It Works</a>
          <a
            href="#packages"
            className="px-5 py-2 rounded-lg text-sm font-bold text-white bg-gradient-brand hover:opacity-90 transition-opacity shadow-lg shadow-violet-900/30"
          >
            Get Started
          </a>
        </div>

        {/* Mobile hamburger */}
        <button
          className="md:hidden text-slate-400 hover:text-white transition-colors"
          onClick={() => setOpen(!open)}
          aria-label="Toggle menu"
        >
          {open ? <X size={22} /> : <Menu size={22} />}
        </button>
      </div>

      {/* Mobile menu */}
      {open && (
        <div className="md:hidden border-t border-white/5 bg-[#0a0a12] px-4 py-5 flex flex-col gap-4">
          <a href="#packages" className="text-slate-300 text-sm font-medium" onClick={() => setOpen(false)}>Packages</a>
          <a href="#how-it-works" className="text-slate-300 text-sm font-medium" onClick={() => setOpen(false)}>How It Works</a>
          <a
            href="#packages"
            className="px-5 py-3 rounded-lg font-bold text-white text-center bg-gradient-brand hover:opacity-90 transition-opacity"
            onClick={() => setOpen(false)}
          >
            Get Started
          </a>
        </div>
      )}
    </nav>
  )
}
