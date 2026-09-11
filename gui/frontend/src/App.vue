<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue'

// --- State ---
const version = ref('')
const operation = ref<'split' | 'merge' | 'verify'>('split')
const filePath = ref('')
const chunkSize = ref('2G')
const hashAlgo = ref('md5')
const outputDir = ref('')
const mergeMode = ref('quick')
const forceOverwrite = ref(false)
const isDark = ref(true)
const isRunning = ref(false)
const progress = reactive({ percent: 0, label: '就绪', done: 0, total: 0 })
const logs = ref<{time: string, message: string, level: string}[]>([])
const isDragOver = ref(false)

const isSplit = computed(() => operation.value === 'split')
const isMerge = computed(() => operation.value === 'merge')
const isVerify = computed(() => operation.value === 'verify')

// --- Wails runtime ---
declare global {
  interface Window {
    go: {
      main: {
        App: {
          GetVersion(): Promise<string>
          Split(req: {source: string, size: string, hash: string, output: string}): Promise<Error>
          Merge(req: {manifest: string, mode: string, output: string, force: boolean}): Promise<Error>
          Verify(req: {manifest: string}): Promise<Error>
          WindowMinimise(): Promise<void>
          WindowClose(): Promise<void>
        }
      }
    }
    runtime: {
      EventsOn(name: string, cb: (data: any) => void): void
    }
  }
}

// --- Load saved config ---
function loadConfig() {
  try {
    const saved = localStorage.getItem('cutx-config')
    if (saved) {
      const c = JSON.parse(saved)
      if (c.chunkSize) chunkSize.value = c.chunkSize
      if (c.hashAlgo) hashAlgo.value = c.hashAlgo
      if (c.isDark !== undefined) isDark.value = c.isDark
    }
  } catch {}
}

function saveConfig() {
  localStorage.setItem('cutx-config', JSON.stringify({
    chunkSize: chunkSize.value,
    hashAlgo: hashAlgo.value,
    isDark: isDark.value,
  }))
}

// --- Theme toggle ---
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  saveConfig()
}

// --- File drop ---
function onDrop(e: DragEvent) {
  e.preventDefault()
  isDragOver.value = false
  // Wails handles the actual drop via OnFileDrop event
}
function onDragOver(e: DragEvent) {
  e.preventDefault()
  isDragOver.value = true
}
function onDragLeave() {
  isDragOver.value = false
}

// --- Wails event listeners ---
onMounted(() => {
  loadConfig()
  document.documentElement.classList.toggle('dark', isDark.value)

  // Get version
  window.go?.main?.App?.GetVersion().then((v: string) => {
    version.value = v
  }).catch(() => {
    version.value = 'v1.0.0'
  })

  // Listen for Wails events
  const wailsRuntime = (window as any).runtime
  if (wailsRuntime) {
    wailsRuntime.EventsOn('progress', (data: any) => {
      progress.percent = data.percent || 0
      progress.label = data.label || ''
      progress.done = data.done_bytes || 0
      progress.total = data.total_bytes || 0
    })
    wailsRuntime.EventsOn('log', (data: any) => {
      logs.value.push({ time: data.time, message: data.message, level: data.level })
      // Keep last 200 logs
      if (logs.value.length > 200) logs.value = logs.value.slice(-200)
    })
    wailsRuntime.EventsOn('file-drop', (path: string) => {
      filePath.value = path
    })
    wailsRuntime.EventsOn('operation-complete', (data: any) => {
      isRunning.value = false
      if (data.success) {
        progress.label = '完成'
        progress.percent = 100
      } else {
        progress.label = '失败'
      }
    })
  }
})

// --- Actions ---
async function startOperation() {
  if (isRunning.value) return
  if (!filePath.value) {
    logs.value.push({ time: new Date().toLocaleTimeString('zh-CN', {hour12:false}), message: '✗ 请先选择文件', level: 'error' })
    return
  }
  isRunning.value = true
  progress.percent = 0
  progress.label = '准备中...'
  logs.value = []

  try {
    if (isSplit.value) {
      await window.go.main.App.Split({
        source: filePath.value,
        size: chunkSize.value,
        hash: hashAlgo.value,
        output: outputDir.value,
      })
    } else if (isMerge.value) {
      await window.go.main.App.Merge({
        manifest: filePath.value,
        mode: mergeMode.value,
        output: outputDir.value,
        force: forceOverwrite.value,
      })
    } else if (isVerify.value) {
      await window.go.main.App.Verify({
        manifest: filePath.value,
      })
    }
  } catch (e: any) {
    isRunning.value = false
    logs.value.push({ time: new Date().toLocaleTimeString('zh-CN', {hour12:false}), message: '✗ ' + (e?.message || String(e)), level: 'error' })
  }
}

// --- Window controls ---
function minimiseWindow() {
  window.go?.main?.App?.WindowMinimise?.()
}
function closeWindow() {
  window.go?.main?.App?.WindowClose?.()
}

// --- Helpers ---
function formatBytes(b: number): string {
  if (b >= 1073741824) return (b / 1073741824).toFixed(1) + ' GB'
  if (b >= 1048576) return (b / 1048576).toFixed(1) + ' MB'
  if (b >= 1024) return (b / 1024).toFixed(1) + ' KB'
  return b + ' B'
}

// Save config when values change
watch([chunkSize, hashAlgo, isDark], saveConfig)
</script>

<template>
  <div class="h-screen flex flex-col" :style="{ background: 'var(--bg-base)' }">
    <!-- Custom Title Bar (frameless) -->
    <div class="titlebar-drag flex items-center justify-between px-4 py-2.5"
         :style="{ background: 'var(--bg-panel)', borderBottom: '1px solid var(--color-border)' }">
      <div class="flex items-center gap-2.5">
        <div class="w-7 h-7 rounded-lg flex items-center justify-center font-bold text-sm"
             style="background: var(--color-accent); color: var(--bg-base)">C</div>
        <span class="text-sm font-semibold" style="color: var(--color-text)">CutX</span>
        <span class="text-xs" style="color: var(--color-text-muted)">{{ version }}</span>
      </div>
      <div class="flex items-center gap-1 no-drag">
        <button @click="toggleTheme" class="p-1.5 rounded-lg transition-colors hover:bg-white/5"
                :title="isDark ? '切换浅色' : '切换深色'">
          <svg v-if="isDark" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-text-muted)">
            <circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.2 4.2l1.4 1.4M18.4 18.4l1.4 1.4M1 12h2M21 12h2M4.2 19.8l1.4-1.4M18.4 5.6l1.4-1.4"/>
          </svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-text-muted)">
            <path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z"/>
          </svg>
        </button>
        <button @click="minimiseWindow" class="p-1.5 rounded-lg transition-colors hover:bg-white/5" title="最小化">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-text-muted)"><line x1="5" y1="12" x2="19" y2="12"/></svg>
        </button>
        <button @click="closeWindow" class="p-1.5 rounded-lg transition-colors hover:bg-red-500/20" title="关闭">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-text-muted)"><line x1="6" y1="6" x2="18" y2="18"/><line x1="6" y1="18" x2="18" y2="6"/></svg>
        </button>
      </div>
    </div>

    <!-- Main Content -->
    <div class="flex-1 overflow-y-auto p-5 space-y-4">
      <!-- Card 1: Operation + File -->
      <div class="card">
        <div class="flex gap-2 mb-4">
          <button v-for="op in [
            {key:'split', label:'切割', icon:'⬇'},
            {key:'merge', label:'合并', icon:'⬆'},
            {key:'verify', label:'校验', icon:'✓'}
          ]" :key="op.key"
            @click="operation = op.key as any"
            class="px-4 py-2 rounded-lg text-sm font-medium transition-all"
            :style="operation === op.key
              ? { background: 'var(--color-accent)', color: 'var(--bg-base)' }
              : { background: 'var(--bg-base)', color: 'var(--color-text-muted)', border: '1px solid var(--color-border)' }">
            {{ op.icon }} {{ op.label }}
          </button>
        </div>

        <div class="drop-zone" :class="{ 'drag-over': isDragOver }"
             @dragover="onDragOver" @dragleave="onDragLeave" @drop="onDrop"
             @click="filePath = ''">
          <div class="text-2xl mb-1" style="color: var(--color-text-muted)">📂</div>
          <p class="text-sm" style="color: var(--color-text-muted)">拖拽文件到此处</p>
          <p v-if="filePath" class="text-xs mt-1.5 break-all font-mono" style="color: var(--color-accent)">{{ filePath }}</p>
        </div>
      </div>

      <!-- Card 2: Split Parameters (only for split) -->
      <div v-if="isSplit" class="card">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-xs font-medium mb-1.5" style="color: var(--color-text-muted)">切割大小</label>
            <select v-model="chunkSize" class="cursor-pointer">
              <option value="2G">2G</option>
              <option value="1G">1G</option>
              <option value="500M">500M</option>
              <option value="100M">100M</option>
              <option value="50M">50M</option>
              <option value="10M">10M</option>
              <option value="1M">1M</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-medium mb-1.5" style="color: var(--color-text-muted)">校验算法</label>
            <select v-model="hashAlgo" class="cursor-pointer">
              <option value="md5">MD5 (快速)</option>
              <option value="sha256">SHA-256 (安全)</option>
            </select>
          </div>
        </div>
        <div class="mt-4">
          <label class="block text-xs font-medium mb-1.5" style="color: var(--color-text-muted)">输出目录</label>
          <input v-model="outputDir" placeholder="留空 = 自动创建 cutx-<文件名>/ 子目录" class="font-mono">
        </div>
      </div>

      <!-- Card 3: Merge Parameters (only for merge) -->
      <div v-if="isMerge" class="card">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-xs font-medium mb-1.5" style="color: var(--color-text-muted)">合并模式</label>
            <select v-model="mergeMode" class="cursor-pointer">
              <option value="quick">quick (快速，不校验)</option>
              <option value="verify">verify (边合并边校验)</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-medium mb-1.5" style="color: var(--color-text-muted)">输出目录</label>
            <input v-model="outputDir" placeholder="留空 = 清单所在目录" class="font-mono">
          </div>
        </div>
        <label class="flex items-center gap-2 mt-4 cursor-pointer">
          <input type="checkbox" v-model="forceOverwrite" class="w-4 h-4 rounded" style="accent-color: var(--color-accent)">
          <span class="text-sm" style="color: var(--color-text-muted)">强制覆盖已存在的文件</span>
        </label>
      </div>

      <!-- Start Button -->
      <button class="btn-primary" @click="startOperation" :disabled="isRunning"
              style="padding: 14px; font-size: 15px;">
        {{ isRunning ? '运行中...' : '▶ ' + (isSplit ? '开始切割' : isMerge ? '开始合并' : '开始校验') }}
      </button>

      <!-- Card 4: Progress + Log -->
      <div class="card">
        <!-- Progress -->
        <div class="flex items-center gap-3 mb-2">
          <div class="flex-1">
            <div class="progress-track">
              <div class="progress-fill" :style="{ width: progress.percent + '%' }"></div>
            </div>
          </div>
          <span class="text-sm font-mono w-12 text-right" style="color: var(--color-accent)">
            {{ progress.percent.toFixed(0) }}%
          </span>
        </div>
        <div class="flex justify-between text-xs mb-3" style="color: var(--color-text-muted)">
          <span>{{ progress.label }}</span>
          <span v-if="progress.total > 0">{{ formatBytes(progress.done) }} / {{ formatBytes(progress.total) }}</span>
        </div>

        <!-- Log -->
        <div class="rounded-lg p-3 h-44 overflow-y-auto font-mono text-xs leading-relaxed"
             style="background: var(--bg-base); border: 1px solid var(--color-border)">
          <div v-for="(log, i) in logs" :key="i" class="whitespace-pre-wrap">
            <span style="color: var(--color-text-muted)">[{{ log.time }}] </span>
            <span :style="{
              color: log.level === 'success' ? 'var(--color-success)' :
                     log.level === 'error' ? 'var(--color-error)' :
                     log.level === 'warning' ? 'var(--color-warning)' :
                     'var(--color-text)'
            }">{{ log.message }}</span>
          </div>
          <div v-if="logs.length === 0" style="color: var(--color-text-muted)">等待操作...</div>
        </div>
      </div>
    </div>
  </div>
</template>
