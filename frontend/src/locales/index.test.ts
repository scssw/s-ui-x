import { beforeEach, describe, expect, it, vi } from 'vitest'

const storage = new Map<string, string>()

const stubLocalStorage = () => {
  vi.stubGlobal('localStorage', {
    getItem: (key: string) => storage.get(key) ?? null,
    setItem: (key: string, value: string) => storage.set(key, value),
    removeItem: (key: string) => storage.delete(key),
    clear: () => storage.clear(),
  })
}

describe('locale loading', () => {
  beforeEach(() => {
    storage.clear()
    vi.resetModules()
    vi.unstubAllGlobals()
    stubLocalStorage()
  })

  it('loads simplified Chinese messages on startup', async () => {
    const { i18n, loadInitialLocaleMessages } = await import('./index')

    await loadInitialLocaleMessages()

    expect(i18n.global.availableLocales).toEqual(['zhHans'])
  })

  it('ignores an old stored locale and starts in simplified Chinese', async () => {
    storage.set('locale', 'ru')
    const { i18n, loadInitialLocaleMessages } = await import('./index')

    await loadInitialLocaleMessages()

    expect(i18n.global.availableLocales).toEqual(['zhHans'])
    expect(i18n.global.locale.value).toBe('zhHans')
  })

  it('loads and stores locales when changed', async () => {
    const { i18n, setI18nLocale } = await import('./index')

    const selectedLocale = await setI18nLocale('zhHans')

    expect(selectedLocale).toBe('zhHans')
    expect(storage.get('locale')).toBe('zhHans')
    expect(i18n.global.locale.value).toBe('zhHans')
    expect(i18n.global.availableLocales).toEqual(['zhHans'])
  })

  it('falls back to simplified Chinese for unsupported locales', async () => {
    const { i18n, setI18nLocale } = await import('./index')

    const selectedLocale = await setI18nLocale('missing')

    expect(selectedLocale).toBe('zhHans')
    expect(storage.get('locale')).toBe('zhHans')
    expect(i18n.global.locale.value).toBe('zhHans')
    expect(i18n.global.availableLocales).toEqual(['zhHans'])
  })
})
