export type Env = {
  DB: D1Database
  CORS_ORIGINS?: string
}

export type Person = { id: number; name: string }
export type Shop = { id: number; name: string }
export type Item = { id: number; name: string }

export type ExpenseShare = {
  person_id: number
  person_name?: string
  share: number
}

export type Expense = {
  id?: number
  person_id: number
  shop_id: number
  item_id: number
  date: string
  name: string
  amount: number
  person_name?: string
  shop_name?: string
  shares: ExpenseShare[]
}

export type ExpenseShareInput = {
  person_id: number
  share: number
}

export type MonthStats = {
  month: string
  armin: number
  ramin: number
  total: number
  armin_share: number
  ramin_share: number
}
