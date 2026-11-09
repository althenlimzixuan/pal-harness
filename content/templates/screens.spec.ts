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

// The theme follows prefers-color-scheme. If the app's theme is an in-app setting instead, change
// this hook to set it (a stored setting, a class on <html>, a toggle click) after the page loads.
async function setTheme(page: Page, theme: (typeof themes)[number]) {
  await page.emulateMedia({ colorScheme: theme })
}

const out = process.env.SCREENS_DIR ?? path.join('test-results', 'screens')

test.beforeAll(() => {
  mkdirSync(out, { recursive: true })
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
        await page.screenshot({ path: path.join(out, `${p.name}-${theme}-${w.name}.png`), fullPage: true })
      })
    }
  }
}
