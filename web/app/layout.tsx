import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
  display: 'swap',
})

export const metadata: Metadata = {
  title: 'VectorBoost — Grow Your Social Media',
  description: 'Real followers, likes & views for TikTok, Instagram and YouTube. Fast delivery, affordable prices. Pay with M-Pesa.',
  keywords: 'SMM panel Kenya, buy followers Kenya, TikTok followers Nairobi, Instagram followers Kenya, M-Pesa social media',
  openGraph: {
    title: 'VectorBoost — Grow Your Social Media',
    description: 'Real followers, likes & views. Pay with M-Pesa. Fast delivery.',
    url: 'https://innbucks.org',
    siteName: 'VectorBoost',
    type: 'website',
  },
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={inter.variable}>
      <body>{children}</body>
    </html>
  )
}
