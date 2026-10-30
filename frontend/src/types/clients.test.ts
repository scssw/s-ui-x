import { describe, expect, it } from 'vitest'
import { addMonths, getMonthlyReset } from './clients'

describe('client monthly reset and expiry helpers', () => {
  it('resets on Nov 30 (31 days) when created on Oct 30', () => {
    // 2026-10-30 15:00:00
    const oct30 = new Date(2026, 9, 30, 15, 0, 0)
    const reset = getMonthlyReset(oct30)

    expect(reset.resetDays).toBe(31)
    expect(reset.nextResetDate.getFullYear()).toBe(2026)
    expect(reset.nextResetDate.getMonth()).toBe(10) // November
    expect(reset.nextResetDate.getDate()).toBe(30)
    expect(reset.nextResetDate.getHours()).toBe(15)
  })

  it('resets by 30 days when created on Jan 31', () => {
    // 2026-01-31 10:00:00
    const jan31 = new Date(2026, 0, 31, 10, 0, 0)
    const reset = getMonthlyReset(jan31)

    expect(reset.resetDays).toBe(30)
    const expectedTime = Math.floor((jan31.getTime() + 30 * 24 * 3600 * 1000) / 1000)
    expect(reset.nextReset).toBe(expectedTime)
  })

  it('resets by 30 days when created on March 31 (next month April has 30 days)', () => {
    const mar31 = new Date(2026, 2, 31, 12, 0, 0)
    const reset = getMonthlyReset(mar31)

    expect(reset.resetDays).toBe(30)
  })

  it('resets on Feb 15 when created on Jan 15', () => {
    const jan15 = new Date(2026, 0, 15, 12, 0, 0)
    const reset = getMonthlyReset(jan15)

    expect(reset.resetDays).toBe(31)
    expect(reset.nextResetDate.getMonth()).toBe(1)
    expect(reset.nextResetDate.getDate()).toBe(15)
  })

  it('handles year rollover (Dec 20 -> Jan 20)', () => {
    const dec20 = new Date(2026, 11, 20, 12, 0, 0)
    const reset = getMonthlyReset(dec20)

    expect(reset.nextResetDate.getFullYear()).toBe(2027)
    expect(reset.nextResetDate.getMonth()).toBe(0)
    expect(reset.nextResetDate.getDate()).toBe(20)
  })

  describe('addMonths quick expiry', () => {
    it('correctly adds 1, 3, 6, 12 months for standard dates', () => {
      const base = new Date(2026, 9, 30, 12, 0, 0) // Oct 30

      const m1 = addMonths(base, 1) // Nov 30
      expect(m1.getFullYear()).toBe(2026)
      expect(m1.getMonth()).toBe(10)
      expect(m1.getDate()).toBe(30)

      const m3 = addMonths(base, 3) // Jan 30, 2027
      expect(m3.getFullYear()).toBe(2027)
      expect(m3.getMonth()).toBe(0)
      expect(m3.getDate()).toBe(30)

      const m6 = addMonths(base, 6) // Apr 30, 2027
      expect(m6.getFullYear()).toBe(2027)
      expect(m6.getMonth()).toBe(3)
      expect(m6.getDate()).toBe(30)

      const m12 = addMonths(base, 12) // Oct 30, 2027
      expect(m12.getFullYear()).toBe(2027)
      expect(m12.getMonth()).toBe(9)
      expect(m12.getDate()).toBe(30)
    })

    it('clamps to end of month when day exceeds target month days', () => {
      const jan31 = new Date(2026, 0, 31, 12, 0, 0)
      const feb = addMonths(jan31, 1)
      expect(feb.getMonth()).toBe(1)
      expect(feb.getDate()).toBe(28) // 2026 is non-leap year
    })
  })
})
