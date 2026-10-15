---
name: C2 (cmd & ctrl)
description: >-
  A dark, quiet table for four-player Commander. Navy-steel surfaces, one
  signal-orange accent for the action, a separate gold for whose turn it is,
  and a fixed set of game-signal rings that mean the same thing in every skin.
  Three type voices: a grotesque display face, a plain UI sans, and a mono for
  counts. The card art and the seat colours carry the table; the chrome stays out
  of the way.
source: client/src/app.css
---

# DESIGN.md: the C2 design brief

This is the brief an agent reads before building or changing UI in `client/`.
It is written from the code, not from taste: every token below is defined in
[`client/src/app.css`](client/src/app.css) or a file it names, and every rule
cites where it comes from. **If this file and the code disagree, the code wins:
fix this file in the same PR.** If you need a token that does not exist, add it
to `app.css` (and this file) rather than hard-coding a one-off value. Short
paths below are relative to `client/src/` (`lib/…`, `routes/…`) or to
`client/src/lib/components/` (`board/…`, `decks/…`, `*.svelte`).

Related reading: [ADR 0156](docs/decisions/0156-a-front-door-site-shell-front-page-and-a-discord-gate.md)
(site shell), [ADR 0111](docs/decisions/0111-action-dock.md) (action dock),
[ADR 0105](docs/decisions/0105-legal-action-highlights.md) (the ready colour),
[ADR 0125](docs/decisions/0125-a-walkthrough-that-keeps-up.md) (hints and
labels), [ADR 0119](docs/decisions/0119-a-stack-you-can-follow.md) (the stack),
[ADR 0128](docs/decisions/0128-playmats.md) (playmats),
[ADR 0158](docs/decisions/0158-signed-in-pages-tables-with-history-decks-and-the-catalog.md)
(decks and catalog), and [PLAN.md](PLAN.md) §3 and §5.

## 1. Identity

C2 is a game table first and a website second. It should feel like a calm,
well-lit table at night: dark steel surfaces, crisp type, and colour reserved
for things that mean something. Card art and the four seat colours are the
loudest things on screen; the chrome (bars, panels, buttons) is flat, quiet and
always in the same place, so a player's eyes stay on the cards. One accent marks
the action you can take; one gold marks whose turn it is; a small fixed set of
rings says what a card is doing (targeted, attacking, blocking, ready). The table
is automation-first: it does the work and offers the next sensible step, and the
manual sandbox actions are there as a fallback, never the main path
([PLAN.md](PLAN.md) §5.1). The public front-door pages may be a little more
expressive, but they are made of the same tokens.

Key characteristics:

- Dark by default (`color-scheme: dark`), navy-steel inks, flat chrome.
- One chrome accent (`--accent`, signal orange in the default skin), one turn
  signal (`--priority`, gold), and game-signal colours that never change meaning.
- Three type voices: display (Bricolage Grotesque), ui (Instrument Sans), mono
  (JetBrains Mono, for numbers and labels).
- Tablet (820px wide) is the smallest screen we design for.
- No decoration that competes with card art: no gradients on controls, no
  emoji, no illustrations.

## 2. Colour

All colours are CSS custom properties on `:root` in `app.css`. Components use
the tokens, never hex, except where a section below says a value is fixed.
Values shown are the default skin, **War room** (the bare `:root` block).

### Ground, surfaces and ink

| Token | War room | Role |
|---|---|---|
| `--bg` | `#070d14` | Page ground. `body` adds one soft pool of `--accent-glow` at the top. |
| `--bg-1`, `--bg-2`, `--bg-3` | `#0c1520`, `#121e2c`, `#1a2a3c` | Ground steps. Older code; prefer the surface names. |
| `--surface` | `#121e2c` | Cards, panels, dialogs. |
| `--surface-raised` | `#1a2a3c` | Default buttons, raised rows, option numbers. |
| `--surface-hover` | `#22354b` | Hover fill for buttons and rows. |
| `--surface-sunken` | `#0a121b` | Inputs, option rows, logs: things set into a surface. |
| `--border` | `rgba(140,180,225,0.11)` | Hairline between surfaces. |
| `--border-strong` | `rgba(140,180,225,0.2)` | Button, input and dialog edges. |
| `--fg` | `#e8eff7` | Body text and titles. |
| `--fg-muted` | `#8fa3ba` | Ledes, secondary text, hints. |
| `--fg-dim` | `#5d7187` | Placeholders, mono labels, separators. **Not for text a player must read to act**: it is 3.2 to 3.9:1 on the dark skins' grounds and surfaces. |
| `--overlay-ink` | `#fff` | What washes are mixed from (black in the light skin). |
| `--overlay-faint`, `--overlay`, `--overlay-strong` | 4%, 6%, 10% of `--overlay-ink` | Hover washes on transparent controls. |
| `--scrim` | 72% of `--bg` | Behind full-screen dialogs. |
| `--shadow-ink` | `#000` | Shadow colour (warm brown in the light skin). |

### The accent family

| Token | War room | Role |
|---|---|---|
| `--accent` | `#ff7a1a` | The one chrome accent: the primary button, selection, the eyebrow, the wordmark's Cs, focus outlines on inputs. |
| `--accent-strong` | `#ffa55e` | Hover of the primary, link colour, `:focus-visible` outline on buttons. |
| `--accent-soft` | `rgba(255,122,26,0.16)` | Selected-row wash, input focus ring, `::selection`. |
| `--accent-fg` | `#1f0b00` | Text on an `--accent` fill. |
| `--accent-line`, `--accent-glow` | 50% and 7% of `--accent` | Derived borders and glows; follow any skin or custom accent. |
| `--gold`, `--gold-strong`, `--gold-soft` | aliases of the accent | **Legacy names** from before skins. New code uses `--accent` for chrome and `--priority` for the turn. |

A player can override the accent with one colour (`display.accent`);
`accentVars` in [`lib/skins.ts`](client/src/lib/skins.ts) derives
`--accent-strong`, `--accent-soft` and an `--accent-fg` chosen for contrast.
So never hard-code an accent shade: mix from `--accent` with
`color-mix(in srgb, var(--accent) N%, transparent)`.

### Game signals

What a ring or glow **means** on the board. They keep their meaning in every
skin (the comment above them in `app.css`). Only `--ready` and `--agent` are
retuned: by the light and high-contrast skins, and `--ready` by the colour-blind
palette.

| Token | Value | Meaning | Where |
|---|---|---|---|
| `--priority` (`-strong`, `-soft`) | `#d9b45c` | You have priority; whose turn it is. Not the chrome accent. | `.seat-panel.seat-turn` edge, the priority ring in `board/PlayerIdentity.svelte`, the dock's "your turn next" ring |
| `--ready` (`-glow`, `-soft`, `-ink`) | `#2fd8ee` | This card, pile or button has a legal action **now** ([ADR 0105](docs/decisions/0105-legal-action-highlights.md)). Shown only while you owe a decision. | `.card.ready` outline in `board/Card.svelte` |
| `--target` | `#6fe3a4` | A legal target while you choose one. | `.card.targetable` |
| `--pick` | `#ffd07a` | Picked, or the source of what you are doing. | hand and strip picks, targeting arrows |
| `--attack` | `#ff7a7a` | Attacking, or a player you can attack. | `.card.attacking` glow, combat arrows |
| `--block` | `#91c3ff` | Blocking. | combat arrows |
| `--danger` (`-soft`, `-line`) | `#ff6b6b` | Destructive actions, errors, the "this pass ends your turn" ring. | `button.danger`, `.prompt-hint.error` |
| `--magenta` | `#b08aff` | A player override that is armed (hold, bluff). | `board/ActionDock.svelte`, `BluffChip.svelte` |
| `--agent` (`-soft`, `-border`) | `#c3a6ff` | A seat played by an outside AI agent ([ADR 0122](docs/decisions/0122-an-agent-at-the-table-a-local-mcp-seat.md) §7). Never the bot chip's seat tint. | `AgentChip.svelte` |
| `--mint` | `#5fd4a4` | Connected, live, success. | `.tag-connected`, Lobby `.chip.live` |
| `--rose` | `#ff8aa5` | Accent within the ribbon stack lane only. | `board/StackLaneRibbon.svelte` |

Readiness is never colour alone: the ring is an outline, and the kind of action
is a pip **shape** (bolt, drop, star, sword, shield; ADR 0105 §7). The ready
ring is outset so it never hides the keyboard focus ring.

### Seat colours

Fixed per absolute seat, the same for every viewer
([`lib/colors.ts`](client/src/lib/colors.ts)); skins do not change them.

| Seat | 0 | 1 | 2 | 3 | spectator / out of range |
|---|---|---|---|---|---|
| Colour | `#5fb0ff` blue | `#c98bff` purple | `#ffb45f` orange | `#5fd4a4` green | `#888888` |

A seat's area sets `--seat-color` and sits on `.seat-panel` (`app.css`): a 9%
wash of the seat colour from the top-right corner and a seat-tinted edge, so
whose board is whose reads before any name. Use `seatColor(seat)`; never pick
seat hues by hand.

### Colour identity

`--identity-w`, `-u`, `-b`, `-r`, `-g`, `-c` (`#eadb9c`, `#3f8fe0`, `#8d6fb3`,
`#e2533a`, `#35a866`, `#94a3b3`) are a deck's colour identity, used for the aura
on the Decks grid ([ADR 0158](docs/decisions/0158-signed-in-pages-tables-with-history-decks-and-the-catalog.md) §2).
They are fixed: no skin may override them.

### Skins

A skin is a block of palette tokens keyed on `:root[data-theme="<id>"]`; the
list is `SKINS` in [`lib/skins.ts`](client/src/lib/skins.ts). A skin overrides
ground, surface, ink, border and accent tokens only.

| id | Name | Ground | Accent | Notes |
|---|---|---|---|---|
| `warroom` | War room | `#070d14` | `#ff7a1a` | Default (`DEFAULT_SKIN`), the bare `:root` block. |
| `arcade` | Arcade night | `#0c0a1a` | `#ff4f9a` | |
| `prism` | Prism | `#0b0c10` | `#8f6bff` | |
| `classic` | Classic | `#0b0a09` | `#d9b45c` | The original gold. |
| `light` | Light | `#f3efe6` | `#9a7a2c` | Valid but **not in the picker** (`PICKER_SKINS`) until the table's on-art chips have light styles. Retunes `--ready` and `--agent`. |
| `high-contrast` | High contrast | `#000` | `#ffd400` | White borders, no ready glow. |

Two more root switches stack on any skin: `data-colorblind="1"` swaps `--ready`
for a near-white (near-black on light), and the custom accent above. All root
attributes are written by `applyRootSettings` in
[`lib/rootSettings.ts`](client/src/lib/rootSettings.ts).

**Every new colour must work in every picker skin.** Build from tokens and
`color-mix`, then look at it under at least War room, Classic and High contrast.

## 3. Type

Three voices, self-hosted (`client/public/fonts`, `@font-face` in `app.css`),
each with a metric-close fallback.

| Voice | Token | Family | Weights | For |
|---|---|---|---|---|
| display | `--font-display` | Bricolage Grotesque | 500–800 | Page and dialog titles, life totals, the step label, the wordmark, front-page hero. `h1`–`h3` default to it. |
| ui | `--font-ui` | Instrument Sans | 400–700 | Everything else: names, buttons, body copy. The `:root` default. |
| mono | `--font-mono` | JetBrains Mono | 500–700 | Counts, zone labels, P/T, sequence numbers, keyboard hints (`.kbd`), eyebrows, tags. Use `font-variant-numeric: tabular-nums` where numbers line up. |

The mono voice is part of C2's identity: counts and small labels are mono,
often uppercase with 0.06–0.14em tracking (`.tag`, `.page-eyebrow`,
`.prompt-src`, the global `h2`). Keep it to labels and numbers; running text is
always the ui voice.

There are no size tokens. The scale in use (px, from `app.css` and
`Page.svelte`):

| Size | Use |
|---|---|
| `clamp(38px, 3.9vw, 54px)` / 800 | Front-page hero title only (`routes/FrontPage.svelte`). |
| 34px / 700, −0.02em | Page title (`.page-title`, `Page.svelte`): one size on every site page. |
| 28px (`1.75rem`) / 800 | Bare `h1` outside the frame. |
| 17px / 700 | Dialog and prompt titles (`.prompt-title`); front-page lede. |
| 15px | Page lede (`.page-lede`), `.btn.lg` on the front page. |
| 14px / 700 | `.btn`. |
| 12–13.5px | The table's and dialogs' working text (`.prompt-opt` 13px, `.prompt-hint` 12.5px, `.btn.sm` 12.5px). |
| 10–11px mono | Tags, eyebrows, counts, key caps. |

Text scale is a player setting (`accessibility.textScale`, 0.9 to 1.5) applied
as `--font-scale` on `:root`'s `font-size`. Only rem and em sizes follow it, so
**use rem or em for body and reading text you add**; px is for chrome whose box
must not grow.

## 4. Space, radius, elevation, motion

**Space.** There is no spacing scale token; the code uses px. The working set is
4, 6, 8, 10, 12, 16, 24px gaps (by frequency in `client/src`), 22px card
padding (`.card`), and the frame's `--page-gutter`: 28px, 20px below 820px.
Page content gets `40px` top and `56px` bottom padding and a `28px` gap under
the head (`Page.svelte`). Reuse these rather than inventing new steps.

**Radius.**

| Token | Value | Use |
|---|---|---|
| `--radius-sm` | 4px | Key caps, tiny badges. |
| `--radius` | 8px | Buttons, inputs, `.btn`. |
| `--radius-lg` | 12px | `.card` and panels. |
| `--radius-xl` | 16px | Dialogs (`.prompt-modal`). |
| `999px` | (literal) | Pills: `.tag`, chips, scrollbar thumbs. |

Many components still use literal 6, 10 and 14px radii; new code uses the
tokens.

**Elevation.** Surfaces are told apart by a step in tone and a hairline
(`--border`). Shadows are for what floats (menus, popovers, dialogs, the dock)
and the seat panels' soft drop (`.seat-panel`).

| Token | Value | Use |
|---|---|---|
| `--shadow-sm` | `0 1px 2px rgba(0,0,0,.4)` | A lifted chip. |
| `--shadow` | `0 8px 24px rgba(0,0,0,.5)` | Menus, popovers. |
| `--shadow-lg` | `0 24px 60px rgba(0,0,0,.65)` | Dialogs. |
| `--shadow-glow`, `--ring`, `--ring-gold` | gold glow / 2px accent / 2px `--gold` ring | Stack lanes only (`board/StackLane*.svelte`). |

**Layers.** The table's z-index ladder (ADR 0111 §3, checked in code): game log
drawer 30, top bar and strip 40, vote panel 50, action dock and its sheets 55,
hints 58, mana menus 60 and 70, full-screen dialogs 200, hover zoom 300, card
context menu 3000. The site header is sticky at 30. Put a new layer on this
ladder and say where in a comment.

**Motion.** One easing, `--ease: cubic-bezier(0.2, 0.7, 0.1, 1)`. Hover and
state transitions are 120ms; dialogs fade in 160ms and rise in 220ms
(`prompt-fade`, `prompt-in`). Transition named properties, never `all`. Game
animations live in [`lib/animations.ts`](client/src/lib/animations.ts) and scale
with `animations.speed`.

**Reduced motion is honoured three ways**, and new motion must respect all of
them: the setting `accessibility.reduceMotion` (seeded from and kept in step
with the OS signal, `lib/settings.ts`) sets `data-reduce-motion="1"`, which
cuts every animation and transition in `app.css`; `animations.enabled` gates the
game animations; and components with their own keyframes add an
`@media (prefers-reduced-motion: reduce)` block. Information that motion
carries must still arrive without it (ADR 0119 §3: the outcome badge shows; the
flight does not).

## 5. Layout and screens

- **Tablet is the minimum screen.** Design at 820px wide and up
  ([PLAN.md](PLAN.md) §3: no phone-width table layouts). Below 820px the site
  frame and header wrap; nothing gets a layout of its own (`Page.svelte`,
  `SiteHeader.svelte`). The `max-width: 599px` rules that exist on the table
  (ADR 0111 §8, ADR 0119 §1) are kept but not extended; don't add new ones.
- **One breakpoint for site pages: 819px** (`max-width: 819px`). Don't add
  page-specific breakpoints.
- **Site pages sit in `Page.svelte`**: the header, then a centred column, `wide`
  (1200px, tables and grids) or `prose` (72ch, reading). The footer is drawn
  once by `App.svelte`. No horizontal page scroll at 820px.
- **The table is its own world.** `routes/Game.svelte` is fixed-position, has
  its own command bar and the action dock, and never shows the site header or
  footer. The bottom-right corner belongs to the action dock; the stack pile is
  left; the hover zoom is top-right; the coach card is bottom-left (ADR 0111 §4,
  ADR 0119 §1). New table UI must not cover the dock (`--dock-w`, `--dock-h`).
- **Installed app.** Under `display-mode: standalone` `#app` pays back the
  safe-area insets; respect them in anything fixed to an edge.

## 6. Components: what exists and when to use it

| Need | Use | Notes |
|---|---|---|
| A non-game page | `Page.svelte` with `eyebrow`, `title`, `lede`, `meta`, `actions`, `width` | Title is sentence case, display face, one size. The page's first-use tip renders under the lede (`PageTipSlot`), never over controls. |
| Site navigation | `SiteHeader.svelte`, `SiteFooter.svelte` | Grouped nav (ADR 0156 §3). Don't add links ad hoc; the Discord placements hide when no invite is configured. |
| A button on a site page, or a link styled as one | `.btn`, `.btn.primary`, `.btn.ghost`, `.btn.discord`, `.btn.sm` | 40px tall (32px `sm`), `--radius`, 700 14px. **One primary per surface.** `.btn.discord` is the one brand colour that is not ours (`#5865f2`). |
| A button in the table or a dialog | plain `<button>`, with `.primary`, `.danger` or `.ghost` | Default is the secondary style (raised surface, hairline). `button.ghost` is borderless and muted, unlike `.btn.ghost`. Disabled is 45% opacity; in the dock, buttons are **disabled, not hidden** (ADR 0111 §1). |
| A text link with intent ("See the roadmap →") | `.link-arrow` | Accent-strong, the arrow nudges 2px on hover. |
| A content block on a site page | `.card` inside the frame | `--surface`, hairline, `--radius-lg`, 22px padding; scoped by `:where(.site-page)` so the table's own `.card` is untouched. |
| A small status label | `.tag` (+ `.tag-connected`, `-disconnected`, `-connecting`) | Mono, uppercase, pill. Pages also have local `.chip` styles in the same shape (Lobby, About, Join); reuse the pattern, keep it mono. |
| An icon | `Icon.svelte` with a name from [`lib/icons.ts`](client/src/lib/icons.ts) | One stroke set: 24-unit grid, 1.75 stroke, round caps, `currentColor`. Decorative by default; pass `label` when the icon is a control's only content. Add a glyph to `icons.ts` rather than using an emoji or a new icon font. |
| The brand | `Wordmark.svelte` (`bar`, `header`, `hero`) | The mark's Cs are `--accent` and its diamond `--fg`, so it follows the skin. Keep `client/public/icons` in step (`client/tools/gen-icons.sh`). |
| A full-screen dialog not about the turn (settings, bug report, deck import, zone browser) | `.prompt-backdrop` + `.prompt-modal`, and mount `<ModalLayer />` inside its `{#if}` | Scrim with 8px blur, z 200, max 760px, `--radius-xl`, `--shadow-lg`. Body family: `.prompt-title`, `.prompt-src`, `.prompt-hint`, `.prompt-options` / `.prompt-opt`, `.prompt-count`, `.prompt-foot`. |
| Anything the game asks during play | An action-dock request: inline, or a `DockSheet` that grows up from the dock | **Never a centred modal** (ADR 0111 §2–3). The dock owns Enter / Escape; its primary sits in the corner. Register `<ModalLayer kind="sheet" />`. |
| A copy-plus-rare-action control | `SplitButton.svelte` | The wide part does the everyday thing, the narrow part the rare one. |
| A first-use tip | a `*.hint.ts` beside the feature, drawn by `HintLayer` / `TipCard` | Card plus ring, no scrim, never over the dock, a decision or an open dialog (ADR 0125 §3.5–3.6). |
| A keyboard hint | `.kbd` | Mono key cap. Dock buttons also carry `aria-keyshortcuts`. |
| An AI seat | the bot chip (dashed, seat-tinted, robot glyph) or `AgentChip.svelte` (solid violet, spark glyph) | Never swap their looks (ADR 0122 §7). |
| Not production | `EnvBadge.svelte` | Not dismissible, click-through. |

Before writing new CSS for a site page, check `app.css`'s shared primitives
(ADR 0156 §3 item 2): a page keeps its own CSS only for what is truly its own.

## 7. Copy and labels

- **Sentence case** for titles, headings, buttons and menu items ("Keep hand",
  "Try a practice game", "Sign in with Discord"). Not title case.
- **Accessible names are lower case** ("your hand", "turn and phase
  indicator") and, where tests, hints or the tutorial use them, are contract
  labels.
- **Render registered labels from [`lib/labels.ts`](client/src/lib/labels.ts)
  as `L.<key>`, never as a literal** (ADR 0125 §2; `labels.test.ts` enforces
  it). Renaming one is an edit to the registry plus the e2e specs and any
  `*.hint.ts` in the same PR. The generated list is
  [docs/labels.md](docs/labels.md).
- Name controls by their visible text. A button says what it does ("Copy
  invite", "Discard"), not "OK" or "Continue".
- Hint copy: a title of at most 32 characters, a body of at most 140, second
  person, present tense, no engine vocabulary, no rules teaching (ADR 0125 §3.4).
- Use the typographic ellipsis (`…`) and real quotes. Loading states end in `…`.
- The short name is **C2**; the full name is *cmd & ctrl* (AGENTS.md §1).

## 8. Accessibility floor

Every PR that touches UI meets this:

- **Visible focus on every interactive element.** Buttons and `.btn` get a 2px
  `--accent-strong` outline at 2px offset on `:focus-visible`; inputs get an
  `--accent` border and a 3px `--accent-soft` ring. Links get the browser's ring
  unless their component styles one. Never remove an outline without a replacement. The
  `accessibility.alwaysShowFocus` setting forces outlines on `:focus` too.
- **Contrast.** Body text is `--fg` or `--fg-muted` on a surface (both above
  6:1 in every picker skin). `--fg-dim` is for non-essential text only. Text on
  an accent fill is `--accent-fg`. Check a new colour pair under High contrast.
- **Not colour alone.** Every signal also has a shape, an outline, a word or a
  position (ADR 0105 §7). Seat colours are paired with names.
- **Reduced motion** as in §4.
- **Semantics.** `<button>` for actions, `<a>` for navigation, labels on form
  controls, `aria-label` on icon-only controls. A dialog is a `role="dialog"`
  with a name; a dock request is a non-modal dialog (ADR 0111 §2). Announce
  changes through the app's polite live region, once, not per frame.
- **Keyboard.** Every dock button keeps its shortcut, shows it in its tooltip
  and carries `aria-keyshortcuts` (ADR 0111 §1); shortcuts are listed in the
  `?` overlay. A hint never takes focus or swallows a key.

## 9. App screens and public pages

**The table, Tables (lobby), Decks, Catalog, Settings and Admin are app
screens.** Dense, quiet, predictable: tokens as given, the 12–15px working
sizes, controls in fixed places, smart defaults with the manual fallback one
menu away. Nothing animates for its own sake.

**The front door may go further.** The front page (`#/`), the gate page
(`#/request-access`), Roadmap, About and Guide address visitors who have not
sat down yet. They may use the display face at hero size, more space, a
wider measure for the pitch, and the accent on one phrase of the headline (as
`FrontPage.svelte`'s `.hot` does). They still use the same tokens, the same
`.btn` and `.card` primitives, sentence case, and the same accessibility floor.
They are not a separate brand.

**Imagery.** C2 has no photographs and no illustrations. Card images come from
Scryfall, and only where cards are the subject: the table, Decks, and the
Catalog (public by rule, since card data is never gated, [PLAN.md](PLAN.md) §3
and [ADR 0042](docs/decisions/0042-card-catalog-page.md)'s amendment). **The
front-door pages carry no card art**; the Roadmap publishes card names and
caveats only, never art or oracle text
([ADR 0092](docs/decisions/0092-public-roadmap-and-site-portal.md) decision 1,
held by a test). Where an art crop is the main view of a card, print the artist
credit (`lib/artistCredit.svelte.ts`, #2982). A playmat is drawn under a scrim
of `--bg` at the owner's wash (30–90%, default 58%) so cards keep their
contrast (ADR 0128 §8, §10). Never generate card art or card text.

## 10. Known gaps and inconsistencies

- `app.css` and `lib/icons.ts` point at a "docs/design foundations sheet" that
  is not in the repo. This file is that sheet now.
- `.card.picked` and `.card.blocking` in `board/Card.svelte` use hex near
  `--pick` and `--block` (`#ffe69a`, `#9ec7ff`) instead of the tokens. The
  tokens are canonical.
- The Classic skin's accent equals `--priority` (`#d9b45c`), so chrome and the
  turn signal share a hue there, though the comment says a skin retunes a
  signal when its accent collides.
- The stack's top-item arrow is `var(--gold)` (`lib/stackArrows.ts`), which is
  the accent alias, though ADR 0119 calls it gold; and the Decks coverage ring
  (`decks/CoverageRing.svelte`) draws its amber with `--priority`. Neither is
  the turn signal; don't copy them.
- `--gold*`, `--bg-1..3` and literal radii (6, 10, 14px) are older names and
  values still in use; don't add more.
- The global `h2` is a small mono uppercase label, not a heading size;
  `.prompt-title`, `.step h2` and others override it locally.
- `accessibility.colorblindPalette` swaps `--ready` only; the seat palette does
  not read it yet (`lib/rootSettings.ts`).
- There are no tokens for spacing, type sizes or durations; §3 and §4 describe
  the working values.

## 11. Do and don't

Do:

- Use `app.css` tokens and `color-mix` from them; check the result in War room,
  Classic and High contrast.
- Put every site page in `Page.svelte` and build it from `.btn`, `.card`,
  `.link-arrow` and `.tag`.
- Keep one primary action per surface, and in the dock keep it in the corner.
- Use `--priority` for turn and priority, `--accent` for chrome, and the signal
  tokens only for what they mean.
- Render contract labels from `labels.ts`; write sentence case.
- Use mono for counts and labels, with tabular numbers.
- Give every control a visible focus state and every icon-only control a name.
- Design at 820px and up, and look at 1440px too.

Don't:

- Hard-code a hex for chrome, a seat hue, or an accent shade.
- Use a game-signal colour for decoration, or the accent to mean "your turn".
- Open a centred modal for something the game asks during play.
- Add a phone-width table layout or a new page breakpoint.
- Put card art on the front-door pages, or anything from Scryfall behind a
  sign-in.
- Use emoji, gradients on controls, glass panels, or decorative shapes.
- Add motion that ignores `data-reduce-motion` or `prefers-reduced-motion`.
- Hide a dock control that cannot be pressed; disable it.
