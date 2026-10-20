# ADR 0028 — Right-click admin context menu for per-card overrides

**Status:** Implemented · 2026-09-10 · Branch `feat/admin-context-menu`
**Amended:** 2026-09-16 · #170 closeout: the stale #164 references corrected, and the two unshipped items handed to their own issues (see the amendment at the end)
**Amended:** 2026-10-19 · [#4131](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4131): the table-wide sandbox verbs are gated (see the last amendment)
**Relabelled:** 2026-10-09 · [ADR 0143](0143-gameplay-settings-overhaul.md) §3.2: the "Enable admin overrides" setting (`gameplay.adminOverrides`, unchanged) is now **Manual card controls on right-click**, under Settings → Gameplay → Advanced.
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
  "before #164 ships", but #164 had shipped in #171.) (Amended
  2026-10-04 by [ADR 0115](0115-commanders-die.md): the commander now
  dies into the graveyard first, and the item answers the CR 903.9a
  state-based action's question "yes" in the same click.)
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

## Amendment (2026-10-13): the `reveal_card` verb ([#671](https://github.com/krakenhavoc/cmd_and_ctrl/issues/671))

The 2026-09-16 amendment recorded the owner's decision; this one records
how it was built, and a place where the build departs from the issue's
wording. It is a third exception to §1 (after `to_bottom` and
`remove_from_combat`): a new verb, because no existing action can make a
seat a knower of a hand card.

1. **One verb, two shapes.** `reveal_card {card_id, to?}`. Hand cards
   only: a card anywhere else is refused (`ErrNotInHand`), because a card
   in a public zone is already known and one in a library is hidden by
   rules this menu has no business overriding. Gated by
   `requireCardController` like the rest of the menu (the admin bypasses),
   and the client shows the rows only with "Manual card controls on
   right-click" on.
2. **Without `to`: "Reveal to table".** `game.RevealForEffect`, so every
   seat becomes a knower (eliminated seats too, as for any reveal) and the
   grouped `reveal_cards` events fire; the `reveals` frame and banner show
   it. The label is "revealed a card from their hand".
3. **With `to`: "Show to <player>".** `AddKnower(to)` and nothing else:
   no reveal frame, no banner, no wire change beyond that one seat's view
   carrying the card, which `keepKnownInHandZone` and `Hand.svelte`
   already handle. The target must be seated and not eliminated, and not
   the owner. The Comprehensive Rules have no action for this
   (CR 701.20a reveals to all players), so it is a table courtesy.
4. **The event.** The issue asked for "a distinct event kind, or no
   event" for the private show. The build emits `EventRevealCards` with
   `Target` set and `RevealSeq` zero instead, because the log projection
   (`protocol/log.go` `revealEntry`, `renderRevealText`) was written
   ahead of this verb for exactly that shape: it words it "P1 revealed a
   card from their hand to P2" for every viewer and names the card to
   nobody, with a test pinning it. The reveal frame keys on a non-zero
   `RevealSeq` and skips it. The event comment still defines
   `EventRevealCards` as a table reveal for every producer except this
   one.
5. **How long it lasts.** As long as `KnownBy` does; the verb adds no
   expiry of its own. Every exit from a hand already does the right
   thing: a move to the library (`ClearKnown`) or face down in exile
   (the knowledge is replaced) drops it, a mulligan clears hand and
   library, and a move to a public zone makes every seat a knower anyway.
   A test pins the library exit.
6. **Undo.** One entry, from the ordinary `Room.ApplyAction` path;
   `cloneCard` deep-copies `KnownBy`, so undo takes the knowledge back.
7. **`known_by_seats`.** `CardView.known_by_seats` (omitempty) lists the
   other seats that know a hand card. `FilterViewFor` stamps it on the
   viewer's own hand cards only and `redactCardForViewer` zeroes it
   everywhere else; the unexported `knowers` map still never reaches the
   wire. The client shows a small count badge on the card, with the names
   in its tooltip.
8. **Bots.** Not enumerated: it is a sandbox verb, so it is absent from
   `internal/legal` and from the `aiseat` allow-list.

## Amendment (2026-10-19): the table-wide sandbox verbs are gated ([#4131](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4131))

§4 says the menu "grants no authority" because every verb it sends passes
`requireCardController`. That was true of the verbs that name a card. A
second group took no caller or ownership check at all, on a "sandbox among
friends" reading, and #4025's review found the one that is not a marker
flip: `announce_trigger` put a real stack item on the table under a
controller of the caller's choosing, naming any card as its source.

**Owner decision (Ian, 2026-10-19): gate the risky ones.** Only the
affected player, or the controller of the affected card, can use them, and
there is no table-wide override beyond what already exists for the admin
(`Caller == uuid.Nil`, which keeps every verb below for repairs). The
rule per verb, in `actions/actions.go` and `actions/table_verb_gates.go`:

| Verb | Who may send it |
|---|---|
| `announce_trigger` | As yourself (`Player` is in `playerScopedActions`), with a source you control. Once the source has left the battlefield, its owner and the player who controlled it as it last existed there (the last-known controller on its `EventLTB`, CR 603.3a and 603.10a) also qualify, so a stolen creature's dies trigger can still be announced by its thief (`requireTriggerSource`) |
| `set_commander_damage` | Only with yourself as the damaged player (`to` is the caller) |
| `mark_damage` | The card's controller, with a negative delta as with a positive one (it used to be open "so opponents can undo") |
| `set_goaded` | The creature's controller, who may set or clear any goad on it; any other seat only goads as itself (`by` is the caller). Clearing every goad is the controller's alone |
| `set_promise` | Either end of the pair (`from` or `to`) |
| `set_monarch`, `set_initiative` | A seat takes the marker for itself, or releases or hands it on while it holds it. Crowning another seat, or clearing a marker you do not hold, is refused (`ErrMarkerNotYours`) |
| `advance_step`, `clear_combat` | The active player or the priority holder (`ErrNotTurnOrPriority`), who are the two seats that may pass. While priority is parked for a block declaration the active player is the one |
| `start_vote` | As yourself (`Player` is in `playerScopedActions`) |
| `end_vote` | The player who called the vote, or any seat once that player is eliminated or gone, so a vote never sits open for ever (`ErrNotVoteInitiator`) |
| `counter_spell`, `counter_ability` | Unchanged: the priority holder. A counter effect may name any object on the stack, so the object's controller is the wrong test, and the priority holder is the seat casting the counter |

A seated admin (Caller is its seat, `Action.Admin` set, ADR 0110 §3) keeps
the card overrides, so the card-controller gates read `overrideCaller` as
before; it also keeps `advance_step` and `clear_combat`. The player gates
read `Caller`, as every `playerScopedActions` verb does, so a seated admin
acts as its own seat there.

What does not change: the bot's improvisation dispatches with
`Caller == uuid.Nil` (ADR 0033 §8), so its four sandbox verbs are
untouched; the tutorial table sends nothing in this group; the dev-only
spawner is an HTTP route, not an action. ADR 0167 §4's sandbox list still
closes all of these to a controlled turn, before any of these gates run.

**The client** hides what a seat can no longer use rather than failing
silently. The monarch and initiative buttons were already on a seat's own
identity only, and the promise `+`/`-` already sends `from: viewer`, so the
menu and panels already matched. The two that did not: "Clear ALL combat"
in the card menu is offered only to the active player, the priority holder
and the admin (`mayClearCombat`), and "end vote" shows only to the player
who called the vote or, once that player has gone, to any seat
(`mayEndVote`, in the action dock and in `VotingPanel`). The refusals
reach the client as `bad_request` frames with a sentence a player can read
(`ws.classifyActionError`).
