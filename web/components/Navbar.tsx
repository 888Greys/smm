'use client'

import Link from 'next/link'
import { useState } from 'react'
import { Menu, X, Zap } from 'lucide-react'

export default function Navbar() {
  const [open, setOpen] = useState(false)

  return (
    <nav className="fixed top-0 left-0 right-0 z-50 glass border-b border-border-dim">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-gradient-brand flex items-center justify-center">
            <Zap size={16} className="text-white" />
          </div>
          <span className="font-bold text-lg tracking-tight">
            Inn<span className="gradient-text">Bucks</span>
          </span>
        </Link>

        {/* Desktop nav */}
        <div className="hidden md:flex items-center gap-8">
          <a href="#packages" className="text-sm text-slate-400 hover:text-white transition-colors">Packages</a>
          <a href="#how-it-works" className="text-sm text-slate-400 hover:text-white transition-colors">How It Works</a>
          <a href="#packages" className="btn-primary text-sm">
            Get Started
          </a>
        </div>

        {/* Mobile hamburger */}
        <button
          className="md:hidden text-slate-400 hover:text-white"
          onClick={() => setOpen(!open)}
        >
          {open ? <X size={22} /> : <Menu size={22} />}
        </button>
      </div>

      {/* Mobile menu */}
      {open && (
        <div className="md:hidden border-t border-border-dim bg-surface px-4 py-4 flex flex-col gap-4">
          <a href="#packages" className="text-slate-300" onClick={() => setOpen(false)}>Packages</a>
          <a href="#how-it-works" className="text-slate-300" onClick={() => setOpen(false)}>How It Works</a>
          <a href="#packages" className="btn-primary text-center" onClick={() => setOpen(false)}>Get Started</a>
        </div>
      )}

      <style jsx>{`
        .btn-primary {
          background: linear-gradient(135deg, #7c3aed, #06b6d4);
          color: white;
          padding: 0.5rem 1.25rem;
          border-radius: 0.5rem;
          font-weight: 600;
          transition: opacity 0.2s;
        }
        .btn-primary:hover { opacity: 0.9; }
      `}</style>
    </nav>
  )
}
