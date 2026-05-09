export interface Package {
  id: string
  name: string
  platform: 'tiktok' | 'instagram' | 'youtube'
  price_kes: number
  description: string
}

export interface CreateOrderResponse {
  order_id: number
  message: string
  error?: string
}

export interface OrderStatus {
  order_id: number
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled'
  package_name: string
  platform: string
  description: string
  price_kes: number
  created_at: string
  error?: string
}
