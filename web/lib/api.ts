import { Package, CreateOrderResponse, OrderStatus } from './types'

const API = process.env.NEXT_PUBLIC_API_URL || 'https://api.innbucks.org'

export async function getPackages(): Promise<Package[]> {
  const res = await fetch(`${API}/api/packages`, { next: { revalidate: 300 } })
  if (!res.ok) throw new Error('Failed to load packages')
  return res.json()
}

export async function createOrder(data: {
  package_id: string
  profile_link: string
  phone: string
  referral_code?: string
}): Promise<CreateOrderResponse> {
  const res = await fetch(`${API}/api/orders`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return res.json()
}

export async function getOrderStatus(orderId: number): Promise<OrderStatus> {
  const res = await fetch(`${API}/api/orders/${orderId}`, { cache: 'no-store' })
  return res.json()
}
