# Next.js Web Practices (App Router)

## Tech Stack

| Concern | Library |
|---|---|
| Framework | Next.js (App Router, TypeScript strict) |
| Styling | Tailwind CSS |
| Components | shadcn/ui (Radix UI primitives) |
| Server state | TanStack Query (React Query v5) |
| Client state | Zustand |
| Forms | React Hook Form + Zod |
| Auth | NextAuth.js (Auth.js v5) — Credentials provider → Go API |
| HTTP client | `fetch` (native) wrapped in typed API client |
| E2E testing | Playwright |
| Unit testing | Jest + React Testing Library + MSW |
| Linting | ESLint (`eslint-config-next`) + Prettier + Tailwind plugin |

## Project Structure

```
app/
  (auth)/                   # route group — unauthenticated pages
    login/
    register/
  (protected)/              # route group — requires auth (middleware guards)
    dashboard/
    settings/
  api/
    auth/[...nextauth]/     # NextAuth.js route handler
components/
  ui/                       # shadcn/ui generated components — minimal customisation
  [feature]/                # feature-specific components (colocated with pages if small)
lib/
  api/                      # typed fetch wrappers to Go API — one file per domain
  auth.ts                   # NextAuth config (providers, callbacks)
  validations/              # Zod schemas — source of truth for form + API types
  utils.ts                  # cn() and other utilities
hooks/                      # custom React hooks (useX pattern)
store/                      # Zustand stores — one file per domain
types/
  api.ts                    # API response types (aligned with Go API contract)
public/                     # static assets
```

## Server vs Client Components

- **Default to Server Components**. Add `"use client"` only when required:
  - Event handlers (`onClick`, `onChange`, etc.)
  - React hooks (`useState`, `useEffect`, `useContext`)
  - Browser-only APIs (`window`, `localStorage`)
  - TanStack Query hooks
- Data fetching belongs in Server Components — not `useEffect`
- Pass server-fetched data as props to Client Components when needed
- Never import a Server Component into a Client Component

## Data Fetching

**Server Components** (preferred for initial page data):
```typescript
// Direct fetch in async Server Component
const data = await fetch(`${process.env.API_URL}/api/v1/users`, {
  headers: { Authorization: `Bearer ${session.accessToken}` },
  next: { revalidate: 60 }, // or cache: 'no-store' for dynamic
})
```

**Client Components** (user-triggered or real-time):
```typescript
// TanStack Query hook
const { data, isLoading } = useQuery({
  queryKey: ['users', filters],
  queryFn: () => apiClient.users.list(filters),
})
```

Never fetch the same data in both a Server Component and a Client Component.

## Authentication (NextAuth.js v5 + Go API)

- Credentials provider calls Go API `POST /api/v1/auth/login`
- Store `accessToken` + `refreshToken` + expiry in encrypted JWT session
- `jwt` callback: refresh access token when within 60s of expiry
- `session` callback: expose `accessToken` to client (never `refreshToken`)
- `middleware.ts` uses `auth()` to protect all `/protected` routes
- Tokens stored in `httpOnly` session cookie — never `localStorage`

```typescript
// middleware.ts
export { auth as middleware } from '@/lib/auth'
export const config = { matcher: ['/(protected)(.*)'] }
```

## State Management

**TanStack Query** — all server/async state:
- Query keys are arrays: `['users']`, `['users', id]`, `['users', filters]`
- Mutations invalidate related queries via `queryClient.invalidateQueries`
- Optimistic updates for UX-critical mutations (likes, toggles)
- `staleTime: 60_000` as default; override per query as needed

**Zustand** — client-only UI state:
- One store per domain: `useAuthStore`, `useUIStore`
- Keep stores flat — no nested objects
- Never put server data in Zustand (that's TanStack Query's job)

## Forms

React Hook Form + Zod is the standard:

```typescript
const schema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
})
type FormData = z.infer<typeof schema>

const { register, handleSubmit, formState: { errors } } = useForm<FormData>({
  resolver: zodResolver(schema),
})
```

- Zod schema is the single source of truth for validation and TypeScript type
- Client-side validation runs first; API validation errors map to field errors via `setError`
- Never submit a form without client-side validation

## TypeScript

- Strict mode on (`"strict": true` in `tsconfig.json`)
- No `any` — use `unknown` and narrow with type guards or Zod
- API response types defined in `types/api.ts`, aligned with Go API contract
- Zod schemas generate types via `z.infer<>` — do not duplicate type definitions
- All async functions have explicit return type annotations

## API Client

Typed wrappers in `lib/api/` — one file per domain:

```typescript
// lib/api/users.ts
export const usersApi = {
  list: (filters: UserFilters): Promise<ApiResponse<User[]>> =>
    apiFetch('/api/v1/users', { params: filters }),
  getById: (id: string): Promise<ApiResponse<User>> =>
    apiFetch(`/api/v1/users/${id}`),
}
```

- `apiFetch` base wrapper handles: auth header injection, RFC 7807 error parsing, token refresh
- Never call `fetch` directly in components or TanStack Query functions
- All 401 responses trigger token refresh then retry once; second 401 → redirect to login

## Performance Defaults

- `next/image` for all images — never raw `<img>` tags
- `next/font` for all fonts — zero layout shift
- `next/dynamic` with `{ ssr: false }` for heavy client-only components (charts, editors)
- `Suspense` boundaries at page level and around any slow async Server Component
- `loading.tsx` files for streaming UI while data loads

## Error Handling

- `error.tsx` per route segment for recoverable errors
- `global-error.tsx` in root for catastrophic failures
- RFC 7807 errors from Go API mapped to user-friendly messages in `lib/api/errors.ts`
- Never show raw API error messages to users

## Routing Conventions

- Route groups `(name)` for layout separation — auth vs protected vs public
- Dynamic segments `[id]` with typed `params` via `generateStaticParams` where possible
- Parallel routes `@slot` for complex dashboard layouts
- Intercepting routes `(.)path` for modals over pages (e.g. photo lightbox)

## shadcn/ui Conventions

- Run `npx shadcn@latest add <component>` — never copy-paste component code manually
- Customise via `className` prop and Tailwind — do not edit generated component internals
- Extend with wrapper components in `components/[feature]/` — never modify `components/ui/`
- Use `cn()` from `lib/utils.ts` for conditional class merging

## Testing

- **Unit**: Jest + React Testing Library. Test behaviour, not implementation.
- **API mocking**: MSW (Mock Service Worker) — no manual `jest.mock('fetch')`.
- **E2E**: Playwright. Test critical user journeys end-to-end against a real API.
- Fixtures: `tests/unit_test/{use_case}.json`, `tests/e2e/{use_case}.json`
- Never test shadcn/ui internals — test your components that use them
