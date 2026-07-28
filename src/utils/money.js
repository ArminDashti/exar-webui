export function formatMoney(value) {
  const n = Number(value) || 0
  return new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 2,
    minimumFractionDigits: 0,
  }).format(n)
}
