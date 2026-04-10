# ADR 0002 — Client framework

- **Status:** accepted
- **Date:** 2026-04-10
- **Sprint:** S01 — Go server + client scaffold
- **Deciders:** project owner
- **Confirmed in:** S03. Tentative qualifier dropped after the
  S01–S03 sprints landed with Svelte 5 + runes working cleanly. No
  collaborator friction, no missing libraries, no regrets.

## Context

PLAN.md §4 specifies "TypeScript + Vite. React or Svelte for menus, lobby,
deck manager, chat. PixiJS canvas for the play area." PLAN.md §7.2 leaves the
React-vs-Svelte-vs-Vue call open as "familiarity beats theoretical best".

Since the play area is PixiJS (a canvas-based renderer that doesn't care
what framework it lives in), the framework choice only affects menus, lobby,
deck manager, and chat — the "chrome" around the game. Performance is
essentially a non-factor at our scale (≤8 concurrent users).

## Options

### A. React

The default. Biggest ecosystem, best tooling, most tutorials, most PixiJS
integration examples (`@pixi/react`). If a collaborator has touched modern
web frameworks, they've touched React.

**Pros:** ubiquity, hiring-style familiarity, `@pixi/react` exists if we want
to drive Pixi declaratively.

**Cons:** verbose for simple apps, render model mismatches WebSocket-driven
state updates (without a state library like Zustand/Jotai). Lots of boilerplate
for a hobby project.

### B. Svelte

Compiler-based. Reactive assignments, built-in stores, no virtual DOM. Much
less boilerplate than React for simple components. Stores map cleanly onto
"incoming WebSocket delta → UI update" without an extra state library.

**Pros:** less ceremony, faster iteration, built-in reactivity fits the
WebSocket-stream model, smaller bundles.

**Cons:** smaller ecosystem, fewer PixiJS integration examples, less familiar
to collaborators who've only used React.

### C. Vue

Competent third choice. Nothing is wrong with Vue. It loses on both "most
familiar" (that's React) and "least ceremony" (that's Svelte) so it has no
clear lane here.

## Decision

**Svelte**, tentatively.

Rationale:
- A game client is fundamentally "server state streams in → UI reflects it".
  Svelte stores are the closest thing to that mental model in any modern
  framework without extra glue.
- This is a hobby project optimising for momentum, and "less boilerplate" is
  worth more than "familiar to more people" at our scale.
- The play area is PixiJS (framework-agnostic), so the framework only matters
  for the chrome — where Svelte's ergonomic edge is biggest.
- Bundle size matters for a first-load feel even on a low-traffic VPS.
- "Tentative" — if the first week of S01 reveals collaborator friction or a
  missing library, swapping to React during S01 costs a day. Swapping after
  S05 would cost a week.

## Consequences

- `client/` scaffolds with `npm create vite@latest client -- --template svelte-ts`
  shape (recreated by hand to avoid the interactive scaffolder).
- State from the server lives in Svelte stores (`writable`, `derived`).
- If we introduce `@pixi/react`-style declarative Pixi later, we'll write a
  thin Svelte wrapper — or just imperative Pixi directly, which is fine.

## Revisit if

- A collaborator is blocked by unfamiliarity after one week of trying.
- We hit a missing library (e.g., no good Svelte equivalent for a critical
  third-party component).
- We decide to ship `@pixi/react` and the React ecosystem becomes a net win.

## References

- [Svelte](https://svelte.dev)
- [Vite](https://vitejs.dev)
- [PixiJS](https://pixijs.com) — framework-agnostic by design
