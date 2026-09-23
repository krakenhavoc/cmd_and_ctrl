# ADR 0091 — Hideaway: a face-down exile its source's controller may look at, linked to that source, with a free play

**Status:** Accepted · 2026-09-23 · S43 — hand special actions and face-down objects
**Issues:** [#1331](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1331) (this seam),
[#1306](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1306) (deck tracker "Aang is so flashy" — Rabble Rousing)
**Trackers:** [#886](https://github.com/krakenhavoc/cmd_and_ctrl/issues/886) (S43 — hand special actions and face-down objects),
[#889](https://github.com/krakenhavoc/cmd_and_ctrl/issues/889)
**Numbering:** swept with the AGENTS.md §4 loop on 2026-09-23 — `git fetch origin --prune`, then
`git ls-tree --name-only <ref> docs/decisions/` over every remote head plus the recent issue claim
comments. The highest number present anywhere is **0090** (`0090-preparation-cards.md`, reserved on
#1328 and since merged to `develop` in #1344); `0091` appears on no branch, and the claim comment on
#1331 reserves it.

**Related:** [ADR 0069](0069-face-down-objects.md) (the face-down kinds and their viewers table,
amended here), [ADR 0082](0082-casting-face-down-and-turning-face-up.md) and its amendments (the
face-down machinery), [ADR 0066](0066-granted-cast-and-play-permissions.md) (per-object cast and play
grants, and the grant-not-inline posture), #1239 (the epoch fix to the effects package's
exiled-with record).

---

## Context

CR 702.75a, pinned edition (`MagicCompRules 20260819.txt`, effective August 7, 2026):

> Hideaway is a triggered ability. "Hideaway N" means "When this permanent enters, look at the top
> N cards of your library. Exile one of them face down and put the rest on the bottom of your library
> in a random order. The exiled card gains 'The player who controls the permanent that exiled this
> card may look at this card in the exile zone.'"

Every hideaway card then prints a linked ability (CR 607.2a) — "you may play the exiled card without
paying its mana cost", behind a condition. And three more rules decide the details:

- **CR 406.3** — once a player is allowed to look at a card exiled face down, "that player may
  continue to look at that card until it leaves the exile zone … even if the instruction allowing
  the player to do so no longer applies".
- **CR 406.3a** — a card exiled face down that may be played from exile is turned face up just
  before it is announced.
- **CR 305.2b / 305.3** — a land played during a resolution still needs a land drop, and only on its
  controller's turn.

`engine-seams.md`'s face-down row named hideaway as its last open item: "a face-down EXILE with a cast
permission rather than a CR 708.2 object". #1331 listed the three missing pieces: no `FaceDownKind`
whose viewer is the controller of the exiling permanent, nothing linking the permanent to "the
exiled card", nothing granting the free play of that card.

## Decision 1 — A seventh kind, `FaceDownHidden`, whose viewer is another object's controller

`FaceDownHidden` (`"hideaway"`) is an exile kind like `exiled` and `foretold`: `IsPermanentState` is
false, so it has no CR 708.2 body and keeps its catalog entry for the play. Its row in ADR 0069's
viewers table is the first whose answer is not a property of the card: **the controller of the
permanent that exiled it**. `faceDownViewersLocked` reads it through the link (Decision 2), so it is
right when the card lands.

Reusing a kind fails for rules reasons: `foretold` names the owner and `exiled` names nobody. The
controller of a Windbrisk Heights is usually its owner, and "usually" is exactly the bug a reuse
would ship.

Because the answer names another object, it can change after the card lands. When the permanent
changes hands, `hideawayKnowersSweepLocked`, run with the state checks, adds the new controller as a
knower. It does not count as "fired" and changes no characteristic. Nobody is ever removed, because
CR 406.3 lets a player who has looked keep looking. That also settles the permanent leaving the
battlefield: nobody new may look, and whoever already did still can.

## Decision 2 — The link is card-carried: `Card.HiddenBy`, an object reference

`Card.HiddenBy` is the exiling permanent as an OBJECT, `{instance, CR 400.7 epoch}` — the same
`PermissionCardRef` shape cast permissions already use to name one object rather than one card
(`game.ObjectRefOf`). The exile route stamps it before the face-down knowers are computed
(`zoneRoute.HiddenBy`, reached through `ExileHiddenForEffect`), and `MoveCard` clears it on every
move. `HiddenCardsForEffect(ref)` is CR 607.2a's "the exiled card": the cards still in exile that
exactly that object hid.

The effects package already has an exiled-with record, `b27ExiledWith`, read off the event log, and
#1239 made it epoch-exact. We did not reuse it, for two reasons:

- The first reader of the link is the face-down viewer rule. That rule answers about one `Card`
  inside the engine, from `faceDownViewersLocked` and the state checks. It has no event log to walk,
  and it must not depend on a label.
- The link is printed on the exiled card itself ("the permanent that exiled *this card*"), so the
  card is where it belongs.

It is exact in the same way #1239 made the log record exact: the epoch pins the incarnation. A
hideaway land that is bounced and replayed is a new object. It hides a new card, and it has no claim
on the card its earlier self hid.

The object reference also means a linked ability that resolves after its permanent has left still
finds the card. Rulings on the Lorwyn lands confirm that the exile is what the ability reads, not
the permanent. An activated ability reads the reference off `StackItem.SourceEpoch`
(`effects.HiddenRefOfActivation`). A triggered one captures `game.ObjectRefOf(*source)` in its
`Build`.

## Decision 3 — A hidden spell is an ADR 0066 grant the card's ability writes; a hidden land is played on the spot

`effects.PlayHiddenCard` splits on what was hidden. A **land** is played during the resolution
(Decision 5). A **spell** gets a grant for the controller:

- `Cost "{0}"` — "without paying its mana cost".
- **Not** `CastOnly` — hideaway says *play*.
- `TimingFlash` — a spell played "as part of the resolution" ignores its own timing (CR 608.2g).
- Until end of turn — the zero `Duration`.

The engine had everything else. A face-down exiled card is cast out of exile and turned face up as it
moves (CR 406.3a — foretell's path). A grant names one object at one epoch, so a card that leaves
exile loses it.

**Declared: a grant, not an inline cast, for a spell.** It is ADR 0066's posture on every "you may
cast it" a resolution offers (cascade, Malcolm), and for the same reason: the announce path has no
frame for a cast collected from inside a resolution. What it adds over the printed card is a
response window before the cast and the choice of when in the turn to take it; it takes nothing
away (flash timing, no cost), so no card is weaker than printed.

## Decision 5 — A hidden land is played during the resolution (2026-09-23 amendment)

The first version of this ADR sent a hidden land through the grant too, which made it wait for a main
phase with the stack empty — weaker than printed, and under house rules a card cannot be `full` while
it is. So a land is played as part of the resolution (CR 608.2g), which is what the card says.

`Game.PlayLandDuringResolutionForEffect(player, card)` is a **land play**, not a "put onto the
battlefield" (CR 305.4). It runs the CR 614 window with `landPlay` set and finishes on
`executeEntryToBattlefieldLocked`, the same finisher a paused land play from hand already uses. That
finisher bumps the land-drop tally (CR 305.2a counts a land played during a resolution), clears the
face-down state, marks the table and fires the ETB, and it can pause and resume (a shockland's
payment). It needs no main phase and no empty stack: that window is CR 305.1's, for a land played
from hand with priority. Two things still stop it, both CR 305's "ignore any part of an effect that
instructs a player to do so": no land drop left (CR 305.2b), and not the player's turn (CR 305.3).
`CanPlayLandDuringResolutionForEffect` is that gate on its own, and `PlayHiddenCard` asks it first,
so nobody is offered a play the rules would ignore. When the gate says no, the land stays hidden for
a later activation.

The "you may" is the existing `may_cast` prompt (`QueueMayCastForEffect`, cascade's). An open prompt
freezes the table, so answering it is still inside the resolution, and the bot already answers it.
Cascade and Malcolm can reuse the same function for a land if a card ever needs it.

## Decision 4 — `effects.Hideaway(name, n)` is CR 702.75a whole

`Hideaway` is the permanent's ETB trigger:

1. "Look at the top N" makes the controller, and only the controller, a knower
   (`LookAtTopOfLibraryForEffect`).
2. "Exile one of them" is a mandatory `choose_cards` pick (Min 1, Max 1). There is no prompt when the
   library shows one card. The pick is already enumerated for the bot (#544).
3. The pick goes to `ExileHiddenForEffect`, which uses the shared exit primitive, so a commander
   hidden off its owner's library is still offered the command zone (CR 903.9) and carries no link
   if it goes there.
4. The rest go to the bottom in a random order through the keyed RNG. This runs in the
   continuation, after the exile, over the cards still in the library.

The trigger captures its permanent as an object in `Build`, so the card is linked to that
incarnation even if the permanent is gone before the trigger resolves (CR 603.10). The card is then
hidden with nothing able to play it, which is the printed outcome.

## Cards

| Card | Shape | Completeness |
|---|---|---|
| Rabble Rousing | Hideaway 5; one Citizen per attacker (OncePerBatch), then ten creatures opens the card | full |
| Mosswort Bridge (#294) | Hideaway 4 land; total power 10 | full |
| Windbrisk Heights (#298) | Hideaway 4 land; attacked with three creatures this turn | full |
| Spinerock Knoll (#299) | Hideaway 4 land; one opponent dealt 7 damage this turn | full |

The three lands are the ones on the card-coverage roadmap, written as one table
(`hideaway_lands.go`). Each condition is part of the effect ("you may play … if …"), so it is asked
when the ability **resolves**. Rabble Rousing's check reads the board after its Citizens land, because
it runs in the token creation's continuation. The grant-not-inline deviation is ADR-level and shared
by every card, so none of the four carries a per-card caveat.

## Out of scope

- **Evercoat Ursine** — two hideaways and "one of them". `HiddenCardsForEffect` already returns both;
  `PlayHiddenCard` grants the first.
- **Watcher for Tomorrow** — "when this creature leaves the battlefield, put the exiled card into its
  owner's hand". That leaves-the-battlefield trigger needs the departed object's epoch captured when
  it triggers, and no helper does that yet.
- The other hideaway cards (Fight Rigging, Collector's Cage, Wiretapping, Cemetery Tampering, Clive's
  Hideaway, Howltooth Hollow, Shelldock Isle, Smuggler's Buggy, Widespread Thieving). Each is its
  own condition on the same two words.

## Consequences

- ADR 0069's viewers table has a fourth row (dated amendment there).
- `Card` grows `HiddenBy`. It is carried by the snapshot and pinned by the drift test.
- The wire needs no new field. `face_down_kind: "hideaway"` is public and labels the card back
  "HIDEAWAY". `face_visible` is true for the source's controller. The play is the existing
  `exile_play` stamp.

## Test plan

- `game/hideaway_test.go` covers:
  - the face-down landing and its one viewer;
  - the object link, including a later epoch finding nothing;
  - the look following a change of control, with CR 406.3 keeping the old viewer;
  - the link dropped on leaving exile;
  - a snapshot round trip;
  - a hidden spell cast for `{0}` and a hidden land played from exile under a grant;
  - `PlayLandDuringResolutionForEffect` refused on another player's turn (CR 305.3) and a land play on its own.
- `cards/effects/hideaway_cards_test.go` covers:
  - the look-and-choose (one hidden, four on the bottom, controller-only knowers);
  - Rabble Rousing at nine and at ten creatures;
  - Windbrisk Heights refused before an attack and granted after three attackers, then cast for free;
  - Mosswort Bridge at power 9 and at 10;
  - Spinerock Knoll with 7 damage split across two opponents, then 7 to one;
  - a replayed land with no claim on its old card;
  - a hidden land played during the resolution, in combat, spending the land drop — and not offered with no land drop left (Decision 5).
- `legal/hideaway_moves_test.go` — the enumerator offers a hidden card only under its grant, only to
  the holder, and the offered play dispatches.
- `protocol/hideaway_view_test.go` — the controller reads the card, the opponent reads a back labelled
  `hideaway`.
