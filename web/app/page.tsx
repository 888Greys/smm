import { getPackages } from '@/lib/api'
import { Package } from '@/lib/types'
import Navbar from '@/components/Navbar'
import Hero from '@/components/Hero'
import PackagesSection from '@/components/PackagesSection'
import HowItWorks from '@/components/HowItWorks'
import SocialProof from '@/components/SocialProof'
import Footer from '@/components/Footer'

export const revalidate = 300

export default async function Home() {
  let packages: Package[] = []
  try {
    packages = await getPackages()
  } catch {
    // API offline during build — render with empty packages, client will fetch
  }

  return (
    <main className="min-h-screen bg-dark">
      <Navbar />
      <Hero />
      <PackagesSection initialPackages={packages} />
      <HowItWorks />
      <SocialProof />
      <Footer />
    </main>
  )
}
