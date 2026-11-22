// Screenshots of every page, for UI phases (pal-harness templates/screens.spec.ts).
//
// Copy this file once into the project's e2e folder and list the pages below. It runs on the
// project's own Playwright set-up: its webServer or global setup starts the app, and its baseURL
// points at it. The review stage runs it on the final commit when screens changed:
//
//   SCREENS_DIR=<set>/specs/NNN-slug/screens npx playwright test <e2e folder>/screens.spec.ts
//
// Each page is saved in light and dark, at the default width and the narrowest supported one, as
// <page>-<theme>-<width>.png. The PNGs are committed on docs/NNN-slug, read by the UI reviewers and
// shown in the sign-off PR's Screens section.
import { test, type Page } from '@playwright/test'
import { mkdirSync } from 'node:fs'
import path from 'node:path'

// Every page (route) the app has. Add one line per page; a page reached by a click only can use
// `open` instead of `path`.
const pages: { name: string; path?: string; open?: (page: Page) => Promise<void> }[] = [
  { name: 'home', path: '/' },
]

// The default width, and the narrowest the project supports (its minimum window or phone width).
const widths = [
  { name: 'wide', width: 1280, height: 800 },
  { name: 'narrow', width: 360, height: 740 },
]

const themes = ['light', 'dark'] as const

// A full-page capture only grows the page itself. Content inside an inner scroll container (a
// main panel with overflow: auto) would be cut off, so each inner scroll container is let grow to
// its content first.
async function growInnerScroll(page: Page) {
  await page.evaluate(() => {
    for (const el of Array.from(document.querySelectorAll<HTMLElement>('body *'))) {
      const style = getComputedStyle(el)
      const scrolls = /(auto|scroll)/.test(style.overflowY) && el.scrollHeight > el.clientHeight
      if (scrolls) {
        el.style.overflowY = 'visible'
        el.style.height = 'auto'
        el.style.maxHeight = 'none'
      }
    }
  })
}

// The theme follows prefers-color-scheme. If the app's theme is an in-app setting instead, change
// this hook to set it (a stored setting, a class on <html>, a toggle click) after the page loads.
async function setTheme(page: Page, theme: (typeof themes)[number]) {
  await page.emulateMedia({ colorScheme: theme })
}

const out = process.env.SCREENS_DIR ?? path.join('test-results', 'screens')

// The warm-up: a dev server (Vite, Next.js) compiles on its first load, and that cold start can
// outlast a test's timeout. One load of the first page, with a longer timeout and no screenshot,
// runs before the first capture.
test.beforeAll(async ({ browser }, testInfo) => {
  mkdirSync(out, { recursive: true })
  test.setTimeout(120_000)
  const page = await browser.newPage({ baseURL: testInfo.project.use.baseURL })
  const first = pages.find((p) => p.path !== undefined)
  await page.goto(first?.path ?? '/', { timeout: 110_000 })
  await page.waitForLoadState('networkidle', { timeout: 110_000 })
  await page.close()
})

for (const p of pages) {
  for (const theme of themes) {
    for (const w of widths) {
      test(`${p.name} ${theme} ${w.name}`, async ({ page }) => {
        await page.setViewportSize({ width: w.width, height: w.height })
        await setTheme(page, theme)
        if (p.open) {
          await p.open(page)
        } else {
          await page.goto(p.path ?? '/')
        }
        await page.waitForLoadState('networkidle')
        await growInnerScroll(page)
        await page.screenshot({ path: path.join(out, `${p.name}-${theme}-${w.name}.png`), fullPage: true })
      })
    }
  }
}
