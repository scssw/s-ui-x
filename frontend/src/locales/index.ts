import { createI18n } from 'vue-i18n'

type LocaleCode = 'zhHans'
type LocaleMessages = Record<string, unknown>

const DEFAULT_LOCALE: LocaleCode = 'zhHans'

const localeLoaders: Record<LocaleCode, () => Promise<{ default: LocaleMessages }>> = {
  zhHans: () => import('./zhcn'),
}

const loadedLocales = new Set<LocaleCode>()

const normalizeLocale = (_value?: string | null): LocaleCode => DEFAULT_LOCALE
const initialLocale = DEFAULT_LOCALE

export const i18n = createI18n({
  legacy: false,
  locale: initialLocale,
  fallbackLocale: DEFAULT_LOCALE,
  messages: {},
})

const loadMessages = async (localeCode: LocaleCode) => {
  if (loadedLocales.has(localeCode)) {
    return
  }
  const messages = await localeLoaders[localeCode]()
  i18n.global.setLocaleMessage(localeCode, messages.default)
  loadedLocales.add(localeCode)
}

export const loadLocaleMessages = async (localeCode: string) => {
  const normalized = normalizeLocale(localeCode)
  await loadMessages(DEFAULT_LOCALE)
  if (normalized !== DEFAULT_LOCALE) {
    await loadMessages(normalized)
  }
  return normalized
}

export const loadInitialLocaleMessages = () => loadLocaleMessages(initialLocale)

export const setI18nLocale = async (localeCode: string) => {
  const normalized = await loadLocaleMessages(localeCode)
  i18n.global.locale.value = normalized
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('locale', normalized)
  }
  return normalized
}

export const locale = (() => {
  return 'zh-cn'
})()

export const languages = [{ title: '简体中文', value: 'zhHans' }]
