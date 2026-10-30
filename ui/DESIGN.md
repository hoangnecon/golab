# GoLab UI — Design & Engineering Contract

Binding for every human and agent that touches `ui/`. If code and this file disagree, fix the code. If this file is wrong, change this file in the same commit. Do not work around it.

Scope: `ui/` only. The Go backend (`internal/`, `cmd/`) and its HTTP/SSE contract are out of scope and must not be changed from UI work.

Verify before reporting done (all must pass, run from `ui/`):

```
./node_modules/.bin/tsc --noEmit      # pnpm typecheck
node scripts/check-design.mjs         # pnpm check:design
./node_modules/.bin/vite build        # pnpm build   (regenerates dist/, embedded by assets.go)
```

Design intelligence source: the `ui-ux-pro-max` skill (installed in `.agents/skills/ui-ux-pro-max`). Its generated direction for this product (developer tool, dense dashboard, dark-first) was *Minimalism & Swiss* with IBM Plex Sans + JetBrains Mono. The product owner narrowed the palette to **black, white and gray only**. This file is the project-specific refinement; where they differ, this file wins.

---

## 1. Product principles

1. **Tool, not website.** One job: run Colab notebooks under browser accounts and read the tool-call log. Clarity and density over decoration.
2. **Depth by surface, not by effect.** Hierarchy comes from stepped surfaces (`canvas` → `surface` → `raised`), hairline borders and type weight. No gradients, glass, blur, glow, illustrations.
3. **Monochrome.** Black, white and grays only; every colour token is a pure gray (R = G = B). No hue anywhere: not for the primary action, not for state, not for errors. Emphasis comes from contrast (`fg` fill vs. `surface`), weight and shape.
4. **Nothing that does not carry information or an action.** No taglines, helper paragraphs, version strings, decorative icons. Icons exist to speed recognition of an action or an empty state.
5. **One way to do each thing.** Every recurring need has exactly one component (section 7).
6. **State is shape + label.** `StatusDot` encodes state by shape (section 7) and always has a spoken label; where space allows the label is visible (`Status`).

## 2. Information architecture

The model mirrors how the tool is used:

- **Project** = one Colab notebook. Holds a name and a notebook URL.
- **Run** = that notebook opened under one **Account** (the backend calls it a *session*). A run is created by choosing an account.
- **Account** = a Chrome profile signed in to Google (the backend calls it a *profile*).
- **Logs** = MCP tool calls made by the AI agent (the backend calls it *activity*). Viewable per run, per project, per account, or globally.

Translate backend names only at the UI boundary.

**Projects exist only in the UI.** The backend has no project entity. Saved projects live in `localStorage` (`golab.projects`); a session whose notebook matches no saved project appears as an implicit project named after the run. The project id is a hash of the normalised notebook URL (origin + path), so runs and projects join without a table. Consequences: projects are per browser, are not shared or backed up, and a saved project with zero runs is lost if storage is cleared. Fixing that needs a backend `projects` endpoint (out of scope for UI work; report it).

### Routes

Hash routes, built only with `shared/lib/paths.ts`, parsed only by `shared/lib/location.ts`. Pages receive the path segments after the section as `segments`.

| URL | Screen |
|---|---|
| `#/projects` (default) | project cards, New project |
| `#/projects/<pid>` | project: runs, tool calls; rail with notebook and actions |
| `#/projects/<pid>/runs/<sid>` | run: tool calls; rail with status, details and controls |
| `#/accounts` | accounts table, New account |
| `#/accounts/<id>` | account: runs, tool calls of its runs; rail with browser state and actions |
| `#/logs?q=&tool=&run=&account=&errors=1` | global log; filters live in the URL |

Never write a `#/…` string by hand. Updating query params from a controlled input uses `replaceParams` (synchronous; `location.replace` reverts typed text).

## 3. Architecture

```
src/
  main.tsx                  entry
  index.css                 font imports, colour tokens (CSS variables), focus ring, reduced motion
  app/
    App.tsx                 providers + Shell + route outlet + offline banner
    routes.ts               ROUTES registry (projects, accounts, logs); NAV_ITEMS (pinned: accounts, logs); useRoute()
  api/                      the only code that knows URLs
    http.ts  endpoints.ts  types.ts
  data/
    ControlPlaneProvider    server state: one reducer, initial load, SSE ticks, all mutations, activity merged by id
    ProjectsProvider        saved projects (localStorage) + createProject / removeProject
    projects.ts             storage, URL normalisation, project id
    useWorkspace.ts         read-only derived model: projects with runs, lookups, callsOf(runs)
    useActivityDetail.ts    one call's request/response payloads, fetched on expand, cached by id
    useEventStream.ts
  shared/                   feature-agnostic; MUST NOT import from domain/, features/ or app/
    layout/                 ShellProvider, Shell, Sidebar (ROW), Page, Detail (Detail, Section, Panel, PanelBody, PanelActions, Metrics)
    ui/                     Button/ButtonLink, IconButton/IconLink, Field (Input, SearchInput, Select, Checkbox, FormError),
                            Dialog, ConfirmDialog (useConfirm), Table (TableCard, TBody, Th, Td), StatusDot (Status),
                            Meter, PropertyList, Avatar, EmptyState, CodeBlock
    notice/                 NoticeProvider (toasts)
    lib/                    cn, format, status, paths, location
  domain/                   blocks reused by several features; may use data/ and shared/
    runs/                   RunsTable, useRunActions
    logs/                   CallsTable, CallDetail
  features/<name>/          projects, accounts, logs
    <Name>Route.tsx         maps segments to a page
    <Name>Page.tsx          container: reads data, owns dialog/confirm state, renders <Page>
    <Verb>Dialog.tsx        form dialog
    ProjectsNav.tsx         the sidebar project tree (injected into Shell)
```

### Layering

`app` → `features` → `domain` → `shared` / `data` → `api`. A feature never imports another feature; anything two features need goes to `domain/` (or `shared/` if it knows no business concept). `shared` imports nothing above itself, so `App` injects the sidebar tree as a prop.

### Patterns in use

| Pattern | Where | Rule |
|---|---|---|
| Composition root | `app/App.tsx` | Providers nest by dependency: Notice → ControlPlane → Projects → Shell. |
| Registry | `app/routes.ts` | Adding a section = one entry in `ROUTES`. |
| Container / presentational | `features/*`, `domain/*` | `*Page` touches data; tables and dialogs receive props. |
| Derived read model | `data/useWorkspace` | Joins server state with saved projects. Pages never join by hand. |
| Single store behind a typed facade | `data/ControlPlaneProvider` | Components call `useControlPlane()`; never `fetch` / `api.*`. |
| Variant maps | `Button`, `StatusDot` | Variants are a `Record<Variant, string>` inside the component, never ad-hoc class strings at call sites. |
| Slot injection | `Shell.tree` | Shared layout exposes slots instead of importing data. |
| Hook-as-service | `useConfirm`, `useNotice`, `useShell`, `useRunActions` | Cross-cutting capabilities are hooks. |
| Native platform first | `<dialog>`, hash routing, `inert` | Prefer the browser feature over a dependency. |

### State rules

- Server state lives only in `ControlPlaneProvider`; saved projects only in `ProjectsProvider`. Local `useState` is for transient UI.
- Navigation state lives in the URL hash.
- Sidebar open/closed lives in `ShellProvider` (`localStorage` key `golab.sidebar`). Compact screens start closed.
- Live data arrives by SSE (`/api/events`, `type: "tick"`); the UI never polls.
- Go nil slices serialize as `null`; normalise in `api/`/`data/`, never in components.
- `kernel_state` is a free string from the backend. `shared/lib/status.ts` maps the Jupyter names it knows and renders anything else as neutral. Do not add guesses.

### Adding things

- **New screen:** `features/<name>/` with `<Name>Route`, pages and dialogs → add to `ROUTES`.
- **New endpoint:** `api/types.ts` → `api/endpoints.ts` → a provider action → UI.
- **New primitive:** only if no existing one fits and it is used in ≥2 places. Put it in `shared/ui/`.

## 4. Layout

- Two columns on the `app` background: sidebar | an **inset content frame** (`main`: `rounded-xl bg-canvas ring-1 ring-line`, `relative`, inside a `p-2` gutter; `md:pl-0` while the sidebar is open). **No top bar over the sidebar.** The frame uses `ring`, not `border` (a border shifts the header by 1px); `relative` keeps transformed children (diamond dots) from widening the page.
- **One toggle, one position.** `Shell` renders a single `IconButton` (32px) fixed at `left-3 top-5` in every state and viewport; only its icon changes.
- **Sidebar:** full viewport height, `w-64` with `p-2`, on the `app` background (no border; the frame edge separates it). Top to bottom: brand row (`h-14`, `pl-10`, text at the grid text column); the **project tree** (scrolls): a "Projects" header with a `sm` New project icon button, then each project (notebook icon, name, chevron in the trailing slot) expanding to its runs (name, `StatusDot` in the trailing slot, a hairline guide at the icon column); pinned destinations **Accounts** and **Logs** under a hairline (icon + label; active = `bg-hover`, `font-medium`). Nothing sits under them: no status text, no version, no footer. Projects start expanded.
- **Closed means gone:** width 0, `inert`, removed from layout. Shortcut `Ctrl/⌘ + B`.
- **Compact (<768px):** the sidebar is an overlay drawer with scrim, starts closed, closes after navigation.
- **Every route renders through `<Page trail meta? action?>`:** `h-14` header with a hairline: breadcrumb trail separated by `/` (last crumb is the `h1`), quiet count in `meta`, and **at most one** `action` (a `sm` button: the create action of the page, or Pause/Resume on Logs). Entity controls never go in the header; they live in the rail. No tabs. Then the scrolling body `p-6` (`p-4` below md). Header left padding is `pl-6` with the sidebar open, `pl-14` when closed (clears the toggle).
- **Content is fluid and left-aligned.** No `max-w-*` page container, no `mx-auto`.
- **Detail pages** (project, run, account) use, top to bottom: `Metrics` (a `gap-px` cell grid of label + value, 2 columns below lg, N at lg; an odd last cell spans 2), then `Detail`: a `main` column (`xl:col-span-2`) of `Section`s (`h-8` title + quiet count: Runs, Tool calls) and a rail (`xl:col-span-1`, first on compact screens) of `Panel`s: facts as `PropertyList`, then a "Controls" `PanelActions` of `menu` buttons with the destructive group last behind a hairline (`separated`). Logs of the entity are a section on the same page; there are no sub-tabs.
- **Index pages:** projects = card grid (`grid-cols-1 md:2 xl:3`); accounts, logs = `TableCard`.
- **Never set a font size on `html`**: all `rem` spacing derives from it.

### Layout grid

Measured in the smoke test; keep these alignments when touching the frame.

| Line | Value | Who sits on it |
|---|---|---|
| y = 36 | header centre | toggle centre, sidebar brand row, page header (frame top 8 + 28) |
| x = 28 | icon column | toggle centre, every sidebar row icon, tree guide line |
| x = 48 | text column | brand, project / run / nav labels |
| x = 220..244 | trailing slot | chevrons, run status dots, the `+` button |

Sidebar rows use the exported `ROW` class (`h-8`, rounded, colour transition).

## 5. Visual tokens

Tokens are closed: `tailwind.config.js` **replaces** the theme. If a utility is not listed here, it does not exist.

### Colour (defined once in `index.css`; light and dark follow the OS; every value is a pure gray)

| Token | Use |
|---|---|
| `app` | window background behind the sidebar and the content frame |
| `canvas` | content frame, inputs, metric cells |
| `surface` | cards, tables, panels, dialogs |
| `raised` | avatar, dialog footer, expanded log row |
| `hover` | hover / active-nav background |
| `line` | hairlines, dividers, meter track |
| `line-strong` | control borders, idle status dot |
| `fg` / `muted` / `subtle` | primary / secondary / tertiary text (`subtle` for hints, headers, timestamps only). `fg` is also the primary-button fill, meter fill, status shapes, focus ring and toast background |
| `scrim` | modal / drawer backdrop |

Rules: no hue tokens (`accent`, `warning`, `danger`, `success`, … are rejected by the guard), no raw palette colours, no hex/`rgb()` in components, no arbitrary values `[...]`. The only alpha modifier in use is `/85` on primary-button hover.

### Type

| Class | Size | Use |
|---|---|---|
| `text-display` | 24/32 | reserved, currently unused |
| `text-title` | 16/24 | page title, dialog title, brand |
| `text-body` | 14/20 | default |
| `text-caption` | 12/16 | meta lines, labels, table headers, timestamps |

Fonts are bundled (`@fontsource`, no network at runtime): **IBM Plex Sans** (400/500/600) and **JetBrains Mono** (400/500). `font-mono` only for machine identifiers (tool names, IDs, ports, paths, timestamps in logs). Weights: `font-medium` emphasis, `font-semibold` titles. Numbers in columns are `tabular-nums` (also enabled globally).

### Shape, space, depth, motion

- Radius: `rounded` (8px) for controls and sidebar rows, `rounded-lg` (12px) for containers (panels, tables, cards, metric grid, toasts), `rounded-xl` (16px) for the content frame and dialogs, `rounded-full` for dots, avatars, meters and the empty-state icon. Nothing else.
- Shadow: `shadow-overlay` only, on things that float (dialog, toast). Everything else uses borders.
- Heights: buttons `h-9` (`md`) / `h-8` (`sm`), inputs and selects `h-9`, icon buttons `size-8` (`md`) / `size-6` (`sm`), table header `h-10`, table row `h-12`, property row `h-11`, page / dialog header `h-14`. Icons `size-4`, `strokeWidth={1.75}`, `aria-hidden`.
- Motion: `transition-colors` on interactive elements, 200ms on sidebar width and page title padding, `transition-all` on the meter. No entrance animations, pulses, shimmer. `prefers-reduced-motion` is honoured globally.
- Focus: one global `:focus-visible` outline (2px `fg`). Never removed or restyled per component.

## 6. Content rules

- Labels are nouns, buttons are verbs. Sentence case, no trailing periods, no exclamation marks.
- **State** = `StatusDot` (+ text where space allows): `Idle`/`Busy`/… kernel, `Running`/`Stopped` browser. In call tables success is the dot alone (spoken label) and only `Error` is written out, so the exception stands out. Never a pill or badge.
- **No value** is the em dash `—` from `format.NONE`. Never `N/A`, `None`, `Unknown`, `null`.
- **Time:** relative (`formatRelative`, English) in tables and detail pages; clock time (`formatTimestamp`) in the log table.
- **Empty state:** see section 7 (Empty states).
- All formatting goes through `shared/lib/format.ts`.
- No lorem, no example data, no hard-coded environment strings in the UI.

## 7. Components — the full inventory

| Need | Use | Notes |
|---|---|---|
| Text action | `Button`, `ButtonLink` | `size` `md` (h-9, default) or `sm` (h-8). Variants: `primary` (`bg-fg text-canvas`; max one per view, also the confirm button of dialogs), `secondary` (surface, strong border), `ghost`, `menu` (full-width, left-aligned, `px-2`, muted icon; rows of `PanelActions`). There is no danger variant: destructive actions are `menu` rows in the `separated` group and always confirm. Optional `icon`. `ButtonLink` for navigation that looks like a button. |
| Icon-only action / link | `IconButton`, `IconLink` | `label` required; `size` `md` (32px) or `sm` (24px). |
| Form controls | `Input`, `SearchInput`, `Select`, `Checkbox`, `FormError` | Always labelled; `hideLabel` in toolbars; `hint` for "Optional". `FormError` is the bordered inline error with an icon. |
| Modal | `Dialog` + `DialogActions` | Native `<dialog>`, `rounded-xl`, `h-14` header with Close; `DialogActions` is a `raised` footer bar. **Mount only while open.** |
| Destructive confirm | `useConfirm()` | Required for restart kernel, detach run, delete project, delete account. |
| Dense data | `TableCard`, `TBody`, `Th`, `Td` | Rows divided by `TBody` (`divide-y`); horizontal scroll on narrow screens; cells one line (`whitespace-nowrap`). |
| Runs | `RunsTable`, `useRunActions` | `domain/runs`. Read-only, links to the run page; `showProject` / `hideAccount` drop columns that would repeat the page context. Controls (`useRunActions`) live on the run page. |
| Tool calls | `CallsTable`, `CallDetail` | `domain/logs`. `hideRun` on the run page. **Every row expands in place** (click the row or its chevron; links inside the row keep their own behaviour) into `CallDetail`: the error (failed calls), then **Request** and **Response** panels side by side (stacked below lg), each with its size, a truncation note and a Copy icon button. |
| Code / JSON | `CodeBlock` | Monochrome highlighting (keys `fg`, strings `muted`, literals `fg` medium, punctuation `subtle`); `max-h-96`, scrolls, wraps long lines. `plain` for non-JSON text. |
| Usage | `Meter` | `h-1`, `line` track, `fg` fill; the number beside it carries the exact value. Always has a label. |
| Label/value rows | `PropertyList` | Values right-aligned, truncate. |
| State | `StatusDot`, `Status` | `tone` → shape: `ok` solid disc, `warn` hollow ring, `bad` diamond, `idle` dim disc. `label` mandatory; `Status` = dot + visible label. |
| Account monogram | `Avatar` | Neutral; first letter. Accounts only. |
| Frames | `Shell`, `Sidebar`, `Page`, `Metrics`, `Detail`, `Section`, `Panel`, `PanelBody`, `PanelActions` | |
| Empty / transient | `EmptyState`, `useNotice()` | |

**Not allowed:** Badge, Chip, Tag, Pill, Tabs, stat tiles outside `Metrics`, any hue, gradients, progress rings, skeleton shimmer, third-party tooltip or toast libraries, `alert()/confirm()/prompt()`.

### Feedback — which channel?

| Situation | Channel |
|---|---|
| Form submit fails | `FormError` (`role="alert"`) inside the form; dialog stays open |
| Row action fails | `notifyError(message)` (8s); toasts are inverted (`bg-fg`) with an icon |
| Row action succeeds and the UI visibly changes | nothing |
| Row action succeeds and nothing visibly changes | `notify(message)` (4s) |
| Daemon connection lost | one-line monochrome banner with a `WifiOff` icon above the page; nothing is shown while connected |

Form actions in the provider **reject**; row actions **never reject** (they notify and resolve to `false`).

### Actions

An unavailable action is **disabled, not hidden** ("Default target for agent" on the active run, `Delete account` on `default`). Destructive actions go through `useConfirm`; after a confirmed destructive action succeeds, navigate to the parent (project or list). Starting a run is a dialog: choose the account, optional name; it can take seconds because the browser may launch, so the dialog stays open with a status line until the call returns.

### Empty states

`EmptyState` with an icon and one sentence. When the missing thing is the user's next step, pass the one primary action (`New project`, `Start run`, `New account`).

### Logs

**Payloads.** The backend records each tool call's arguments and the envelope the agent received (`GET /api/activity/{id}` → `request` / `response`: `{ body, bytes, truncated }`); the list and the stream carry only `has_payload`. The UI fetches a call's payloads when its row opens (`useActivityDetail`, cached by id) and never puts them in the list state. Bodies over 64 KB are stored as a prefix (`truncated`: shown as plain text with "First N of M"). Fields named like credentials (`password`, `secret`, `token`, `*_token`, `api_key`, `authorization`, `cookie`, …) are replaced by `[redacted]` before storage; secrets inside free text (code cells, file contents) are stored as sent. Calls recorded before capture existed say "Payloads were not recorded for this call."

The global page has search, tool, run, account, errors-only (all in the URL) as a toolbar above the table, and Pause/Resume (local freeze) as the header `action`. The backend serves only the last 50 calls and the stream only the latest 10, so the provider **merges by `id`** (cap 500); never replace the list with a tick. Calls of a detached run cannot be attributed to a project or account and show only in the global log.

## 8. Accessibility (non-negotiable)

- Icon-only controls have names; decorative icons are `aria-hidden`.
- `<nav>` with `aria-current="page"`; closed sidebar is `inert`.
- Dialogs are native modal `<dialog>` with `aria-labelledby`.
- Dot-only state has `role="img"` and a label; meters expose `role="meter"` values.
- Contrast ≥4.5:1 for text in both themes; `subtle` is the lowest allowed text colour.
- Everything keyboard-operable with a visible focus ring.
- Check light and dark, sidebar open and closed, and 390px for any new screen.

## 9. TypeScript & code rules

- `strict`, `noUnusedLocals`, `noUnusedParameters` stay on. No `any`. No non-null `!` except `getElementById('root')`.
- Imports across folders use `@/…`; relative only within the same folder.
- Function components; props typed with an `interface`; no default exports except `App`.
- No `console.*`, dead code, commented-out code, or `TODO`.
- Comments explain *why*, not *what*.
- No new dependency without a written reason. Runtime deps: `react`, `react-dom`, `lucide-react`, `clsx`, `tailwind-merge`, `@fontsource/ibm-plex-sans`, `@fontsource/jetbrains-mono` (fonts bundled so the embedded binary works offline).
- Do not memoize without a measured reason (exception: provider context values).

## 10. Enforcement

`scripts/check-design.mjs` fails on: raw palette colours, hex/`rgb()` in source, arbitrary Tailwind values, font sizes other than `caption|body|title|display`, hue tokens (`accent`, `warning`, `danger`, `success`, `info`, `sidebar`), radii other than `rounded`/`rounded-lg`/`rounded-xl`/`rounded-full`, shadows other than `shadow-overlay`, gradients/blur, `animate-pulse|bounce|ping|spin`, `mx-auto`, `alert/confirm/prompt`, `console.log|debug|info`, and `any`.

It cannot judge taste. Before finishing any UI change, answer honestly:

1. Can I delete an element without losing information or an action?
2. Did I add a second way to do something the inventory already covers?
3. Does it work with the sidebar open and closed, at 390px, in light and in dark? Does it still sit on the layout grid?
5. Did any colour other than black, white or gray sneak in?
4. Does every state (loading, empty, error) look defined by this file?

## 11. Change protocol

- Changing a token, size, or component API: update this file, `tailwind.config.js` / `index.css`, `scripts/check-design.mjs`, and every call site in one change.
- Never edit `ui/dist/` by hand; run the build (embedded by `ui/assets.go`).
- The daemon accepts browser requests only from its own origin (`internal/api/guard.go`: loopback `Host`, same `Origin`, no CORS). `vite.config.ts` therefore sets `changeOrigin: true` and rewrites `Origin` to the daemon for the dev server's own origins only. Do not switch the proxy back to the string shorthand without `changeOrigin`, do not widen the rewrite to other origins, and call the API with relative paths only.
- Never change `internal/` or `cmd/` as part of UI work. If the UI needs a backend change, stop and report it.
