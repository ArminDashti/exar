const gdm = [0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334]
const jMonthLen = [31, 31, 31, 31, 31, 31, 30, 30, 30, 30, 30, 29]

export function toJalali(gy: number, gm: number, gd: number): [number, number, number] {
  const gy2 = gy - 1600
  const gm2 = gm - 1
  const gd2 = gd - 1

  let gDayNo = 365 * gy2 + Math.floor((gy2 + 3) / 4) - Math.floor((gy2 + 99) / 100) + Math.floor((gy2 + 399) / 400)
  gDayNo += gdm[gm2] + gd2
  if (gm > 2 && ((gy % 4 === 0 && gy % 100 !== 0) || gy % 400 === 0)) {
    gDayNo++
  }

  let jDayNo = gDayNo - 79
  const jNp = Math.floor(jDayNo / 12053)
  jDayNo %= 12053
  let jy = 979 + 33 * jNp + 4 * Math.floor(jDayNo / 1461)
  jDayNo %= 1461

  if (jDayNo >= 366) {
    jy += Math.floor((jDayNo - 1) / 365)
    jDayNo = (jDayNo - 1) % 365
  }

  let i = 0
  for (; i < 11 && jDayNo >= jMonthLen[i]; i++) {
    jDayNo -= jMonthLen[i]
  }
  return [jy, i + 1, jDayNo + 1]
}

export function monthKeyFromGregorian(date: string): string | null {
  const m = /^(\d+)-(\d+)-(\d+)$/.exec(date.trim())
  if (!m) return null
  const gy = Number(m[1])
  const gm = Number(m[2])
  const gd = Number(m[3])
  const [jy, jm] = toJalali(gy, gm, gd)
  return `${String(jy).padStart(4, '0')}/${String(jm).padStart(2, '0')}`
}
