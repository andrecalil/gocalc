# Calculator Web

React 19 + TypeScript + Vite 6 + Tailwind 4 + HeroUI v3 + TanStack Query + Zustand.

## Run

```bash
npm install
npm run dev          # http://localhost:5173 (proxies /api/* → localhost:8080)
npm run build        # type-check + production build
npm run preview      # preview the production build
```

## Test

```bash
npm test                    # run once
npm run test:watch          # watch mode
npm run test:coverage       # with coverage
```

## Layout

- `src/api/` — fetch wrapper (`calculate`) and typed error class
- `src/hooks/` — TanStack Query mutation wired to stores + toast
- `src/stores/` — Zustand stores: `calcStore` (session), `historyStore` (persisted), `themeStore` (persisted)
- `src/components/` — `Calculator` (container + keyboard handling), `Display`, `Keypad`, `History`, `ThemeToggle`
- `src/styles/globals.css` — Tailwind + HeroUI imports; gold accent override
- `src/test/` — Vitest setup + MSW handlers + render helpers
