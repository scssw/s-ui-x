import RandomUtil from "@/plugins/randomUtil"

export interface Link {
  type: "local" | "external" | "sub"
  remark?: string
  uri: string
}

export interface Client {
  id?: number
	enable: boolean
	name: string
  subSecret?: string
	config?: Config
	inbounds: number[]
  links?: Link[]
	volume: number
	expiry: number
  up: number
  down: number
  desc: string
  group: string
  limitIp?: number
  ipLimitMode?: 'monitor' | 'enforce'
  lastOnline?: number
  lastIpCount?: number
  delayStart?: boolean
  autoReset?: boolean
  resetDays?: number
  nextReset?: number
  totalUp?: number
  totalDown?: number
}

const defaultClient: Client = {
  enable: true,
  name: "",
  config: {},
  inbounds: [],
  links: [],
  volume: 0,
  expiry: 0,
  up: 0,
  down: 0,
  desc: "",
  group: "",
  limitIp: 0,
  ipLimitMode: 'monitor',
  lastOnline: 0,
  lastIpCount: 0,
  delayStart: false,
  autoReset: false,
  resetDays: 0,
  nextReset: 0,
  totalUp: 0,
  totalDown: 0,
}

type Config = {
  [key: string]: {
    name?: string
    username?: string
    [key: string]: any
  }
}

export function updateConfigs(configs: Config, newUserName: string): Config {
  for (const key in configs) {
    if (configs.hasOwnProperty(key)) {
      const config = configs[key]
      if (config.hasOwnProperty("name")) {
        config.name = newUserName
      } else if (config.hasOwnProperty("username")) {
        config.username = newUserName
      }
    }
  }
  return configs
}

export function shuffleConfigs(configs: Config, key?: string) {
  const keys = key ? [key] : Object.keys(configs)
  keys.forEach(k => {
    switch (k) {
      case "mixed":
      case "socks":
      case "http":
      case "anytls":
      case "trojan":
      case "naive":
      case "hysteria2":
        configs[k].password = RandomUtil.randomSeq(10)
        break
      case "shadowsocks":
        configs[k].password = RandomUtil.randomShadowsocksPassword(32)
        break
      case "shadowsocks16":
        configs[k].password = RandomUtil.randomShadowsocksPassword(16)
        break
      case "shadowtls":
        configs[k].password = RandomUtil.randomShadowsocksPassword(32)
        break
      case "hysteria":
        configs[k].auth_str = RandomUtil.randomSeq(10)
        break
      case "tuic":
        configs[k].password = RandomUtil.randomSeq(10)
        configs[k].uuid = RandomUtil.randomUUID()
        break
      case "vmess":
      case "vless":
        configs[k].uuid = RandomUtil.randomUUID()
        break
    }
  })
}

export function randomConfigs(user: string): Config {
  const mixedPassword = RandomUtil.randomSeq(10)
  const ssPassword16 = RandomUtil.randomShadowsocksPassword(16)
  const ssPassword32 = RandomUtil.randomShadowsocksPassword(32)
  const uuid = RandomUtil.randomUUID()
  return {
    mixed: {
      username: user,
      password: mixedPassword,
    },
    socks: {
      username: user,
      password: mixedPassword,
    },
    http: {
      username: user,
      password: mixedPassword,
    },
    shadowsocks: {
      name: user,
      password: ssPassword32,
    },
    shadowsocks16: {
      name: user,
      password: ssPassword16,
    },
    shadowtls: {
      name: user,
      password: ssPassword32,
    },
    vmess: {
      name: user,
      uuid: uuid,
      alterId: 0,
    },
    vless: {
      name: user,
      uuid: uuid,
      flow: "xtls-rprx-vision",
    },
    anytls: {
      name: user,
      password: mixedPassword,
    },
    trojan: {
      name: user,
      password: mixedPassword,
    },
    naive: {
      username: user,
      password: mixedPassword,
    },
    hysteria: {
      name: user,
      auth_str: mixedPassword,
    },
    tuic: {
      name: user,
      uuid: uuid,
      password: mixedPassword,
    },
    hysteria2: {
      name: user,
      password: mixedPassword,
    },
  }
}

export function createClient<T extends Client>(json?: Partial<T>): Client {
  defaultClient.name = RandomUtil.randomSeq(8)
  const defaultObject: Client = { ...defaultClient, ...(json || {}) }

  // Add missing config
  defaultObject.config = { ...randomConfigs(defaultObject.name), ...defaultObject.config }
  
  return defaultObject
}

export function addMonths(date: Date, months: number): Date {
  const result = new Date(date)
  const targetYear = result.getFullYear() + Math.floor((result.getMonth() + months) / 12)
  const targetMonth = (result.getMonth() + months) % 12
  const day = result.getDate()
  const daysInTargetMonth = new Date(targetYear, targetMonth + 1, 0).getDate()
  result.setFullYear(targetYear, targetMonth, Math.min(day, daysInTargetMonth))
  return result
}

export function getMonthlyReset(startDate: Date = new Date()) {
  const year = startDate.getFullYear()
  const month = startDate.getMonth()
  const day = startDate.getDate()

  const nextMonth = (month + 1) % 12
  const nextYear = month === 11 ? year + 1 : year
  const daysInNextMonth = new Date(nextYear, nextMonth + 1, 0).getDate()

  let nextResetDate: Date
  let resetDays: number

  if (day <= daysInNextMonth) {
    nextResetDate = new Date(startDate)
    nextResetDate.setFullYear(nextYear, nextMonth, day)
    resetDays = Math.round((nextResetDate.getTime() - startDate.getTime()) / (24 * 3600 * 1000))
  } else {
    resetDays = 30
    nextResetDate = new Date(startDate.getTime() + resetDays * 24 * 3600 * 1000)
  }

  return {
    resetDays,
    nextReset: Math.floor(nextResetDate.getTime() / 1000),
    nextResetDate,
  }
}

