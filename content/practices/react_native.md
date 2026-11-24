# React Native Practices (Expo)

## Tech Stack

| Concern | Library |
|---|---|
| Framework | React Native + Expo (managed workflow) |
| Navigation | Expo Router (file-based) |
| Styling | NativeWind (Tailwind on RN) |
| Server state | TanStack Query v5 |
| Client state | Zustand |
| Forms | React Hook Form + Zod |
| Token storage | `expo-secure-store` |
| Local storage | `react-native-mmkv` (non-sensitive fast KV) |
| Images | `expo-image` |
| Push notifications | `expo-notifications` (abstracts FCM + APNs) |
| Biometric auth | `expo-local-authentication` |
| OTA updates | `expo-updates` |
| Build & submit | EAS Build + EAS Submit |
| Unit testing | Jest + React Native Testing Library |
| E2E testing | Maestro |

## Project Structure

```
app/
  (auth)/                   # unauthenticated screens
    login.tsx
    register.tsx
    _layout.tsx
  (tabs)/                   # main tab navigation
    index.tsx               # home tab
    explore/
      index.tsx
      [id].tsx              # dynamic route — same pattern as Next.js
    profile.tsx
    _layout.tsx             # tab bar config
  _layout.tsx               # root layout — auth guard, providers
  +not-found.tsx            # 404 equivalent
components/
  ui/                       # base components (NativeWind styled)
  [feature]/                # feature-specific components
lib/
  api/                      # typed fetch wrappers — same pattern as web
  auth.ts                   # token management (SecureStore read/write)
  validations/              # Zod schemas — shared with web where possible
  utils.ts                  # shared utilities
hooks/                      # custom hooks (useX pattern)
store/                      # Zustand stores
types/
  api.ts                    # API response types — aligned with Go API contract
constants/                  # colours, spacing, config values
```

## Navigation (Expo Router)

Same file-based mental model as Next.js App Router:

| Next.js | Expo Router | Purpose |
|---|---|---|
| `app/(auth)/login/page.tsx` | `app/(auth)/login.tsx` | Route group |
| `app/[id]/page.tsx` | `app/[id].tsx` | Dynamic route |
| `layout.tsx` | `_layout.tsx` | Layout / navigator config |
| `loading.tsx` | Suspense boundary | Loading state |
| `not-found.tsx` | `+not-found.tsx` | 404 |

- Use `<Link>` from `expo-router` for navigation (not `navigation.navigate`)
- Typed routes enabled in `app.config.ts` — `href` props are type-safe
- Deep links and universal links handled automatically by Expo Router
- Auth guard in root `_layout.tsx` — redirect to `/(auth)/login` if no session

## Styling (NativeWind)

Same Tailwind class names as the web project — shared design language:

```tsx
<View className="flex-1 bg-white px-4 py-6">
  <Text className="text-xl font-semibold text-gray-900">Hello</Text>
</View>
```

- Configure `tailwind.config.js` aligned with web project's config (shared tokens)
- Platform variants: `ios:pt-12 android:pt-8` for platform-specific overrides
- `styled()` wrapper for third-party components that don't accept `className`
- Never mix NativeWind `className` with `StyleSheet.create` on the same component

## Authentication

- JWT tokens stored in `expo-secure-store` — hardware-backed, encrypted at rest
- **Never AsyncStorage for tokens or any sensitive data**
- Access token + expiry stored; refresh token stored separately
- On app foreground (`AppState` change): check expiry, refresh if within 60s
- On 401 from API: refresh once, retry request, then redirect to login
- Biometric unlock via `expo-local-authentication` as optional layer over token retrieval

```typescript
// lib/auth.ts — token helpers
export const saveTokens = async (access: string, refresh: string) => {
  await SecureStore.setItemAsync('access_token', access)
  await SecureStore.setItemAsync('refresh_token', refresh)
}
export const getAccessToken = () => SecureStore.getItemAsync('access_token')
```

## State Management

Same pattern as web — TanStack Query for server state, Zustand for client state:

- TanStack Query persisted to MMKV via `createSyncStoragePersister` — enables offline data
- Zustand stores for UI state: `useUIStore` (loading overlays, modals), `useAuthStore` (user session)
- Background refetch on app foreground via `focusManager.setEventListener`

```typescript
// Refetch queries when app comes to foreground
focusManager.setEventListener((handleFocus) => {
  const sub = AppState.addEventListener('change', (state) => {
    handleFocus(state === 'active')
  })
  return () => sub.remove()
})
```

## Forms

Same React Hook Form + Zod pattern as web. Zod schemas shared across web and mobile where possible:

```typescript
// validations/trip.ts — usable in both web and mobile
export const createTripSchema = z.object({
  title: z.string().min(1).max(100),
  startDate: z.string().datetime(),
})
```

- Use `Controller` from React Hook Form for RN inputs (not `register` — RN has no DOM refs)
- Inline validation errors displayed below each field
- Disable submit button while `formState.isSubmitting`

## API Client

Same typed wrapper pattern as web (`lib/api/`). Platform difference: inject auth header from SecureStore:

```typescript
const apiFetch = async (path: string, options?: RequestInit) => {
  const token = await getAccessToken()
  return fetch(`${process.env.EXPO_PUBLIC_API_URL}${path}`, {
    ...options,
    headers: { Authorization: `Bearer ${token}`, ...options?.headers },
  })
}
```

- `EXPO_PUBLIC_*` prefix for env vars accessible in JS bundle (not secret)
- Secrets (API keys, internal config) go in EAS environment variables, accessed server-side only

## Performance Defaults

- `FlashList` (Shopify) for all scrollable lists — never `FlatList` for lists > 50 items
- `expo-image` for all images — never RN `Image`. Handles caching, progressive loading
- `useMemo` / `useCallback` / `React.memo` only after profiling — not pre-emptively
- Hermes engine is on by default in Expo SDK 50+ — do not disable
- Avoid heavy computation on the JS thread — use `expo-task-manager` for background work

## Push Notifications

- Register Expo Push Token on login, store in user profile on API
- Deregister (or mark inactive) token on logout
- Handle three notification states:

| State | Handler |
|---|---|
| Foreground | `Notifications.addNotificationReceivedListener` |
| Background tap | `Notifications.addNotificationResponseReceivedListener` |
| Killed state | Check `Notifications.getLastNotificationResponseAsync` on launch |

- Never assume notifications are enabled — always check permissions before registering

## Build & Deploy (EAS)

```bash
eas build --platform all --profile preview   # internal testing build
eas build --platform all --profile production # store submission build
eas submit --platform ios                     # submit to App Store
eas submit --platform android                 # submit to Play Store
eas update --branch production               # OTA JS-only update (no store review)
```

- `app.config.ts` (not `app.json`) for dynamic config — reads from `process.env`
- Three EAS build profiles: `development` (dev client), `preview` (internal), `production` (store)
- OTA updates for JS changes. Native changes (new native modules, permissions) always require a full build
- Bundle IDs: `com.<company>.<product>` for iOS, same for Android package name
- **App version:** every phase that changes the app bumps the app version (`version`, and the
  build number EAS does not set itself) in a plan task, the runbook's release step, whatever the
  flow overrides. A shipped build without a bump cannot be told apart from the last one.
- **Native dependencies:** a task that adds or upgrades a native dependency records the resolved
  versions of the shared libraries it pulls in (for Android, `./gradlew :app:dependencies` filtered
  by the library, such as `androidx.credentials`), against the existing native modules. A
  mismatch is a major plan finding: every test can pass while the device breaks.

## Native builds on Windows

- A local Android build builds arm64 only and runs at low process priority, through Gradle's own
  flag (it covers the Gradle daemon and its children, from any shell):
  - `cd android && ./gradlew assembleDebug -PreactNativeArchitectures=arm64-v8a --priority=low`;
  - or `npx expo run:android` with `ORG_GRADLE_PROJECT_reactNativeArchitectures=arm64-v8a` set and
    `org.gradle.priority=low` in the user-level `%USERPROFILE%\.gradle\gradle.properties`
    (`GRADLE_USER_HOME`, which `expo prebuild --clean` never regenerates); never pass `--all-arch`.
- All-ABI builds are for EAS only, never local: the first local all-ABI build took 31 minutes and
  flooded the owner's PC.

## TypeScript

- Strict mode on
- Expo Router provides typed `href` — use `Href` type from `expo-router`
- `EXPO_PUBLIC_*` env vars typed via `env.d.ts`
- No `any` — use `unknown` and narrow, or Zod for external data

## Testing

- **Unit**: Jest + React Native Testing Library. Same fixture pattern: `tests/unit_test/{use_case}.json`
- **API mocking**: MSW with `msw/native` adapter — same mocks as web where API surface is shared
- **E2E**: Maestro — YAML flow files in `tests/e2e/`. Test critical journeys on simulator + real device
- Never test navigation internals — test that the correct screen content appears
- `jest.config.js` uses `jest-expo` preset
- Jest shims for native modules export only what the native module exports, with a parity test
  (`testing.md`, Next.js / React-Specific)
- Lint and format through the repo's scripts (`npm run lint`, `npm run format`), never a bare
  `npx prettier` (`workflow.md`, **Check before every push**)

```yaml
# tests/e2e/create_trip.yaml — Maestro flow
appId: com.company.etp
---
- launchApp
- tapOn: "Create Trip"
- inputText:
    id: "trip-title-input"
    text: "Tokyo 2025"
- tapOn: "Save"
- assertVisible: "Tokyo 2025"
```

## Device testing

- Start Metro for device tests with file watching: `npx expo start`, never `CI=1` or
  `--non-interactive`, so a file added in a fix round bundles without a restart. This holds in
  every fix round and every restart, not only the first start.
- Never run `adb disconnect` or `adb kill-server`, and never toggle the phone's wireless debugging.
- **A fresh app for each test.** The staging-test pre-flight clears the app's data on each phone with
  `adb -s <serial> shell pm clear <package>` (never the bare `adb shell pm clear <package>`, which
  names no phone). Auto mode blocks it unless the SSOT repo's committed `.claude/settings.json`
  allows exactly that command; the committed file, because a worktree set never sees
  `settings.local.json`. For
  each package (production and development variants) and each serial in `doc/devices.yaml`,
  `permissions.allow` holds:

  ```json
  "Bash(adb -s R5CX123 shell pm clear com.example.app)",
  "Bash(adb -s R5CX123 shell am force-stop com.example.app)"
  ```

  The first device test of a session (a device-check task, or the staging-test pre-flight) checks
  the file first. When an entry is missing, the session asks once, in that first message, with the
  exact lines; on the owner's yes it adds them and commits them with the phase's docs. Then
  `adb -s <serial> shell pm clear <package>` and `adb -s <serial> shell am force-stop <package>` run
  as listed, never asked again mid-test. A project that cannot allow them ships a debug-only reset
  in the app (a hidden developer action that clears its storage) and the pre-flight uses that.
- **Screen-capture protection.** An app that blocks screen capture (FLAG_SECURE, or
  `usePreventScreenCapture`) shows black screenshots.
  1. Read its screens with `adb -s <serial> shell uiautomator dump /sdcard/ui.xml`, then
     `adb -s <serial> pull /sdcard/ui.xml`. The screen's text, with the server or phone log, is the
     step's evidence, and the QA or device-check summary row says `screen text: uiautomator`.
  2. The development variant (`APP_VARIANT=development`) leaves the protection off by a build-time
     flag, so the device check can screenshot it; a test proves the production variant keeps it on.
  3. The unlock-free route: the development variant also accepts a test PIN, typed by the session
     with `adb -s <serial> shell input text <pin>` in place of a fingerprint. A build-time flag the
     production build cannot set guards it, and a test proves the production build refuses it. The
     unlock is never skipped, and the test PIN is never a real secret. A project adds the flag and
     the PIN in a plan task the first time a device check needs them.
- The test devices are listed in `doc/devices.yaml`: names and addresses only, no pairing codes.

  ```yaml
  ports: [8081, 8080]               # adb reverse on each device: Metro, and the local API
  devices:
    - name: Samsung A26             # what the owner calls it
      role: primary                 # primary, or second-user: the phone of a second test account
      serial: R5CX123               # adb's USB serial (also inside the wireless mDNS name)
      wireless: 192.168.1.23:41234  # last known address for adb connect
    - name: Pixel 6
      role: second-user
      wireless: 192.168.1.40:39000
  ```

- Bring them up with `pal-harness adb` (run as `.pal-harness/practices/ssot_standard.md` says) before a device test, after an
  owner gap and when the link drops. For each device it keeps one already attached, else connects at
  the saved address, else at the address `adb mdns services` finds now; then it sets `adb reverse` for
  each port. It prints one line per device. Save a new address it prints in `doc/devices.yaml` with
  the phase's docs.
  - Exit 1 names each device that is not ready and the request to send the owner ("on <name>, open
    Settings → Developer options → Wireless debugging → Pair device with pairing code, and send me
    the code and the port"). Pair with `adb pair <ip>:<port> <code>`, then run it again.
  - Exit 2 (no `doc/devices.yaml` yet, or no `adb`): reconnect once (`adb connect <ip>:<port>`) from
  the device details saved in the project's docs (if none are saved, ask the owner), and write
  `doc/devices.yaml` from them. If that fails, stop and ask the owner, giving the device name and what you tried.
- A test script that needs two accounts names the device each account uses.

## Key Differences from Web

| Concern | Next.js Web | React Native |
|---|---|---|
| Routing | `app/page.tsx` | `app/screen.tsx` (Expo Router) |
| Styling | Tailwind `className` | NativeWind `className` |
| Token storage | `httpOnly` cookie | `expo-secure-store` |
| Data persistence | Browser cache | MMKV + TanStack Query persist |
| Images | `next/image` | `expo-image` |
| Forms | `register()` | `Controller` |
| E2E | Playwright | Maestro |
| Deploy | Vercel | EAS Build + EAS Submit |
