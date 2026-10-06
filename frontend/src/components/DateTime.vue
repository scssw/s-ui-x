<template>
  <v-text-field
    :label="$t('date.expiry')"
    v-model="manualDate"
    placeholder="26.11.6"
    @change="parseManualDate"
    @click:clear="clearExpiry"
    clearable
    hide-details
  ></v-text-field>
</template>

<script lang="ts">
import { i18n, locale } from '@/locales'

export default {
  props: ['expiry'],
  emits: ['submit'],
  data() {
    return {
      manualDate: '',
    }
  },
  created() {
    this.formatManualDate()
  },
  computed: {
    locale() {
      return locale
    },
    expDate() {
      return parseInt(this.expiry?? 0)
    },
  },
  methods: {
    formatManualDate() {
      if (this.expDate === 0) { this.manualDate = i18n.global.t('unlimited'); return }
      const date = new Date(this.expDate * 1000)
      this.manualDate = `${date.getFullYear()}.${date.getMonth()+1}.${date.getDate()}`
    },
    parseManualDate() {
      const value = this.manualDate.trim()
      if (!value) {
        this.$emit('submit', 0)
        this.formatManualDate()
        return
      }
      const match = value.match(/^(\d{2,4})[.\-/](\d{1,2})[.\-/](\d{1,2})$/)
      if (!match) { this.formatManualDate(); return }
      let year = Number(match[1])
      if (year < 100) year += 2000
      const month = Number(match[2])
      const day = Number(match[3])
      const now = new Date()
      const date = new Date(year, month - 1, day, now.getHours(), now.getMinutes(), now.getSeconds())
      if (date.getFullYear() !== year || date.getMonth() !== month - 1 || date.getDate() !== day) { this.formatManualDate(); return }
      this.$emit('submit', Math.floor(date.getTime() / 1000))
      this.formatManualDate()
    },
    clearExpiry() {
      this.$emit('submit',0)
    }
  },
  watch: {
    expiry() { this.formatManualDate() },
  }
}
</script>
