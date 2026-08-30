export function formatMoney(value) {
  const n = Math.round(Number(value) || 0)
  return new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 0,
    minimumFractionDigits: 0,
  }).format(n)
}
