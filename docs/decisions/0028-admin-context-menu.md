# ADR 0028 — Right-click admin context menu for per-card overrides

**Status:** Implemented · 2026-09-10 · Branch `feat/admin-context-menu`
**Amended:** 2026-09-16 · #170 closeout: the stale #164 references corrected, and the two unshipped items handed to their own issues (see the amendment at the end)
**Issue:** [#170](https://github.com/krakenhavoc/cmd_and_ctrl/issues/170)

## Context

The engine automates a growing slice of the catalog, but most cards
are still unwired and some wired ones are wrong. When that happens the
table has no recourse: there is no way to send a permanent to the
graveyard by hand, no way to put a counter on something the engine
forgot, no way to route a commander to the command zone, no way to
undo a mis-declared block. The game simply stalls until a developer
opens `gamecli`.

The only manual affordances that existed were a hidden `Shift+click`
chord that adds a `+1/+1` counter (S17 sub-PR 5, always meant to be
temporary) and three "hand / field / lib" buttons inside the zone
browser. Neither is discoverable, and between them they cover a tiny
fraction of what a stuck game needs.

The guiding principle for this work, in the owner's words: *the player
can always fall back to doing it by hand.*

## Decisions

### 1. The menu drives existing actions only — no new verbs

Every row in the menu resolves to an action the server already
implements: `move_card`, `add_counter`, `tap` / `untap`, `mark_damage`,
`declare_attacker`, `declare_blocker`, `clear_combat`, `set_goaded`,
`sacrifice_permanent`, `activate_mana_ability`, `activate_ability`.

This keeps the feature honest — it is a *surface* for capability the
engine already had, not new engine behaviour — and it means the menu
inherits the server's authorisation model for free (see §4).

One exception was unavoidable and is covered in §7.

### 2. A module-scope store, not prop drilling

`client/src/lib/contextMenu.ts` is a `writable<CardMenuOpen | null>`
in the same shape as `zoneBrowser.ts`. `Card.svelte` calls
`openCardMenu({ card, x, y })` from its `oncontextmenu` handler;
`Board.svelte` mounts one `CardContextMenu` bound to the store.

The alternative was threading a `zone` + handler pair through
`BattlefieldRow`, `Hand`, `CommandZone`, `PileButton` and
`ZoneBrowserModal` so each render site could say where its cards live.
Instead the menu resolves the card's zone from the authoritative
snapshot (`locateCard`), so **no** Card render site needed a new prop.
That also means the menu works from inside the zone browser modal,
which a per-parent handler would have had to opt into separately.

### 3. Menu assembly is a pure module with unit tests

`contextMenu.logic.ts` builds `MenuSection[]` from
`(view, card, viewerID, isAdmin)` and nothing else — no DOM, no
stores. `contextMenu.test.ts` pins the option set per zone and the
exact wire payload of every move, so a regression in the `src` / `dst`
owner stamping fails a test instead of silently sending a card to the
wrong player's graveyard. Same split, and same reasoning, as
`zoneBrowser.logic.ts`.

The `MenuItem` tree (label + one of: action, submenu, prompt,
ability-handoff) is deliberately generic, so a future forced prompt
can build a one-section, two-item menu and reuse the same component
rather than growing another modal.

*Corrected 2026-09-16:* this paragraph first named
[#164](https://github.com/krakenhavoc/cmd_and_ctrl/issues/164)'s
commander-zone prompt as that consumer. #164 had already closed on
2026-04-23 through #171, before this menu existed, and its prompt is
the `optional_replacement` pending choice in `ChoicePromptModal.svelte`.
Nothing builds a forced prompt on the `MenuItem` tree yet.

### 4. Per-player setting, off by default, NOT gated on the admin role

The issue offered a choice: hook the toggle into the admin-authz check,
or make it an independent per-player "sandbox mode". We took the
second.

The reason is that the toggle grants no authority. Every action the
menu sends passes through `requireCardController` server-side, which
already restricts a seated player to cards they control and lets an
admin (`Caller == uuid.Nil`) act on anyone's. A non-admin who enables
the setting can do exactly what they could already do through the zone
browser's buttons — just with more of it, and more conveniently.
Gating the toggle would have added an authz surface that protects
nothing.

The menu itself mirrors that gate: `canOverride` returns false for a
card the viewer doesn't control, and the panel renders an explanatory
line instead of a list. This keeps the client from offering clicks the
server would bounce.

Default is off, so a player who has never opened Settings sees no
behaviour change.

### 5. Right-click subsumes the ability popover rather than fighting it

Right-click was already taken: it opens `ManaAbilityMenu` for a
permanent's mana / activated abilities. Rather than inventing a second
chord, the override menu lists those same abilities as its **first
section** when the setting is on, handing the click back to
`Board.svelte` so the existing sacrifice-picker and targeting flows run
unchanged. With the setting off, right-click behaves exactly as before.

One gesture, one menu, no lost functionality either way.

### 6. `Shift+click` for `+1/+1` is removed, not kept alongside

The S17 chord is deleted. It was documented as a stopgap in its own
commit message and in [ADR 0013](0013-replacement-effects.md); keeping
it would leave two mental models for the same operation, one of which
is invisible.

The replacement is strictly more capable — both directions on every
counter type in `COUNTER_STYLES`, plus a custom-counter prompt — and
lives behind a labelled setting with help text, which is a far better
discovery story than an undocumented modifier key. The one real cost
is that the affordance now needs a one-time opt-in; that is the point
of the issue.

### 7. `to_bottom` on `move_card` — the one server change

`game.MoveCard` always `PushTop`s, so "put this on the bottom of your
library" was not expressible. The issue lists that destination for both
the battlefield and the hand, and it is exactly the kind of thing a
manual override exists for (bottoming a card an unimplemented scry or
Brainstorm should have handled).

`move_card` gains an optional `to_bottom` flag, dispatched to a new
`Game.MoveCardByIDToBottom`. That method calls the existing
`MoveCardByIDAsCommander` unchanged and then reorders the destination
zone, rather than duplicating the replacement pipeline. If a
replacement rewrote the destination — a commander routed to the command
zone by CR 903.9 — the card is not in `dst`, the `Remove` fails, and the
reorder is a no-op. The replacement's destination wins, which is
correct.

*Corrected 2026-09-17 ([#707](https://github.com/krakenhavoc/cmd_and_ctrl/issues/707)):
the reorder could not survive a PAUSE. When the CR 903.9 prompt is
queued the card has not moved at all, so the `Remove` failed for a
commander that was going to the bottom of its owner's library anyway,
and it landed on top when its owner declined. The bottom now rides the
shared exit primitive's route (`zoneRoute.ToBottom`, ADR 0013 §5f),
which applies it against the SETTLED destination after the answer
arrives — same rule as before, one prompt later, and only for a library
destination, which is the only zone whose bottom means anything.*

### 8. Drill-down submenus, not flyouts

Rows with children replace the panel body and add a "back" row instead
of opening a second floating panel. One `position: fixed` box means one
set of viewport-clamping maths, it degrades gracefully on a narrow
window, and focus management is trivial — after every content swap the
first enabled control takes focus, so the menu is fully keyboard
operable. Escape closes it from any depth; a `pointerdown` outside
closes it too, before the card underneath gets its own right-click.

## Consequences

- Any card, in any zone, can be moved anywhere by hand. An unwired or
  buggy card no longer wedges the table.
- The commander-zone replacement (CR 903.9) is manually triggerable
  from the battlefield ("Graveyard → command zone"), so the replacement
  path can be exercised by hand. (Corrected 2026-09-16: this line said
  "before #164 ships", but #164 had shipped in #171.)
- Settings schema goes to v7 (`gameplay.adminOverrides`). The migration
  is a pure default-fill; nothing to rescue.
- Player-level counters (poison / energy / experience / rad) are
  deliberately NOT in this menu — it is a card menu, and those belong
  on the player panel. `experience` and `rad` still have no UI at all;
  that is a separate gap.
- "Reveal a card to a specific player" from the issue's hand-overrides
  list is **not** implemented: there is no reveal action server-side,
  and the `KnownBy` machinery has no external entry point. That is a
  real feature, not a menu row. (2026-09-16: still true of the action
  layer. It is now tracked in
  [#671](https://github.com/krakenhavoc/cmd_and_ctrl/issues/671), and
  the engine side has since landed; see the amendment below.)

## Amendment (2026-09-16): #170 closeout — reveal and remove-from-combat

#170 closes with this ADR's menu shipped. Two of its items did not
ship, and each now has its own issue.

**Reveal a hand card:
[#671](https://github.com/krakenhavoc/cmd_and_ctrl/issues/671).** The
Consequences bullet above calls this a real feature. It is now a small
one, because the engine pieces exist. `game.RevealForEffect`
(`server/internal/game/reveal.go`) makes every seat a knower and emits
the reveal events that the S22 broadcast frame (#549) and
`RevealBanner.svelte` display. Thoughtseize already gives one seat
knowledge of a hand card with `AddKnower` and no reveal event, and
`keepKnownInHandZone` plus `Hand.svelte` already send that card to
that seat and draw it face up. The owner decided the shape on
2026-09-16:

- A hand card gets two rows: **"Reveal to table"**, backed by
  `RevealForEffect`, and **"Show to <player>"**, backed by
  `AddKnower(player)` only. The private show does not emit
  `EventRevealCards` with a target, because that event kind means "shown
  to the whole table" and the reveal frame reads it that way.
- Knowledge **persists until the card leaves the hand**.
- The action is **undoable**: undo restores the card's knowers.
- It sits **behind the "Enable admin overrides" setting** (§4) and
  passes `requireCardController` like the rest of the menu.
- The owner sees which seats know each of their hand cards, through a
  `known_by_seats` field sent only on the viewer's own hand cards.
- It is **not offered to bots**.

The `reveal_card` verb is a new server action, so it is a second
exception to §1, alongside `to_bottom` (§7). #671 amends this ADR with
the verb's details before any code. The Comprehensive Rules
(September 19, 2025 edition) have no action for showing a card to one
player: CR 701.20a defines revealing as showing a card to all players,
and CR 701.20e's "look at" is the one-player form an effect can
instruct. So "Show to <player>" is a table courtesy the sandbox allows,
not a rules action. Open PR #643 adds the public log line for reveals.

**Take one creature out of combat:
[#672](https://github.com/krakenhavoc/cmd_and_ctrl/issues/672).** The
issue asked to "toggle attacking / blocking". The menu can declare or
re-declare an attacker or blocker, but the only way back out is "Clear
ALL combat", which wipes every declaration on the table. There is no
server verb to remove a single creature from combat (CR 506.4), so this
is a partial, and #672 adds the verb and the row. It did not block
closing #170.
