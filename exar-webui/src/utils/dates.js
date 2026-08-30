import { isValidJalaaliDate, toGregorian, toJalaali } from 'jalaali-js'

function pad(n) {
  return String(n).padStart(2, '0')
}

/** Convert Gregorian YYYY-MM-DD to Jalali display YYYY/MM/DD */
export function toJalaliDisplay(gregorian) {
  if (!gregorian) return ''
  const [y, m, d] = gregorian.split('-').map(Number)
  if (!y || !m || !d) return ''
  const j = toJalaali(y, m, d)
  return `${j.jy}/${pad(j.jm)}/${pad(j.jd)}`
}

/**
 * Convert Jalali YYYY/MM/DD (or YYYY-MM-DD) to Gregorian YYYY-MM-DD.
 * Returns null if invalid.
 */
export function jalaliToGregorian(jalaliStr) {
  if (!jalaliStr) return null
  const cleaned = String(jalaliStr).trim().replace(/-/g, '/')
  const parts = cleaned.split('/').map(Number)
  if (parts.length !== 3) return null
  const [jy, jm, jd] = parts
  if (!isValidJalaaliDate(jy, jm, jd)) return null
  const g = toGregorian(jy, jm, jd)
  return `${g.gy}-${pad(g.gm)}-${pad(g.gd)}`
}

export function todayJalali() {
  const now = new Date()
  const j = toJalaali(now.getFullYear(), now.getMonth() + 1, now.getDate())
  return `${j.jy}/${pad(j.jm)}/${pad(j.jd)}`
}

export function todayGregorian() {
  const now = new Date()
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
}
