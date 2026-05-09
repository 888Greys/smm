import { ShoppingBag, TrendingUp } from 'lucide-react'

function MPesaStep() {
  return (
    <div className="relative w-20 h-20 rounded-2xl bg-[#00A651] flex items-center justify-center shadow-lg shadow-[#00A651]/25 flex-shrink-0">
      {/* Safaricom / M-Pesa styled icon */}
      <svg width="36" height="36" viewBox="0 0 36 36" fill="none" xmlns="http://www.w3.org/2000/svg">
        <circle cx="18" cy="18" r="14" fill="rgba(255,255,255,0.15)" />
        <path d="M11 13h2l3 7 3-7h2v10h-2v-6l-2.5 6h-1L13 17v6h-2V13Z" fill="white" />
        <path d="M23 13h2v10h-2V13Z" fill="white" opacity="0.7" />
      </svg>
      <span className="absolute -top-2 -right-2 w-6 h-6 rounded-full bg-dark border border-border-dim text-xs font-black text-slate-400 flex items-center justify-center">2</span>
    </div>
  )
}

const steps = [
  {
    icon: ShoppingBag,
    color: 'from-violet-500 to-violet-700',
    shadowColor: 'shadow-violet-900/30',
    step: '01',
    title: 'Choose a Package',
    desc: 'Pick the growth package that fits your goal and budget. From entry-level KES 500 to power combos at KES 2,500.',
  },
  {
    icon: null, // custom M-Pesa step
    color: '',
    shadowColor: '',
    step: '02',
    title: 'Pay with M-Pesa',
    desc: "Enter your profile link and Safaricom number. You'll receive an M-Pesa STK push — just enter your PIN to confirm.",
  },
  {
    icon: TrendingUp,
    color: 'from-green-500 to-emerald-600',
    shadowColor: 'shadow-green-900/30',
    step: '03',
    title: 'Watch Your Account Grow',
    desc: 'Delivery starts within minutes. Track your order in real-time. Followers drip in naturally — no drops.',
  },
]

export default function HowItWorks() {
  return (
    <section id="how-it-works" className="py-24 px-4 relative">
      <div className="absolute inset-0 bg-gradient-to-b from-transparent via-violet-900/4 to-transparent pointer-events-none" />

      <div className="max-w-6xl mx-auto relative">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-black mb-4">
            How It <span className="gradient-text">Works</span>
          </h2>
          <p className="text-slate-400 max-w-xl mx-auto">
            Three simple steps from zero to growing — fully automated, no middleman.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-10 relative">
          {/* Connector line */}
          <div className="hidden md:block absolute top-10 left-[22%] right-[22%] h-px bg-gradient-to-r from-violet-600/30 via-[#00A651]/30 to-green-500/30" />

          {steps.map((step, i) => {
            const Icon = step.icon
            return (
              <div key={i} className="flex flex-col items-center text-center gap-5 relative">
                {/* Icon */}
                {i === 1 ? (
                  <MPesaStep />
                ) : (
                  Icon && (
                    <div className={`relative w-20 h-20 rounded-2xl bg-gradient-to-br ${step.color} flex items-center justify-center shadow-lg ${step.shadowColor}`}>
                      <Icon size={32} className="text-white" />
                      <span className="absolute -top-2 -right-2 w-6 h-6 rounded-full bg-dark border border-border-dim text-xs font-black text-slate-400 flex items-center justify-center">
                        {i + 1}
                      </span>
                    </div>
                  )
                )}

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
