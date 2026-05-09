import { ShoppingCart, Smartphone, TrendingUp } from 'lucide-react'

const steps = [
  {
    icon: ShoppingCart,
    color: 'from-violet-500 to-violet-700',
    step: '01',
    title: 'Choose a Package',
    desc: 'Pick the growth package that fits your goal and budget. From entry-level KES 500 to power combos at KES 2,500.',
  },
  {
    icon: Smartphone,
    color: 'from-cyan-500 to-blue-600',
    step: '02',
    title: 'Pay with M-Pesa',
    desc: 'Enter your profile link and Safaricom number. You\'ll receive an M-Pesa STK push — just enter your PIN to confirm.',
  },
  {
    icon: TrendingUp,
    color: 'from-green-500 to-emerald-600',
    step: '03',
    title: 'Watch Your Account Grow',
    desc: 'Delivery starts within minutes. Track your order status in real-time. Selected packages include a 30-day refill guarantee.',
  },
]

export default function HowItWorks() {
  return (
    <section id="how-it-works" className="py-24 px-4 relative">
      {/* Subtle background */}
      <div className="absolute inset-0 bg-gradient-to-b from-transparent via-violet-900/5 to-transparent pointer-events-none" />

      <div className="max-w-6xl mx-auto relative">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-black mb-4">
            How It <span className="gradient-text">Works</span>
          </h2>
          <p className="text-slate-400 max-w-xl mx-auto">
            Three simple steps from zero to growing — all automated, no middleman.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 relative">
          {/* Connector line (desktop) */}
          <div className="hidden md:block absolute top-12 left-[20%] right-[20%] h-px bg-gradient-to-r from-violet-600/30 via-cyan-500/30 to-green-500/30" />

          {steps.map((step, i) => {
            const Icon = step.icon
            return (
              <div key={i} className="flex flex-col items-center text-center gap-4 relative">
                {/* Icon */}
                <div className={`relative w-20 h-20 rounded-2xl bg-gradient-to-br ${step.color} flex items-center justify-center shadow-lg`}>
                  <Icon size={32} className="text-white" />
                  <span className="absolute -top-2 -right-2 w-6 h-6 rounded-full bg-dark border border-border-dim text-xs font-black text-slate-400 flex items-center justify-center">
                    {i + 1}
                  </span>
                </div>

                <div>
                  <h3 className="text-lg font-bold text-white mb-2">{step.title}</h3>
                  <p className="text-slate-400 text-sm leading-relaxed max-w-xs mx-auto">{step.desc}</p>
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}
