# 0047 — Keyboard shortcuts: one dispatcher, a server-gated keymap, overrides-only persistence

**Status:** Accepted
**Date:** 2026-09-14
**Relates to:** [ADR 0009](0009-smart-priority-autopass.md) (the false-stop / false-skip
asymmetry), [ADR 0033 §1](0033-ai-bot-seat.md) (`legal_moves`), `client/src/lib/timing.ts`
(what S31 sub-PR 2 deleted and why)

## Context

Passing priority is the most-pressed action in any game of Magic, and in this client it
was a mouse trip to the phase widget every single time. The only keyboard affordance that
existed was `,` for the settings panel, hard-coded in `Settings.svelte`.

Adding a global key handler to a client that already has **sixteen** local ones is where
this gets dangerous. `App.svelte`, `Settings.svelte`, `Game.svelte`, `Card.svelte`,
`CardContextMenu`, `ManaAbilityMenu`, `ZoneBrowserModal`, `ChoicePromptModal` and the
whole cost-modal family all listen on `window` or on an element, and almost all of them
are Escape-to-close plus Enter-to-confirm. A global layer that swallowed either key would
break every prompt in the game.

## Decision

### 1. One dispatcher, and the decision lives in a testable module

`client/src/lib/shortcuts.ts` owns the action table, the chord grammar, conflict
detection, the enabled/disabled rule and the entire dispatch decision, as pure functions
over plain data — no DOM, no stores, no Svelte. `ShortcutLayer.svelte` (mounted once, at
the app shell) reads the event, calls `dispatchShortcut`, and invokes the returned
handler. That is the same split `gameLog.ts`, `attackAll.ts`, `reveals.ts` and `timing.ts`
already use, and it is what lets "am I typing?", "is a modal up?" and "is this move
legal?" be unit tests instead of manual QA.

The game route publishes its handlers and the seat's situation through
`lib/shortcutRuntime.ts` and takes both down on destroy. With nothing registered, the
game-scoped rows report "only at a game table" and the global rows (settings, help, mute)
keep working from the lobby.

### 2. Typing beats everything; modals beat the keymap

Checked in that order, before the binding is even looked up:

1. **Typing.** `isTypingTarget` treats every `input`, `textarea` and `select` as typing —
   not just the text-ish ones, because a focused checkbox uses Space and a number spinner
   uses the arrow keys — plus `contenteditable` (including a deep target inside the
   subtree) and `role="textbox|searchbox|combobox"`. That covers the deck-paste box, the
   bug-report form, the mulligan-size and undo-limit spinners, and chat if it returns.
2. **Modals.** Not a list of open-flags. Each modal mounts a `<ModalLayer />` inside its
   own `{#if}`, which registers a layer in `lib/modalLayers.ts` for as long as it is on
   screen. The global layer stands down while the set is non-empty. A flag list would be
   correct until the next modal is added; this way the failure mode for a forgotten line
   is **visible** (shortcuts fire behind your dialog) rather than invisible (a stale flag
   that disables the keymap for the rest of the session). Registration is a set of ids,
   not a counter, so a double unregister across a keyed re-mount cannot drive the depth
   negative.

A modal that wants a key handles it itself — which is what every modal in this client
already does for Escape, and what the settings panel now does for its own toggle key.

### 3. `event.key`, not `event.code`

`event.code` is the physical key position: the key labelled Q on a US board reports
`KeyQ` on AZERTY too, where the keycap says A. That is right for WASD movement and wrong
for a mnemonic keymap. Our shortcuts mean "L for log", "U for undo", and a help overlay
that prints `L` has to match the letter printed on the key the player presses.
`event.key` is the character the layout actually produced, so it does.

The cost of `key` is the shift problem: `?` is Shift+/ on a US layout, Shift+ß on a
German one, and unshifted on others. The rule:

- Single characters are lower-cased. `A` and `a` are one key.
- **Shift is recorded only for ASCII letters and for named keys.** For any other
  printable character it is dropped, because the character the layout produced already
  encodes the shift. So `?` is the chord `?` on every layout, while `Shift+a` stays
  distinct from `a` and `Shift+Tab` stays distinct from `Tab`.
- Modifiers are emitted in a fixed order — `Ctrl+Alt+Shift+Meta+key` — which is what
  makes two chords comparable with `===`, the only thing the dispatcher and the conflict
  detector ever do with them.

### 4. A shortcut is enabled exactly when the move is legal, and the server says so

`passLegal` is `GameView.legal_moves.some(m => m.kind === "pass")`, via
`timing.hasPassMove`. It is not a re-derivation of the priority rules — S31 sub-PR 2
deleted a pile of exactly that and this ADR is not growing it back. An absent move list
is `undefined`, meaning "no information", and the rule stays permissive: a false yes
costs a rejected click, a false no costs the player a window they were entitled to (ADR
0009 §3).

Everything else is a read of state the route already computes for the equivalent button:
`activePlayer` for pass turn, `undos_remaining` for undo, `planAttackAll` for attack-all.
Key and button fire the same function, so a fix to one is a fix to both.

Attack-with-all is **inert with more than one opponent**, and says so. `attackAll.ts`
refuses to guess which seat a wide swing is aimed at; the key inherits the refusal rather
than inventing a default.

### 5. Space is bound, and defers to the focused control

Space is the right key for pass priority and the wrong key to take globally: it activates
the focused button and scrolls the page. So a binding whose key is Space or Enter (with
no modifier beyond Shift) fires only while focus is somewhere inert — the body, the board
background — and yields the moment the player has tabbed to a control. Derived from the
chord, not declared per action, so the rule follows a user who moves something else onto
Space.

Enter is not bound at all, and Escape, Enter, Tab, Shift+Tab and Shift+Enter are in
`RESERVED_CHORDS`: not shipped as a default, and refused by the rebinding UI.

Key auto-repeat never fires an action. Holding Space passes one priority window.

### 6. Persist overrides, never a materialised map

`settings.shortcuts.bindings` stores only the rows the user deliberately changed. An
absent row resolves against `shortcuts.ts`'s current default, so retuning a default later
reaches everyone who never expressed an opinion — and everyone who did keeps their choice.
An explicit unbind is stored as `""`, which is why this cannot be "a full map with holes".

This is the same "don't clobber an explicit decision" rule the v2→v3 `stepStops` migration
had to reconstruct by structurally comparing against the default; here the storage shape
makes it free. `sanitizeOverrides` scrubs the blob on every load: unknown action ids,
unparseable chords and reserved chords are dropped, survivors are canonicalised, and an
override equal to today's default is dropped as no longer a choice.

Schema version 9 → 10.

### 7. Discoverability is generated, not written twice

The `?` overlay renders from `SHORTCUTS`, so an action cannot appear in the overlay
without being dispatchable and cannot be dispatchable without appearing in the overlay.
Rows show the **effective** chord (defaults plus overrides) and grey out with the reason
the move is illegal right now. On-screen controls carry the same binding in their tooltip,
read from the same map, so a hint can never advertise a key that does nothing.

## Default keymap

| Key | Action | Enabled when |
| --- | --- | --- |
| `Space` | Pass priority | `legal_moves` contains a `pass`; defers to a focused control |
| `T` | Pass turn | you are the active player |
| `H` | Hold priority (toggle) | at a table |
| `Shift+P` | Autopass (toggle) | at a table |
| `A` | Attack with all | declare-attackers, ≥1 eligible creature, **exactly one** opponent |
| `U` | Undo | undo budget remaining (admin: unlimited) |
| `D` | Draw a card (sandbox) | at a table, seated, game live |
| `L` | Game log | anywhere at a table, including after elimination |
| `M` | Mute | always |
| `,` | Settings | always — unchanged from S11.5 |
| `?` | Shortcut overlay | always |

Reserved and never bound: `Esc`, `Enter`, `Tab`, `Shift+Tab`.

No default uses Ctrl or Cmd, so nothing shadows close-tab, new-tab, reload, new-window or
browser find. No default uses Alt, because AltGr is Ctrl+Alt on several European layouts
and an Alt binding would eat a letter those players need in order to type. No default is
a function key. A test asserts all of this against the table.

## Consequences

- **Adding a modal now has a step**: put `<ModalLayer />` inside its `{#if}`. Forgetting
  it means shortcuts fire behind the dialog.
- **Adding a shortcut is one entry in `SHORTCUTS` plus one handler** in the route's
  `registerShortcutHandlers` call. The overlay, the rebinding UI and conflict detection
  all follow for free.
- Two actions can share a chord. That state is reachable, visible in both the overlay and
  the Settings rows, and broken deterministically by declaration order rather than being
  silently repaired at load — silently deleting one would make the panel lie about what it
  stored.
- `openSettings()` gained an optional tab argument. Two call sites that passed it as a
  bare event handler had to become `() => openSettings()`.
- Nothing on the server changed.

## Alternatives considered

**Physical-position matching (`event.code`).** Rejected: our keymap is mnemonic, and a
cheat sheet that prints `L` while the player's keycap says something else is worse than no
cheat sheet.

**A `modalOpen` boolean the routes maintain.** Rejected: it is a list of flags wearing a
different hat, and it goes stale the first time someone adds a dialog.

**Storing the full binding map.** Rejected: it freezes every default at the version the
user first loaded, which is the opposite of what the rest of this schema's migration chain
works so hard to get right.

**Leaving Space unbound.** Rejected: it is the action players press hundreds of times per
game, and the focus-deferral rule buys the speed without costing keyboard navigation.
