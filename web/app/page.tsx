import { getPackages } from '@/lib/api'
import { Package } from '@/lib/types'
import Navbar from '@/components/Navbar'
import Hero from '@/components/Hero'
import PackagesSection from '@/components/PackagesSection'
import HowItWorks from '@/components/HowItWorks'
import SocialProof from '@/components/SocialProof'
import Footer from '@/components/Footer'
import LiveToast from '@/components/LiveToast'
import CursorGlow from '@/components/CursorGlow'
import ScrollProgress from '@/components/ScrollProgress'
import RevealOnScroll from '@/components/RevealOnScroll'

export const revalidate = 300

export default async function Home() {
  let packages: Package[] = []
  try {
    packages = await getPackages()
  } catch {
    // API offline during build — client will fetch
  }

  return (
    <main className="min-h-screen bg-dark">
      {/* Fixed overlays */}
      <ScrollProgress />
      <CursorGlow />
      <LiveToast />

      <Navbar />
      <Hero />

      <RevealOnScroll delay={0}>
        <PackagesSection initialPackages={packages} />
      </RevealOnScroll>

      <RevealOnScroll delay={80}>
        <HowItWorks />
      </RevealOnScroll>

      <RevealOnScroll delay={0}>
        <SocialProof />
      </RevealOnScroll>

      <RevealOnScroll delay={0}>
        <Footer />
      </RevealOnScroll>
    </main>
  )
}
