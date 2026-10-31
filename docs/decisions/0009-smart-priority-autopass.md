# ADR 0009 — Smart priority auto-pass (S13.6)

**Status:** Accepted · 2026-04-23 · Sprint S13.6
**Amended by:** S60, 2026-10-09 — [ADR 0143](0143-gameplay-settings-overhaul.md)
replaces `autoPassPriority`, `smartAutoPass` and `alwaysStopOpponentStack` with
one Auto-pass choice (Smart, Careful, Manual), keeps the key windows on in Smart
and Careful, and splits the stops grid by whose turn it is. Its §2.5 is the
current precedence list, and `client/src/lib/autopassDecision.ts` follows it.
The settings that tune auto-pass live under Settings → Gameplay → Passing
priority and its Advanced section (ADR 0143 §3). The autopass toggle is now
**Skip to my turn** (ADR 0143 §4.2): decision 7's safety belt is always on,
`autopassPersistThroughTurns` is gone, and the toggle keeps Smart's key windows
when the viewer can respond.
**Amended by:** S31 sub-PR 2 ([ADR 0033](0033-ai-bot-seat.md) §1, PR #429) — the
mechanism under decisions 2–4 is gone, the policy above it is not.
**Amended by:** S35 (#526) — **decision 6 is reversed on one point: the autopass
toggle no longer overrides a manual pin.** See "Amendment: a manual pin beats the
autopass toggle (#526)" below. The gate chain also moved out of `Game.svelte`
into `client/src/lib/autopassDecision.ts`, which is where the precedence list
now lives in code.
**Amended by:** S35 (#1307), 2026-09-23 — smart autopass asks about
*responses* in *key windows*, mana and land stop counting, and an opponent's
spell you cannot answer now passes. It also adds bluffing and a "considering a
response…" chip. See "Amendment: key windows and response categories (#1307)"
below.
**Amended by:** S59 ([ADR 0118](0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md)
owner decision 8, #2188), 2026-10-04 — rule 8 also holds the viewer's own
main phase when the engine may not see their mana: a spell in hand left out
for mana alone while they control a mana source the engine does not run. See
the precedence list below.
**Amended by:** S60 (#2853), 2026-10-09 — an untargeted activated ability is
no longer a response by default, and the hold toggle clears itself once the
stack it held has emptied. See "Amendment: only real interaction stops you
(#2853)" below.
**Amended by:** S60 (#2871), 2026-10-09 — "stop at a ticked step only when I
can do something" becomes its own setting, and crew, manlands and granted
combat keywords count as a response in combat windows. See "Amendment: ticked
steps and combat abilities (#2871)" below.
**Amended by:** S59 (#2881), 2026-10-09 — the dock's Pass turn no longer jumps
to the next turn. It sends `end_turn`, which passes priority for the active
player until their turn ends, so every step still happens and its triggers
fire. See "Amendment: pass turn walks the turn (#2881)" below.
**Amended by:** S72 (#3555), 2026-10-15 — owner answer 2 of the #2853
amendment is narrowed: an untargeted sacrifice outlet, protection, pump or
prevention effect stops you on an opponent's stack item only when that item
threatens you, and never at an empty-stack end step. Combat is unchanged. See
"Amendment: a free answer stops you only for a threat (#3555)" below.
**Amended again by:** S72 (#3825), 2026-10-15 — an opponent's activated
ability that wipes or edicts without a target (a planeswalker's minus,
Nevinyrral's Disk, Pestilence) is a threat too, read from its catalog row.
**Amended again by:** S72 (#3882), 2026-10-15 — an opponent's triggered
ability that sweeps or edicts (a saga chapter that destroys all creatures, a
Massacre Wurm's enters trigger, a Fleshbag Marauder's edict) is a threat
too, read from its catalog row the same way.
**Amendment proposed:** S75 (#4734), 2026-10-30, not yet accepted — a
response held all game (a free Flawless Maneuver, Ranger-Captain of Eos's
sacrifice) stops you only when one of its declared answers can answer what
is happening now: a threat to you or yours for `protect` and `survive`, an
opponent's upkeep or spell for `restrict`. A move that declares nothing stops
as before. See "Amendment of 2026-10-30 (#4734)" below.
**Input changed by:** [ADR 0142](0142-declared-answers-on-catalog-abilities.md),
2026-10-09 — `interacts` reads what the ability row declares it answers
(`Purpose.Answers`) before the printed-text read, and `combat_interacts`
carries the combat tier. The response classes here are unchanged. Since
ADR 0142 S4 (2026-10-10) the printed-text read is gone, including the combat
read in the "Combat abilities count in combat" amendment below: every catalog
row declares, and a row that declares nothing (one carried on a card
instance in practice) sets `interacts`, a false-positive stop over a false-negative skip.

`hasAnyLegalResponse` no longer walks the viewer's cards running per-action
predicates. The server enumerates the seat's legal moves and ships them as
`GameView.legal_moves`, so the aggregate question is
`legal_moves.some(m => m.kind !== "pass")` — one scan of a field the engine
already computed. Three consequences for a reader of the sections below:

- **Decision 3's conservatism is now structural rather than deliberate.** The
  client no longer has to assume every permanent is "potentially activatable"
  because it cannot see ability lists; the engine tells it exactly which
  activations are live. The false-positive-over-false-negative preference
  survives, and is what makes an ABSENT `legal_moves` (a pre-S31 server, or a
  frame where the seat owes nothing) resolve to "stop" rather than "skip".
- **Decision 4 is obsolete: mana affordability is now in scope**, for free. The
  enumerator pays for what it offers, so a hand of uncastable 7-drops correctly
  reports no response and the window is skippable.
- **Decision 2's memo cache is gone.** It existed because the card walk was
  O(hand + battlefield + command) per call; an array scan does not need it.
  `_resetCacheForTests` survives as a no-op so test teardowns don't break.

The module split decision 2 argues for (`timing.ts` per-action, `priority.ts`
aggregate) still holds and is unchanged.

## Amendment: a manual pin beats the autopass toggle (#526)

Decision 6 below says the autopass toggle "ignores every gate (stops grid,
smartAutoPass predicate, manual pins, `settings.autoPassPriority`)". The pins
half of that was wrong, and a playtester reported it as a bug: *"When autopass
is turned on, and a manual stop is placed on a step, the game continues to
autopass right through without ever stopping."*

**The precedence is now:** a manual pin (decision 5) sits **above** the autopass
toggle. Everything else in decision 6 stands — the toggle still out-votes
`settings.autoPassPriority`, the stops grid, the smartAutoPass predicate and the
#323 own-stack rule.

**Why the original reading doesn't survive contact with play.** Decision 6's
argument was "the point of autopass is 'no more asking'". But the pin is a
*later and narrower* instruction than the toggle: the toggle is standing intent
("I'm tapped out, carry me"), the pin is a click made for one step of one turn,
and the only reason to make it is to interrupt automatic passing. Under the old
order the affordance was dead in exactly the state where a player needs it —
the PhaseDisplay icon lit up and changed nothing. Decision 5's own argument
("if smartAutoPass could skip a pinned step, the affordance would be a lie")
applies with more force to the toggle, not less.

**It does not strand the toggle.** The pin is consumed on the next step
transition, so the cursor holds once and autopass resumes on its own with no
second click. The player who wants out of autopass entirely still clicks the
toggle.

**Two things deliberately kept above the pin:**

- The guards that are not questions about what the viewer *may* do — an open
  pending choice, an owed declare-blockers decision (#328), the CR 732 loop
  breaker ([ADR 0055](0055-loop-breaker.md), #628), mulligans, game over,
  elimination. All of these hold anyway, so the pin changes nothing there.
- The safety belt (decision 7). Entering the viewer's own `precombat_main`
  *disarms the toggle* instead of passing, which holds the cursor too — so a
  pin on your own main phase loses nothing by sitting under it, and gains the
  disarm. Putting the pin first would leave a forgotten toggle armed for the
  rest of that turn.

**Where the code lives.** The chain was ~90 lines inline in one `$effect` in
`Game.svelte`, which is how the bug survived four sprints: no test could reach
it. It is now `autopassDecision(gates) -> "hold" | "pass" | "clear-toggle"` in
`client/src/lib/autopassDecision.ts`, a pure function over resolved booleans,
with `autopassDecision.test.ts` covering the precedence, #526's reported
configuration and #599's declare-attackers window. The component keeps the
reactive reads and the two side effects.

## Amendment: key windows and response categories (#1307)

**Status:** Accepted · 2026-09-23 · Sprint S35

Smart autopass read the right signal — the viewer's own `legal_moves`, which
the enumerator already filters for timing, targets and mana — and then asked
the wrong questions of it:

- **Mana counted as a response.** "Anything but pass" included `mana` moves, so
  an untapped land held every ticked step. Smart autopass almost never skipped.
- **Most windows were never asked.** Unticked steps always passed, which by
  default is every step of an opponent's turn, counterspell in hand or not. An
  opponent's item on the stack always held, whether or not you could answer it.
- **The autopass toggle passed everything**, a live counterspell included.

### Responses and plays

`client/src/lib/responseWindow.ts` sorts each move (`classifyMove`):

| Move | Class |
|---|---|
| `pass`, `mana` | none — never a reason to hold |
| `land`; a `cast` / `activate` in the viewer's own main phase with an empty stack | play |
| a `cast` / `activate` with `targets_stack` | counter |
| any other `cast` | instant |
| any other `activate` | ability |
| `special_action` | special |
| `attack`, `block` | declaration |
| `choice`, `mulligan`, anything unknown | other |

`hasResponse` is "a counter, instant, ability or special move in an enabled
category". The four categories are settings (`respondCounterspells`,
`respondInstants`, `respondAbilities`, `respondSpecialActions`, schema v12),
all on by default. `hasPlay` is a response or a play, declaration or other move,
plus #328's owed block and #599's declared attack. It is asked only on ticked
steps, so a hand holding just a land still stops on your own main phase, and a
land never holds an opponent's window.

`targets_stack` is a server flag on the legal move (#1307 PR 1). A server that
does not send it gets its counterspells classed as instants, which with the
default categories changes nothing.

### Key windows

Outside the ticked steps, smart autopass now stops in the windows where a
response is worth having (`keyWindow`):

- **stackOpp** — the stack holds something the viewer did not put there, or a
  trigger is still queuing (it could be anyone's).
- **combat** — declare attackers or declare blockers with an attack declared.
- **oppEnd** — an opponent's end step.

Quiet steps (an opponent's upkeep with an empty stack, say) still pass.

### Precedence, strongest first

1. The guards: no priority, mulligans / game over / eliminated, a pending
   choice, an owed block (#328), the loop breaker (#628), no step → hold.
2. The safety belt (§7) → clear-toggle.
3. A manual pin (§5, #526) → hold.
4. The autopass toggle: stackOpp with a response → hold; stackOpp with a bluff
   armed → bluff; otherwise pass.
5. `autoPassPriority` off → hold.
6. A non-empty stack: hold-priority → hold; entirely the viewer's own → #323
   (`autoPassOwnStack` ? pass : hold); smart autopass off or
   `alwaysStopOpponentStack` on → hold; a response → hold; a bluff armed →
   bluff; otherwise pass.
7. An empty stack with smart autopass on, in a combat or oppEnd window: a
   response → hold; an instant bluff armed → bluff; otherwise fall through.
8. A ticked step: hold if `smartAutoPass ? hasPlay || engineMayMissMana : true`.
   `engineMayMissMana` is ADR 0118 owner decision 8 ("stop if the engine may
   be wrong", 2026-10-04): the viewer's own main phase, a spell in hand that
   the move list leaves out for mana alone (the Cast anyway row's non-mana
   denials say nothing), and a permanent the viewer controls that carries
   `unimplemented` and has a mana ability on the wire or is a land
   (`client/src/lib/engineMayMissMana.ts`).
9. Pass.

Rules 4, 6 and 7 name a bluff verdict. A bluff only ever replaces a pass. It
never beats a guard, a pin, a real answer or any other hold. See "Bluffing"
below.

### The behaviour change

**With smart autopass on, an opponent's spell you cannot answer now passes.**
That is the S13.6 exit criterion ("Seat 3 holds no instants → cursor
auto-passes through seat 3"), which the old rule 6 never met. A player who
wants every opponent stack item to stop — to read it, or so that stopping
gives nothing away — turns on `alwaysStopOpponentStack` (default off).
Turning smart autopass off does the same, along with its other effects.

### §3 revisited

Decision 3 preferred a false-positive stop to a false-negative skip. That still
holds where the answer is unknown: a missing `legal_moves` while holding
priority counts as a response, and an unrecognised move kind counts as a play.
What changes is that a known *non*-answer no longer holds. Mana and land moves
are known not to be responses, and the enumerator already pays for what it
offers, so the permissive stance is no longer buying safety there. It only
cost clicks. The remaining false negatives are sandbox-only verbs the
enumerator does not list (moving a card by hand). A manual pin or
`alwaysStopOpponentStack` covers those.

### Where the code lives

`responseWindow.ts` holds the questions, `autopassDecision.ts` holds the order,
and `Game.svelte` gathers the gates. `hasAnyLegalResponse` is gone.
`hasNonPassMove` (mana included) stays in `timing.ts` for UI that wants "any
move at all".

### Bluffing

A smart hold is a tell. Automatic passes take one round trip, so any pause
means "they have something". Before this change the only way to bluff was a
manual pin (§5), clicked in advance, one step at a time.

A bluff is a pause the viewer takes when they have nothing. Two settings pick
where (schema v13), and both are off by default:

- `bluffCounterspell` — represent a counter: bluff at an opponent's item on the
  stack.
- `bluffInstant` — represent an instant: bluff there and in the other key
  windows (combat, an opponent's end step).

`bluffMode` picks how:

- **timed** (default) holds for a random delay, then passes. The delay is
  uniform between `bluffDelayMinMs` and `bluffDelayMaxMs` (1500 and 4000 by
  default), clamped to 500–15000 ms (`bluff.ts`).
- **manual** holds until the player clicks next, like a real hold.

Either one also needs the in-game **bluff** button in the phase widget. That is
a session switch (`bluffArmed`) which starts on at game load when either
setting is on, so a player can stop bluffing for the rest of a game without
opening Settings. The widget shows "bluffing — passes in Ns" or "bluffing —
click next" to the viewer only.

> **Amended by ADR 0111 §5 (S56 PR 1, #1958):** the bluff button is now always
> shown, not only when a bluff setting is on. It is a split button: the main
> part arms and disarms in one click (and so does the `b` key); with neither
> kind chosen that click also turns on "represent a counterspell". Its `▾` sets
> which bluff and how, writing the same settings. It is disabled while smart
> auto-pass is off.

The timed bluff lives in `Game.svelte` and is careful about one thing: it must
never pass a window that has changed under it.

- The delay is rolled once per frame (`seq`). A re-run of the effect on the
  same frame keeps the timer.
- Any other verdict cancels it, and so do `next`, any action sent through the
  board, and leaving the game.
- When it fires, it checks the frame is unchanged, that no pass was already
  sent for it, and that the client's action count has not moved (so any send
  path counts). It then re-runs the decision, and passes only if the answer is
  still a timed bluff or a pass.

**Costs, stated:**

- **Every bluff slows the table.** Several seats bluffing timed at a four-player
  table add up. The delay range is adjustable and both bluffs are off by
  default.
- **A timed bluff has a ceiling.** It always ends inside its range, so a pause
  longer than the maximum still means a real answer. Manual mode removes the
  ceiling, at the price of a click.
- `alwaysStopOpponentStack` is the other way to give nothing away on the stack:
  stop every time, answer or not.

### "Considering a response…"

The "Opponent thinking indicator" under "Not done" is resolved, and without the
cross-seat visibility it was waiting on. The client shows a "{name} is
considering a response…" chip, on the seat and in the stack overlay, when a
human seat other than the viewer's has held priority for more than 800 ms in a
response window. A response window is one where the stack is non-empty or the
holder is not the active seat, no blocking choice is open, and nobody owes a
block decision.

It is derived entirely client-side, from public data and elapsed time
(`considering.ts`). There is no protocol change and nothing leaks. An automatic
pass clears in about one round trip, well inside 800 ms. So a real hold, a
timed bluff, a manual bluff and a player who has stepped away all look the
same, which is the point. Bot seats keep their own `botThinking` chip.

## Amendment: only real interaction stops you (#2853)

**Status:** Accepted · 2026-10-09 · Sprint S60

In a real Commander game smart autopass almost never passed an opponent's
stack item. `classifyMove` counted every affordable non-mana activated ability
as an `ability` response, and `respondAbilities` was on by default, so a Mind
Stone, a fetch land, a Clue, a Wayfarer's Bauble or a card with cycling in hand
held every opponent spell. A probe in `internal/legal` reproduced it: three
Forests and Llanowar Elves give only pass and mana moves, and adding Mind Stone
and Evolving Wilds adds two `activate` moves with nothing to say they answer
anything.

**Owner decision 1 (2026-10-09).** By default an opponent's stack item stops you
only for real interaction: an instant you can cast, a counterspell or anything
else targeting the stack, or an activated ability that targets. An untargeted
value ability no longer stops you, and a setting brings the old behaviour back.

**Owner decision 2 (2026-10-09).** The hold toggle keeps holding every stack, an
opponent's item included. Its tooltip and shortcut hint now say so, and it
clears itself once the stack empties.

**Owner answer 2 (2026-10-09, on review).** An untargeted ability that DOES
interact must still stop you by default ("Stop for these too"): a sacrifice
outlet, regeneration, protection, indestructible, hexproof or shroud, phasing,
a blink, damage prevention, a pump or counters. Special actions stay on by
default (owner answer 1; a separate issue will look at them).

**The goal (owner, 2026-10-09).** "Make the autopass as smart as possible and as
convenient as possible so most of the time players are not thinking why do I
have to click to pass or thinking I missed my window to respond." Two halves,
both held: no pointless stops, and no missed windows.

### What changed

- **`has_targets` on the legal move.** The enumerator sets it on every `cast`
  and `activate` announcement that chose at least one target, on the stack or
  anywhere else (`legal.Move.HasTargets`, `docs/protocol.md`). It is per
  announcement, like `targets_stack`, and rides in `capLegalMoves`' key
  `(source, kind, targets_stack, has_targets)`, so a capped list never merges a
  targeted mode into an untargeted one.
- **`interacts` on the legal move.** For an untargeted activation that can still
  answer the stack, the enumerator sets `interacts`
  (`server/internal/legal/interacts.go`, `docs/protocol.md`). It reads the
  ability's shape, because the effect is a closure:
  - the cost: sacrificing, exiling or returning a creature you control (a
    sacrifice outlet or a save). Sacrificing a land, a Food, a Treasure, a Clue
    or a plain artifact is a price, not an answer;
  - the declared purpose (ADR 0126 §6): a pump, a combat-damage shield, damage
    to a creature, a sweep;
  - the printed effect text after the cost: regenerate, protection,
    indestructible, hexproof, shroud, persist, undying, first strike, double
    strike, deathtouch, phasing, prevention or redirection, a power/toughness
    change or +1/+1 and -1/-1 counters (monstrosity and adapt included), a
    blink, returning itself to hand, damage to each creature, a destroy that is
    not "destroy this", and "can't cast".

  On a mana ability only a creature sacrifice outlet (Ashnod's Altar, Phyrexian
  Altar) sets it. The rules lean to "interacts" only where the text cannot say.
  `interacts_internal_test.go` pins a list of printed answers and a list of
  printed value rows. An audit of the catalog's 706 untargeted instant-speed
  activated rows marked 245 as interacting; those outside the plain categories
  (regenerate, a pump, protection, prevention) were read one by one, and the
  value rows it caught at first (charge and storage counters, damage to each
  opponent, "destroy this enchantment", sacrificing Foods, Treasures and
  lands) were moved out by narrowing the rules. `interacts` is in the cap key
  too.
- **A new class.** `classifyMove` keeps `counter` and `instant` as they were and
  splits the rest of the activations:

  | Move | Class |
  |---|---|
  | `activate` with `has_targets` or `interacts` (not `targets_stack`) | ability |
  | `mana` with `interacts`, outside the viewer's sorcery window | ability |
  | any other `activate` | untargeted |

  Cycling is an `activate` from the hand, with no target, so it is untargeted.
  Special actions keep their own `special` class and setting. A server that
  sends no `has_targets` gets every non-counter activation classed as
  untargeted.
- **A fifth "Stop for" category.** `respondUntargetedAbilities` ("Value
  abilities (Mind Stone, fetch lands, Clues, cycling)"), off by default.
  `respondAbilities` now reads "Abilities that target or protect something
  (pumps, sacrifice outlets, regeneration)". Ticking the new one restores the
  pre-#2853 meaning of "Activated abilities". The Settings page says what the
  default stops for.
- **No missed windows: a visible "wait".** While the ADR 0119 §2 stack hold
  counts down to an automatic pass on an opponent's item, the action dock
  already showed "auto-pass in 1.4 s". It now carries a **wait** button beside
  it (`L.waitToRespond`, "wait, let me respond"). One click, or the hold key
  `h`, arms the hold toggle: the verdict turns to hold, the pending pass is
  cancelled, and the hold clears itself when that stack empties. A real answer
  in hand never reaches the countdown at all: the verdict is hold, and the
  timed pass re-asks the decision when it fires and goes only on a pass.
- **The hold toggle is for one stack.** `holdPriority` still holds every
  non-empty stack (rule 6 checks it before the own-versus-opponent split, as
  before). `noteStackForHold`, called with every frame, remembers a live stack
  while the hold is on (an item in either stack representation, or a trigger
  still queuing) and turns the hold off on the first frame after it with
  nothing there. Arming it on an empty stack, before a cast, never clears it.

The precedence list above is unchanged.

### The settings migration (schema v22)

A stored blob is materialised (see v15 → v16), so a stored
`respondAbilities: true` from the all-on v12 default cannot be told from a
chosen one. A `Stop for` list that differs from that default can be: a player
who turned any of the four v12 categories off looked at the list and left
abilities as they wanted them. So, for a blob from before v22:

- all four categories still on (or not stored at all): `respondUntargetedAbilities`
  takes the new default, `false`;
- any of the four off: `respondUntargetedAbilities` follows `respondAbilities`,
  which keeps their old "any ability" meaning.

An account copy from a v21 client goes through the same migrate
(`applySyncedCopy`). From v22 on, the stored value stands, and a non-boolean
falls back to the default.

### Costs

- A player who wanted a stop to crack a fetch land in response must now tick
  "Untargeted abilities", pin the step, or arm the hold.
- `interacts` is read from printed text, so a row whose label is worded unusually
  can be misread either way. The test lists are where a misread is fixed.
- Some combat-relevant untargeted rows still do not stop you: granting flying,
  haste, trample, menace or lifelink, "can block an additional creature", and a
  land or artifact that becomes a creature until end of turn. A board full of
  them would stop on every opponent spell, so they are left out for now.

## Amendment: ticked steps and combat abilities (#2871)

**Status:** Accepted · 2026-10-09 · Sprint S60

The owner's goal, which governs every choice here, is the one #2853 states:
"make the autopass as smart as possible and as convenient as possible so most
of the time players are not thinking why do I have to click to pass or thinking
I missed my window to respond".

### Stop at a ticked step only when I can do something

Rule 8 already skipped a ticked step with nothing to do in it, but only as one
half of `smartAutoPass`, whose other half governs the opponent's stack and the
key windows. A player who turned smart auto-pass off to see every opponent spell
also lost the skip, and nothing on the Settings page said the two were joined.

- **A setting of its own.** `gameplay.stepStopsOnlyWhenCanAct`, "Only stop at
  my ticked steps when I can do something", sits under the step grid, on by
  default. Rule 8 reads it in place of `smartAutoPass`:
  `stepStopsOnlyWhenCanAct ? hasPlay || engineMayMissMana : true`.
  `smartAutoPass` keeps rules 6 and 7 (opponent stack items and the key
  windows).
- **One toggle, not one per step.** The issue asked for a per-step option. One
  global toggle is simpler, and every step wants the same answer: a ticked step
  you cannot act on is dead air, whichever step it is. A player who wants one
  step to stop every time pins it (§5) or turns the toggle off. A per-step
  column would double the grid for a choice nobody has asked to make per step.
- **What counts as something to do** is `hasPlay`, unchanged:
  - on the viewer's own main phase with an empty stack, any cast, activation or
    land the enumerator offers (a value ability and a crew included), plus
    ADR 0118 owner decision 8's spell held back for mana alone. A main phase
    with any play in it is never skipped; only one with mana moves and nothing
    else is;
  - elsewhere, a move in an enabled "Stop for" category, a declaration, a
    choice, an unknown kind, an owed block (#328) or a declared attack (#599).
- **The "Stop for" list** applies whenever either setting is on, so its
  fieldset is disabled only when both are off.

### Combat abilities count in combat

#2853's costs named a gap: some untargeted abilities change a fight without
answering a spell, so `interacts` leaves them out, and smart auto-pass passed
the window to use them. They are now a response in a combat window and nowhere
else, because a board of them would stop the viewer on every spell.

- **`combat_interacts` on the legal move**, a server bit of its own rather
  than a wider `interacts`, so the client can gate it on the window
  (`server/internal/legal/combat_interacts.go`, `docs/protocol.md`). It is set
  on an untargeted activation that `interacts` does not mark, read from the
  ability's shape:
  - a crew cost (CR 702.122), or a crew row whose cost is printed another way;
  - printed effect text that makes it a creature (a manland, an animated
    artifact), gives or takes away flying, haste, trample, menace, lifelink,
    vigilance or reach, lets a creature block an additional creature or any
    number, makes it unblockable, lures ("must be blocked"), forces attacks,
    taps or untaps the enchanted creature, doubles damage, or makes tokens
    that enter tapped and attacking.
- **`combat_defender_only`, for token makers (owner answer, 2026-10-09).** An
  activation that makes a creature token that does not enter tapped, populates
  or amasses, and grants no combat keyword, sets `combat_interacts` and
  `combat_defender_only`. A token is a blocker, so it counts as a response only
  while the viewer is a defending player in this combat (`isDefending`: some
  attacker's defending player is the viewer, which covers a planeswalker they
  control and a battle they protect). When another player is attacked, or
  before attackers are declared, it stays untargeted. Crew, manlands and the
  keyword grants are not narrowed. A token row that also grants haste
  (Sokenzan, Ingris Stingerquill) or makes attacking tokens (Dalkovan
  Encampment) is a plain combat ability.

  Both bits ride in `capLegalMoves`' key, now `(source, kind, targets_stack,
  has_targets, interacts, combat_interacts, combat_defender_only)`. The change to `interacts.go` is
  none: `abilityInteracts` is asked first, and a row it marks is never marked
  again.
- **The combat window** (`inCombatWindow`, `client/src/lib/responseWindow.ts`):
  beginning of combat, declare attackers and declare blockers, or a triggered
  or activated item on the stack whose text names an attack or a block. Combat
  damage and end of combat are too late to make a blocker. There,
  `classifyMove` classes a `combat_interacts` activation as `ability` (on by
  default); elsewhere it stays `untargeted` (off by default). So a crew or a
  manland stops you at an opponent's declare attackers or blockers, and on a
  ticked beginning of combat, and not for an opponent's main-phase sorcery. A
  token maker stops you there only when you are the one attacked.
- **The audit.** Of the catalog's 478 untargeted, instant-speed activated rows
  that `interacts` does not mark, the rules mark 58 as combat abilities (27
  crew rows, the manlands and animated artifacts, the keyword grants, the
  extra-block rows) and 52 as defender-only token makers. The rest were read
  one by one. A tapped token (Automated Assembly Line) and a non-creature token
  (Treasure, Food, Clue, Blood) do not count; "becomes that type" and "becomes
  prepared" do not either. `combat_interacts_internal_test.go` pins both lists.

### The settings migration (schema v23)

A blob from before v23 copies `stepStopsOnlyWhenCanAct` from its
`smartAutoPass`, so nobody's behaviour changes: a player who turned smart
auto-pass off still stops at every ticked step, and everyone else keeps
skipping the empty ones. An account copy from a v22 client goes through the
same migrate (`applySyncedCopy`). From v23 on, the stored value stands, and a
non-boolean falls back to the default (on).

### Costs

- A token maker counts only while you defend, so a boast token maker
  (Dragonkin Berserker, Usher of the Fallen) or Goro-Goro's Dragon, usable only
  while you attack, never stops you. Their own turn's combat is the active
  player's to stop on (a ticked step or a pin).
- `combat_interacts` is read from printed text like `interacts`, with the same
  risk of a misread either way; the test lists are where one is fixed.
- A pump written without a `+N/+N` ("Double this creature's power") is still
  read as value by `interacts`; #2872's declared answers is the place to fix
  that, not this text reader.

## Amendment: pass turn walks the turn (#2881)

**Status:** Accepted · 2026-10-09 · Sprint S59

Decision 6 calls `pass turn` an "active-player-only whole-turn skip". It was
built as one: the server jumped the cursor to the next seat's untap step, so
the steps left in the turn never began. "At the beginning of combat", "at the
beginning of your end step" and every other step trigger was skipped with them,
for every player, and so was the cleanup discard. CR 500.1 has every phase
happen every turn, and CR 500.6 fires an "at the beginning of" trigger as its
step begins.

**Now:** the dock's Pass turn sends a new action, `end_turn`
(`Game.EndTurnByPassing`, `server/internal/game/pass_turn.go`). It is a
standing instruction to pass priority for the active player every time they
would hold it, until their turn ends. The steps are walked by the ordinary
priority engine, so each one begins, its triggers go on the stack the next
time a player would receive priority (CR 117.5, 603.3), and they resolve.

- **Only the active player is passed for.** The other players still get
  priority in every step (CR 117.3d). Their own automatic passing (this ADR's
  smart autopass, or a bot runner) decides whether they stop, so nobody misses
  a window because someone else passed the turn.
- **A decision stops it.** A "may", a target, a trigger ordering, the cleanup
  discard or any other prompt the table waits for, any prompt owed by the
  passing player, the CR 732 loop notice, or a pass the rules refuse (an attack
  requirement, CR 508.1d). The player answers, and the passing resumes on its
  own: `actions.Dispatch` settles it after every action, the room after every
  automatic answer.
- **It ends with the turn**, by any route, and an undo of the `end_turn` takes
  it back. It is not part of a restore point; after a server restart the
  active player presses Pass turn again.
- **A turn passed at a table of bots is slower than the old jump**: each bot
  passes each step at its table pace.

The sandbox jump is unchanged and still reachable: `pass_turn`
(`Game.PassTurn`). [ADR 0143](0143-gameplay-settings-overhaul.md) Q5 moves it to
the ⋯ menu and relabels the dock's control "End turn"; that is its delivery PR
5, not this change.

## Amendment: a free answer stops you only for a threat (#3555)

**Status:** Accepted · 2026-10-15 · Sprint S72

Two bot arena audits (2026-10-14, 64 seat-games, every priority window put
through `autopassDecision` on default settings) found untargeted stack-tier
answers stopping their controller on almost every opponent spell and almost
never being used. A free sacrifice outlet (Viscera Seer, Ashnod's Altar)
stopped on about 120 opponent items it never acted on. Selfless Spirit
(`protect`) stopped on 18 opponent stack items and acted once, and on 8
empty-stack opponent end steps and acted on none. Purphoros (`pump`) and
Batterskull (`protect`) acted on none of theirs. In combat the same answers
were used often (Selfless Spirit: 29 stops, 13 acted). Owner answer 2 of the
#2853 amendment made these answers stop you by default, wherever an opponent's
item waits, and that is the half of the goal ("not thinking why do I have to
click to pass") it missed.

**Owner answers (2026-10-15, "Use recommendations for all").**

1. **Q1 (a): threat-gate a free sacrifice outlet.** On an opponent's stack item
   a `sac_outlet` answer counts only when that item threatens you. At an
   opponent's end step with an empty stack it never counts. Combat is
   unchanged.
2. **Q2 (a): the same gate for `protect`, `pump` and `prevent`.** `remove`,
   `restrict`, instants and flash keep stopping as they do today. `survive`
   (regenerate, indestructible, persist, undying, phasing, a blink), which
   [ADR 0142's 2026-10-15 amendment](0142-declared-answers-on-catalog-abilities.md)
   split from `protect` the same day, is gated with it: the question was
   asked about `protect` before the split, and Selfless Spirit, its example,
   now declares `survive`.
3. **Q3 (a): the server applies it** when it sets `interacts`, as it does for
   ADR 0142's owner answer 4. There is no protocol change, and the client, the
   bot and the MCP seat all get it.

**What "threatens you" means.** An item on the stack that the seat does not
control threatens it when:

- it **targets** the seat, a permanent the seat controls, or an item on the
  stack the seat controls (removal on your creature, a steal, an Aura, burn
  or an edict aimed at you, a counterspell aimed at your spell); or
- it is an **instant or sorcery spell with no target**, which could be a
  wipe, an edict or a steal.

Two readings the recommendation left open, settled here. A target on the
seat itself counts, not only on its permanents: `prevent` answers burn to the
face, and a targeted edict ("target player sacrifices a creature") is an
edict. And "a spell" means an instant or a sorcery. A permanent spell with no
target does nothing to the board while it is on the stack; its enters
trigger is an item of its own, and counts when it targets something of yours.
An untargeted triggered or activated ability is not a threat either, as the
recommendation said: the enumerator cannot tell a Massacre Wurm's enters
trigger from a draw trigger without the effect, so a wipe that arrives as a
trigger is the missed window this costs. An instant or sorcery is the
conservative reading: an opponent's Divination still stops you.

**How it is applied** (`server/internal/legal/interacts.go`):

- A row whose declared answers include `protect`, `survive`, `pump`,
  `prevent` or `sac_outlet` sets `interacts` only while the stack threatens the seat
  (`stackThreatens`). A row that also declares `remove` or `restrict` is not
  gated.
- Without a threat the row sets `combat_interacts` instead (never
  `combat_defender_only`: a pump or a shield helps an attacker too). The
  client counts that in its combat window (beginning of combat, declare
  attackers, declare blockers, an attack or block trigger on the stack), so
  combat stops for these answers as it did, and the empty-stack end step,
  an opponent's upkeep and an untargeted opponent spell do not.
- A mana sacrifice outlet (Ashnod's Altar) cannot carry `combat_interacts`, so
  the server reads the step itself: it sets `interacts` while the stack
  threatens the seat or in beginning of combat, declare attackers or declare
  blockers.
- ADR 0142's owner answer 4 is unchanged: "Sacrifice this creature: …" sets
  `interacts` while its creature is targeted by an opponent's item, and
  `combat_interacts` while it attacks or blocks, whatever the row declares.
  A row that declares nothing still sets `interacts` (ADR 0142 owner answer 3).

`threat_gate_test.go` and `TestAnswerFlagsThreatGate` (internal/legal) pin
both halves: an empty stack, an opponent's removal aimed at another player, an
opponent's creature spell and an untargeted trigger do not set `interacts` for
Viscera Seer, Undercity Troll, Evernight Shade, Spore Frog, Selfless Spirit or
Ashnod's Altar; removal on your creature, an untargeted sorcery, a spell aimed
at you, a counterspell aimed at your spell and a trigger aimed at your
creature do, and Pestilence (`remove`) interacts throughout.

**Measured.** Four seeded lockstep games (mono-black aristocrats, mono-white
weenie, esper control, boros voltron; 9,113 priority windows) replayed through
`autopassDecision` on default settings, before and after: stops on opponent
stack items fell from 96 to 87 and at opponent end steps from 15 to 13, and
none of the stops removed was one the bot acted on (27 and 4 acted on, before
and after). Combat stops were unchanged (244, 128 acted on).

**Costs.**

- A wipe or an edict that arrives as an ability with no target does not stop
  you for a free outlet, a shield or a pump when its catalog row declares
  neither a sweep nor an edict. Tick the step, pin it, arm the hold, or turn
  on "Value abilities" in "Stop for" to see it. An activated ability is read
  since #3825 and a triggered one since #3882, below.
- An opponent's untargeted instant or sorcery still stops you, a Divination
  included: without reading the spell's effect, the gate cannot tell it from a
  Wrath.

### Amendment: an untargeted wipe from an activated ability is a threat (#3825)

**Status:** Accepted · 2026-10-15 · Sprint S72

The first cut above missed the wipes that arrive as an activated ability: a
planeswalker's minus, Nevinyrral's Disk, Pestilence, Pernicious Deed, Ratchet
Bomb, Oblivion Stone. They have no target and are no instant or sorcery, and
a free sacrifice outlet, a regeneration shield or a protect effect matters
most against exactly these.

An activated ability of an item the seat does not control now also threatens
the seat when it **affects permanents the seat controls without targeting
them**. A stack item has no card type to read, so the server reads what the
catalog row it came from declares (`StackItem.AbilityPurpose`, through the
`Params.Ability` ref every catalog activation carries):

- a declared sweep (`Purpose.Sweep`: destroy, exile, bounce, damage, -N/-N or
  sacrifice) threatens when the seat controls a permanent of the class it
  matches (`sweepReaches`). `Matches` is an upper bound (`Partial`), so a
  miss is certain and a hit is a possibility. A sweep that names only the
  opponents' permanents (`OpponentsOnly`) is read the same way: the opponent
  is the ability's controller, so the seat is who it reaches. A creature
  sweep against a board with no creature does not stop you;
- else a row that declares `remove` (an edict, a sweep with no class named)
  threatens, since its candidates cannot be narrowed.

Everything else about an opponent's activated ability is as before: it
threatens by its targets, and an ability that pumps, shields or draws on its
own board does not. An activation with no catalog row (a sandbox activation,
a card carrying its own ability list) is not read. A triggered ability is
read since #3882, below.

Pernicious Deed is not in the catalog yet. When it lands, its `{X}, Sacrifice
this: Destroy each artifact, creature and enchantment with mana value X or
less` row declares the sweep and gets this with no further change.
`ability_threat_test.go` and `TestSweepReaches` (internal/legal) pin it:
Nevinyrral's Disk, Pestilence, The Eternal Wanderer's -4 and Liliana of the
Veil's -6 stop Viscera Seer, Undercity Troll, Evernight Shade, Spore Frog,
Selfless Spirit and Ashnod's Altar; an opponent's Evernight Shade pump, an
ability with no row and the seat's own Pestilence do not.

### Amendment: a sweep or an edict from a triggered ability is a threat (#3882)

**Status:** Accepted · 2026-10-15 · Sprint S72

#3825 left triggered abilities as a recorded cost: a saga chapter that
destroys all creatures, a Massacre Wurm's enters trigger, a "whenever ...
each opponent sacrifices a creature" trigger arrive untargeted and are no
instant or sorcery, so a free outlet, shield or pump did not stop you for
them.

An opponent's triggered ability now threatens the seat by the same rule as an
activated one (`abilityReaches`). A triggered stack item carries the same
`Params.Ability` ref, to the catalog triggered row, and
`StackItem.AbilityPurpose` reads that row's `Purpose`:

- a declared `Sweep` threatens when the seat controls a permanent of its class
  (`sweepReaches`, unchanged). Triggered rows already declared it for the bot:
  the saga chapters (Phyrexian Scriptures II, The Elder Dragon War I, One Ring
  to Rule Them All II, The Great Synthesis II, The Argent Etchings III),
  Toxrill, Noxious Ghoul, Balefire Dragon, Plague Boiler and others;
- an edict is declared `AnswerRemove` on the triggered row, which
  `effects.Register` now accepts on a triggered row and nowhere else off an
  activated one (the guard in ADR 0142 §2 refused it). `effects.TriggerEdict`
  wraps the row. Declared: Fleshbag Marauder, Merciless Executioner,
  Accursed Marauder, Plaguecrafter, Demon's Disciple, Sheoldred (the True
  Scriptures), Tithing Blade, and the dies edicts Butcher of Malakir, Grave
  Pact and Dictate of Erebos. Edicts whose effect is written in a body
  function of the card declare it on the row all the same: Barrensteppe
  Siege (Mardu), Witch of the Moors (both of its rows), Winter, Tormented
  Loner, Witch-king of Angmar, The Eldest Reborn (chapter I), The Grim
  Captain, Portal to Phyrexia, Shadowgrange Archfiend and Momentum Breaker.
  `remove` is a gate input only; the bot prices none of it
  (`PurposeView.Priced` ignores `Answers`);
- a permanent declares its enters effect on the card, not on the row, so an
  enters-watching row with no sweep of its own takes the sweep of its source
  card on the battlefield (Massacre Wurm, Crater Hellion). A card with a
  second enters-watching trigger would read as sweeping through it too, a stop
  over a missed window (decision 3).

A trigger that only touches their own board (a graveyard exile, damage to
each opponent), a death trigger that is not a sweep (Massacre Wurm's life
loss), the seat's own trigger, and a trigger built as an unkeyed closure (no
catalog ref) do not stop you.

**Costs.** The declaration is per row: a body function cannot be read, so a
new edict trigger stops you only once it is wrapped in `TriggerEdict`. The
sweep triggers need nothing. A declared edict stops you for the whole row, so
Winter's enters trigger stops you before it is known whether anything is
sacrificed (its "when you do" edict is a separate item with no row), and
Pharika's Spawn, whose edict needs the escape, is not declared because it
would also stop you for a cast from hand. A sweep a card declares on a mode or
an alternative cost is not read here.

`trigger_threat_test.go` (internal/legal) pins it: Fleshbag Marauder's edict,
Butcher of Malakir's dies edict, Phyrexian Scriptures II, Toxrill's slime
counters and Massacre Wurm's enters trigger stop Viscera Seer, Undercity
Troll, Evernight Shade, Spore Frog, Selfless Spirit and Ashnod's Altar;
Phyrexian Scriptures III, Massacre Wurm's life loss, Purphoros's damage and the
seat's own Fleshbag Marauder do not. The same file pins Barrensteppe Siege's
Mardu edict, Witch of the Moors' and Winter's enters trigger as stops, and
Barrensteppe Siege's Abzan counters and the seat's own Witch of the Moors as
not.

## Amendment: a value instant does not stop you on an opponent's spell (#2877)

**Status:** Accepted · 2026-10-30 · Sprint S78

Since #2853 an untargeted value ability (Mind Stone, cycling) does not stop
you on an opponent's stack item, but every castable instant still did,
however little it had to say. ADR 0142's Q5 kept that question apart. A
2026-10-14 bot arena replay (eight seeded four-seat games, 32 seat-games,
every priority window through `autopassDecision` on default settings)
counted about 75 stops whose only responses were value instants: Altar's
Reap 16 on opponent stack items and 4 in combat, none acted on; Entomb 10 on
the stack (none acted on) and 6 at an opponent's end step (2 acted on);
Frantic Search 8 on the stack (none) and 11 in combat or at an end step (6);
Dig Through Time 19, none. Real answers (Pongify, Counterspell, Swan Song,
Absorb, Force of Negation) were acted on far more often.

**Owner decision (2026-10-30): option (a).** A castable "value" instant
(Opt, Brainstorm) does not stop smart auto-pass on an opponent's stack item
by default, and a setting turns the stop back on. The combat and
opponent's-end-step stops stay: those are rule 7's windows, and they are
where a value instant is cast.

### What changed

- **The server says which casts answer nothing.** A cast move whose
  announcement is declared exactly `value` (ADR 0142's
  [2026-10-30 amendment](0142-declared-answers-on-catalog-abilities.md#amendment-2026-10-30--spells-declare-what-they-answer-2877--accepted))
  carries `value_only`. It never does on a cast that targets the stack, that
  gives up a creature as a cost (Altar's Reap is a sacrifice outlet), or
  that declares nothing; an undeclared instant stops you as before.
- **A new class.** `classifyMove` reads a `value_only` cast as
  `valueInstant` while the stack holds something, and as `instant` with the
  stack empty:

  | Move | Class |
  |---|---|
  | `cast` with `value_only`, stack not empty (not `targets_stack`, not the sorcery window) | valueInstant |
  | `cast` with `value_only`, stack empty | instant |

  The stack question is only asked with an opponent's item on it (a stack of
  your own is #323's, rule 6), and that item is what a value instant has no
  answer to. After it resolves you get priority again with the stack empty,
  and in combat or at an opponent's end step rule 7 holds for the instant.
- **A sixth "What counts as a response" category.**
  `gameplay.respondValueInstants`, labelled **Value instants (Opt,
  Brainstorm) on an opponent's spell** (`L.respondValueInstantsToggle`), off
  by default, under **Instants and flash** in Settings → Gameplay →
  Advanced. It narrows that category: a `valueInstant` counts only when both
  are on, and the checkbox is disabled while instants are off. The list's
  help adds: "A value instant, like Opt or Brainstorm, counts in combat and
  at an opponent's end step, and on an opponent's spell only when you tick
  it above."
- **The pass says why.** When rule 6 passes an opponent's item and the
  viewer held only a value instant, the dock's chip reads "Passed for you:
  only a value instant to respond with" (or "stack hold ended, only a value
  instant to respond with"), with the chip's Undo as usual
  (`passReason.ts`, gate `heldValueInstant`, which the decision never
  reads).
- **The MCP seat's smart pass** (`mcpseat/smart.go`) mirrors the default:
  a `value_only` cast is no response while the stack holds something.

The precedence list is unchanged. Careful stops at every opponent item
anyway, Manual holds everywhere, and Skip to my turn passes everything.

### No missed window

Only the client's automatic pass changes: priority is given exactly as
before (CR 117.3d), and the cast is legal and offered as before. A player who
wants Opt in response to a particular spell has, in that window, the stack
hold's countdown with its **wait** button (ADR 0119 §2, the table's pace),
the hold key `h`, a phase pin, or the pass chip's Undo; for every spell,
the setting. In combat and at an opponent's end step nothing changes.

### The settings migration (schema v31)

A blob from before v31 takes `respondValueInstants` from
`respondUntargetedAbilities`: a player who ticked "Value abilities" (off by
default since v22) asked to be stopped for value plays, so value instants
keep stopping them. Everyone else gets the new default, off. An account
copy from a v30 client goes through the same migrate (`applySyncedCopy`).
From v31 on, the stored value stands, and a non-boolean falls back to the
default.

### Costs

- A value instant cast *in response* for its own sake (Brainstorm to hide
  cards from a discard spell, a cantrip to dig for an answer to the spell on
  the stack) no longer stops you by default. Tick the setting, or use wait,
  hold or a pin.
- Until ADR 0142's spell sweep reaches it, an instant that answers nothing
  but declares nothing still stops you (about 946 instants and 151 flash
  cards on 2026-10-30). That is a pointless stop, not a missed window.
- A spell's other declared answers are not tiered or threat-gated: Altar's
  Reap (`sac_outlet`) still stops you on every opponent spell, as #3555's Q2
  kept for instants. Gating them would be a decision of its own.

## Amendment of 2026-10-30 (#4734): a held answer stops you only when it can answer now

**Status:** Proposed · 2026-10-30 · Sprint S75 — Agents at the table: a
current seat, quieter turns, sharper bots. Issue:
[#4734](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4734). Related:
#3603 and #4186 (the MCP seat's `pass_until: smart`), #3555 and #2877 above.
No code changes with it. The owner's decision is recorded below. The
[open questions](#open-questions-for-the-owner-4734) are not answered yet,
and where the text depends on one it follows the recommended option and says
so, as "(Q1)" to "(Q5)".

The owner's standing goal for auto-pass: no pointless pass clicks, no missed
response windows, and a countdown that is visible and stoppable.

### Why

An MCP seat (Claude, Opus) played table `3376a15c` on cmd.labxp.io on
2026-10-30 with `pass_until: "smart"` throughout. It played the mono-white
"Strength in Numbers" (Adeline). Of the first 43 decisions it was stopped
at (to turn 15), 23 offered nothing but moves it held all game: Flawless Maneuver,
free once Adeline was out; Ranger-Captain of Eos's "Sacrifice this creature:
Your opponents can't cast noncreature spells this turn"; mana; pass. Fifteen
of the 23 were opponent windows: four declare-attackers steps with a trigger
on the stack, three precombat main phases with something on the stack, three
empty end steps, two empty declare-attackers steps, two empty
declare-blockers steps and one end step with something on the stack. The
other eight were the seat's own main phases, which stop by design.

Both moves are a response at every window under the rules above, in the
seat and in the browser alike (checked in `classifyMove` and
`mcpseat/smart.go`, not only in the docs):

- **Flawless Maneuver** declares nothing (`flawless_maneuver.go`), so its
  cast is an `instant` everywhere. #2877 narrowed only a cast declared
  exactly `value`.
- **Ranger-Captain's row** declares `restrict`. #3555 left `restrict` out of
  its threat gate, so the row sets `interacts` on every opponent item and
  counts as an `ability` in every key window.

### Owner decision (2026-10-30)

Option 1 of the issue. Smart auto-pass stops for a response held all game
only when it can answer what is happening, matched through
[ADR 0142](0142-declared-answers-on-catalog-abilities.md)'s declared
answers:

- **protect or survive:** a stack item or a combat threatens something of
  yours, targeted or not;
- **restrict:** at the start of an opponent's turn, or while an opponent's
  spell is on the stack;
- **counter:** an opponent's item is on the stack.

A move with no declared answer still stops, so no response window is
missed. The browser and the MCP seat share the rule, and the countdown and
its stop are unchanged. Not chosen: option 2 (the player marks a held card
"don't stop for this") and option 3 (change sacrifice outlets only).

### The rule

A move's declared answers are ADR 0142's: an activated row's
`Purpose.Answers`, a mana ability's `Answers`, and for a cast the card's,
its chosen modes' or its alternative cost's (`castAnswersOf`). For every
move that declares something other than `value` alone, the server says
whether any one of its answers matches the window the seat is in. A move
that declares several kinds matches when one of them does, the way a modal
card declares the union (ADR 0142 sweep ruling 4).

| Kind | Matches when | From |
|---|---|---|
| `protect`, `survive` | an opponent's item on the stack threatens you or yours, or a combat involves you (both below) | owner decision |
| `restrict` | it is an opponent's upkeep, or an opponent's spell is on the stack | owner decision |
| a counter (`targets_stack`) | an opponent's item is on the stack | owner decision; already true, unchanged |
| `pump`, `prevent`, `sac_outlet` | as `protect` and `survive` | Q1 |
| `combat_grant`, `animate`, `makes_blocker` | a combat involves you (the table below) | Q1 |
| `remove` | today's windows: an opponent's item on the stack, a declared attack, an opponent's end step | Q1 |
| `graveyard` (a new kind) | an opponent's spell is on the stack, or an opponent's item targets a card in a graveyard | Q5 |
| `value` | unchanged: #2877 | — |

In the viewer's own main phase with an empty stack every cast and
activation is a play (`classifyMove`'s sorcery window), whatever it matches.
That is unchanged, so the seat's own main phases still stop.

**A threat on the stack.** The saving kinds are `protect`, `survive`, `pump`,
`prevent` and `sac_outlet`: each keeps or trades away something of yours.
An item on the stack that an opponent controls threatens you when:

- it targets a permanent you control: removal, a steal, an Aura, a fight,
  damage. This counts for every saving kind. Without the effect the server
  cannot tell a Murder from a Control Magic, so it does not try;
- it targets you (burn, a targeted edict, a "target player" effect). This
  counts for `protect`, `prevent` and `sac_outlet`. `survive` and `pump`
  cannot answer it (Q4);
- it targets a spell or ability you control (a counterspell). This counts
  for `protect` ("can't be countered") only (Q4);
- it is an instant or sorcery spell with no target (CR 115.10: a spell
  affects what it does not target). That holds unless its declaration says
  it cannot reach you: it is declared exactly `value`, or it declares a
  `Sweep` whose class holds nothing you control (`sweepReaches`). An
  undeclared untargeted spell threatens, as under #3555;
- it is an activated or triggered ability whose catalog row declares a sweep
  that reaches you, or an edict (`abilityReaches`, #3825 and #3882,
  unchanged).

A permanent spell with no target does not threaten. Its enters trigger is an
item of its own, put on the stack the next time a player would receive
priority (CR 603.3), and judged then. Your own items never threaten you.

**A combat that involves you.** "You or yours is attacked" means a creature
attacks you, a planeswalker you control or a battle you protect (CR 506.3,
310.9). Commander tables use CR 802.2, where every opponent of the attacking
player is a defending player for the whole combat, so the rule asks whether
you are attacked, not whether you are a defending player. The steps:

| Kind | Declare attackers step (after the declaration, CR 508.2) | Declare blockers step (CR 509.2) and a first-strike damage step (CR 510.4) |
|---|---|---|
| `protect`, `survive`, `sac_outlet` | you or yours is attacked, or you attacked | a creature you control is blocking or blocked (CR 509.1h); for `protect`, also an unblocked creature attacking you or yours (Q4) |
| `pump` | the same | a creature you control is blocking or attacking, blocked or not |
| `prevent` | the same | a creature you control is in combat, or an unblocked creature attacks you or yours |
| `combat_grant` (Q1) | the same | a creature you control is in combat |
| `animate`, `makes_blocker` (Q1) | you or yours is attacked (today's `isDefending`) | never: blockers are already declared (CR 509.1) |

Beginning of combat matches only `animate` and `combat_grant`, and only on
your own turn: it is the last moment to crew or animate an attacker, or to
give one a keyword before attacks (CR 507.2). Otherwise it matches nothing,
because nothing is attacking yet. The regular combat damage step and end of
combat match nothing: the damage is dealt and state-based actions are
checked before anyone gets priority (CR 117.5, 704.5g). Under Q4 (b) the
saving kinds' rows become one: in any of the three steps, you or yours is
attacked, or a creature you control is attacking or blocking.

**The start of an opponent's turn** is their upkeep step. No player gets
priority in the untap step (CR 502.4), and the upkeep is the first step in
which anyone does (CR 503.1, 117.3a). A restriction made then lasts through
the main phases where they cast at sorcery speed (CR 117.1a, 611.2a). It
matches throughout their upkeep, with an upkeep trigger on the stack or
without one.

**An opponent could cast** at every window: an opponent may cast an instant
whenever they hold priority (CR 117.1a). So a restriction is not a question
of whether they can. A "can't" beats a "can" (CR 101.2) from the moment it
resolves, and it does nothing to what is already on the stack. It matters
most at the decision's two moments:

- before an opponent's turn reaches its main phases;
- while one of their spells is on the stack. They get priority again after
  it (CR 117.3b, 117.3c), to protect it or to fight over it.

"An opponent's spell" is any spell on the stack an opponent controls, a copy
included (CR 707.10). An opponent's ability on the stack does not match
`restrict`. That is the owner's line. Q5 asks whether graveyard exile should
be the exception, since it does answer a reanimation trigger.

**Counters.** `targets_stack` is read from the announcement; it is not a
declared kind. The enumerator offers such a move only when a legal target is
on the stack (CR 601.2c). Smart auto-pass asks about it only with an
opponent's item there, since a stack that is all your own is #323's. The
decision's rule is what already happens, so nothing changes for counters.

**What keeps today's class.** These moves carry no verdict and are classed
exactly as today, so they stop in the same windows (decision 3, ADR 0142
owner answer 3):

- a cast that declares nothing: 1,133 of the catalog's 1,144 instant-speed
  cards (below), Flawless Maneuver among them;
- a row carried on a card instance, and a sorcery-speed row;
- a targeted activated row, which ADR 0142 §1 does not let declare (Q2);
- a move that declares `remove` (Q1);
- a move whose only declared answer is `value`. #2877's `value_only` covers
  it.

**Owner answer 4 is a match.** A "Sacrifice this creature: …" move matches
while its creature is a target of an opponent's item, or is attacking or
blocking, whatever the row declares (ADR 0142 owner answer 4). So
Ranger-Captain still stops you when it is the one being removed.

### The rules basis

The numbers here were checked against `MagicCompRules 20260925.txt`, the
edition AGENTS.md §6 pins.

- **Auto-pass removes no window.** It is a shortcut (CR 732.1) that passes
  priority for the player (CR 117.3d). The server still gives priority at the
  start of every step that has it, after every item resolves, and after every
  cast or activation (CR 117.3a to 117.3c). A step ends only when everyone
  passes with the stack empty (CR 500.2), and every item resolves only after
  everyone passes (CR 117.4). This rule changes only when the client or the
  seat passes on the player's behalf, never what the player may do.
- **Each new item is a new window.** After an item resolves, or a trigger is
  put on the stack (CR 117.5, 603.3), the question is asked again. So an
  opponent's creature spell that passes is followed by its enters trigger,
  judged on its own.
- **What each kind answers** is the CR's: protection and hexproof stop
  targeting (CR 702.16b, 702.11b), indestructible stops destruction and
  lethal damage but nothing else (CR 702.12b), and an illegal target makes an
  item do nothing to it (CR 608.2b).
- **The turn and combat steps** are the 500 series: the untap step, the
  upkeep, combat's five steps and the extra first-strike damage step
  (CR 502.4, 503.1, 506.1, 508.2, 509.2, 510.4), cited where each is used
  above.

### Which window owns the stop

| What matches | The window | Precedence |
|---|---|---|
| a threat on the stack; `restrict` on an opponent's spell; a counter | an opponent's item on the stack (`stackOpp`) | rule 6: a response → hold. Careful holds there anyway |
| `restrict` in an opponent's upkeep with a trigger on the stack | `stackOpp` | rule 6 |
| `restrict` in an opponent's upkeep, stack empty | no key window today | rule 7's new clause |
| a combat that involves you, in declare attackers or declare blockers | the combat key window | rule 7, as today |
| a combat that involves you, in the first-strike damage step | no key window today | rule 7's new clause |
| anything in your own main phase, stack empty | your sorcery window | rule 8 (a play), unchanged |
| a ticked step | that step | rule 8: `hasPlay` counts a matched answer, not an unmatched one |

**What changes in rules 6 to 8.** `hasResponse` and `hasPlay` read the
verdict first:

- an unmatched move is neither a response nor a play, outside the viewer's
  sorcery window;
- a matched move is a response in any window: a cast is an `instant`, and an
  activation or a mana move is an `ability`. It still needs that "What counts
  as a response" category to be on;
- a move with no verdict is classed as today.

Rule 7 gains one clause after its combat and end-step test: with the stack
empty, in any other window, a matched move in an enabled category → hold.
Its bluffs keep rule 7's combat and end-step windows, and nothing bluffs at
the new clause's windows. The precedence order does not change. The browser
slice adds the clause to ADR 0143 §2.5's list and to the `autopassDecision.ts`
header.

**The countdown.** When nothing matches an opponent's item, the verdict is
pass, and the ADR 0119 §2 stack hold counts down at the table's pace, with
its **wait** button and the `h` key, as today. An empty-stack window that
nothing matches passes at once, as a quiet step does today. The dock's chip
then says "Passed for you: nothing you hold answers this", with its Undo
(`passReason.ts`, a new gate `heldUnmatchedAnswer` that the decision never
reads, like #2877's `heldValueInstant`).

### One rule for the browser and the seat (Q3)

The match is computed once, by the server, for the seat whose moves are
being enumerated. It lives in `server/internal/legal`, in a new
`answer_window.go` beside `interacts.go` and `cast_answers.go`, and it ships
on each move:

- **`answers_now`** on the legal move (`Move.AnswersNow *bool`, omitted when
  nil): `true` when a declared answer matches now, `false` when none does,
  absent when the move keeps today's class (the list above). It is additive.
  It joins the wire cap's key (`capLegalMoves`), and goes into
  `docs/protocol.md` and `protocol.ts`.
- **The old flags stay as they are.** `interacts`, `combat_interacts` and
  `combat_defender_only` are still set as today, so a reader that does not
  know `answers_now` behaves as before: an older seat binary, or the bot's
  Layer A (`aiseat/rules/value_only.go`). A reader that knows the new field
  reads it first and ignores those three on that move.
- **Both readers.** The browser reads it in `classifyMove`. The seat reads
  it in `moveIsResponse`, and adds rule 7's new clause to
  `smartPassApplies`. Neither one works out a threat, a combat or a window
  from the view.

The server computes it for three reasons:

1. The threat reads what the view does not carry in full: the catalog row
   behind an opponent's stack item (`StackItem.AbilityPurpose`), each
   announcement's targets and costs, and who controls each target.
2. #3555 Q3 and #2877 already put this kind of gate on the server, so the
   client, the bot and the seat get one answer.
3. A second copy of the rule is the drift that #3603 and #4186 had to chase.

The rest of smart auto-pass stays where it is. The windows and the
precedence are in `autopassDecision.ts`, and the seat's port of them at the
default settings is in `smart.go`. An agreement fixture keeps those two
from drifting, as `timing_agreement.json` does for cast timing
(`internal/legal/agreement_test.go`):

- a Go test in `internal/legal` builds real games and writes their frames,
  each with its legal moves and a verdict declared by hand, to
  `server/internal/legal/testdata/smart_pass_agreement.json`. It lives there
  because `mcpseat` may not import `internal/game` (its `imports_test.go`);
- `smart_test.go` asserts the seat's verdict on each frame;
- `client/src/lib/smartPassAgreement.test.ts` asserts `autopassDecision`'s
  verdict at the default settings.

Either side drifting fails a test, and neither is the other's oracle.

### Coverage of declared answers, and a wrong one

Measured on 2026-10-30 at `develop` 24a00c8da. The spells were counted on
the 2026-10-14 Scryfall dump.

| Slot | Declared | Measured by |
|---|---|---|
| untargeted instant-speed activated rows, and mana abilities with a sacrifice cost (ADR 0142 §4's scope) | 1,135 of 1,135 | `TestEveryAnswersRowDeclares`, which fails CI on an undeclared row |
| targeted instant-speed activated rows | 0 of 638; ADR 0142 §1 does not allow it | nothing yet (Q2) |
| instant-speed spells (instants and flash cards) | 11 of 1,144: nine `value`, plus Altar's Reap and Deadly Dispute (`sac_outlet`) | `go test ./internal/decks/ -run RealDumpSpellAnswers -v`, with `CMDCTRL_SCRYFALL_DUMP` |

By kind, the 1,135 rows declare `value` 520, `pump` 214, `makes_blocker` 79,
`animate` 78, `sac_outlet` 75, `survive` 60, `prevent` 56, `combat_grant`
54, `remove` 33, `restrict` 25 and `protect` 9. A row can declare more than
one.

**The gap is on spells.** Until they declare, 1,133 instant-speed cards are
classed as today. Flawless Maneuver is one of them, so the rule alone does
not quiet it. On activated rows the rule applies at once, Ranger-Captain's
included. The curated decks hold 56 undeclared instant-speed cards: Flawless
Maneuver, Teferi's Protection, Akroma's Will, Boros Charm, and the removal
and counters around them. Slice 5 declares those first. Until a card
declares, the fallback costs a click, never a window.

**A wrong declaration now costs more.** Before this, a spell's declaration
changed nothing unless it was `value`. Now a declaration that is too narrow
can pass a window the card could answer. Too broad is a pointless stop, as
before. These limit the damage:

- the union rule: a modal or mixed card declares every answer any mode can
  give (sweep ruling 4);
- the registration guard's cost rule: a cost that gives up a creature must
  declare `sac_outlet`, `protect` or `survive`, and `castValueOnly` reads a
  cast's sacrificed creatures from the announcement itself;
- owner answer 4's state read, above;
- the countdown with **wait**, the hold key, a pin, Careful, and the chip's
  Undo. The chip names the reason, so a bug report shows the miss;
- a fix is one line on one card, with a test.

A sweep PR that declares a spell anything but `value` lists each card and
its kind under "Verdicts changed", and merges on green, as ADR 0142 Q6 has
it. The build slice makes the listing count each kind, and every sweep PR
quotes the counts before and after. A CI ratchet for spells waits for the
type line in the checked-in oracle fixture, as ADR 0142's #2877 amendment
says.

### Measured: the #4734 game, replayed

The seat session's transcript holds the board text of every decision. The
replay ran each opponent window that offered only those standing moves
through the rule, with Flawless Maneuver declared `survive` (slice 5). It
reads the board text rather than re-running the engine, so it is
approximate. Repeated frames of one window count once.

| | Opponent windows today | Stop under Q4 (a) | Stop under Q4 (b) |
|---|---:|---:|---:|
| the issue's 15, to turn 15 | 15 | 4 | 5 |
| the whole session, to turn 31 | 68 | 18 | 23 |

What still stops, under (a):

- **Flawless Maneuver**, six times: four in the declare-attackers steps of
  two attacks on the seat, and two while Smaug the Magnificent's trigger was
  aimed at Mother of Runes. The issue named that last stop as the useful one.
- **Ranger-Captain**, twelve times, each with an opponent's spell on the
  stack: Boromir, Wolfbriar Elemental, Tireless Tracker, Smaug the
  Impenetrable, Generous Gift, Eternal Witness, Ghalta and others.

(b) adds two Palantír of Orthanc triggers that target the seat, and three
combat steps where an unblocked flyer or double striker hit the seat with
no block declared.

The rule also adds a stop the seat never had: each opponent's upkeep while
Ranger-Captain is on the battlefield. That was 11 upkeeps in this game, one
of them by turn 15. So in total, 68 stops become 29 under (a) and 34 under
(b). Ranger-Captain alone still stops about twice per opponent turn, which is
the decided rule for `restrict`. Every one of those stops is a window where a
restriction changes what that opponent may do for the rest of the turn.

### Calls made here

Each is overturnable in review.

1. **No history.** The rule reads a move's declared answers and the window.
   It does not track how long a card has been held, so a card drawn this turn
   and one held since turn 3 are treated alike.
2. **An opponent's upkeep, not their draw step.** One stop per opponent turn
   is enough to restrict before their main phases.
3. **An untargeted spell that declares it cannot reach you is no threat:**
   one declared exactly `value`, or a declared sweep whose class holds
   nothing of yours. An undeclared one still threatens, as #3555 has it.
4. **Counters are unchanged** (above).
5. **Smart and Careful alike.** In an empty-stack window the match applies in
   both modes, since rule 7 runs in both (ADR 0143 §2.2). Careful still holds
   every opponent item (rule 6), Manual holds everywhere, and Skip to my turn
   passes everything.
6. **Ticked steps read the match.** With "only stop when I can do something"
   on, an unmatched answer does not hold a ticked step. To stop there every
   time, turn that setting off or pin the step.
7. **No new setting.** Careful, **wait**, the hold key, pins and the chip's
   Undo are the ways to stop anyway. The seat keeps `pass_until: none` and
   `my_turn_or_stack`.
8. **`value` is unchanged.** #2877's windows stand: a value instant still
   stops you in combat and at an opponent's end step.
9. **The old move flags keep their meaning**, so the change is additive on
   the wire. Nothing new is captured, so there is no snapshot change, and
   `answers_now` holds no closure, so `closure_fields.txt` does not change.

### Costs

- `restrict` still stops often: on every opponent spell, and now also at
  every opponent's upkeep. That is three upkeep stops a round at a full
  table. The turn-off-protection rows (Detection Tower, Arcane Lighthouse,
  Shadowspear, the Glittering pair) are used before a targeted play of your
  own, so their stops at an opponent's upkeep and spells are pointless ones,
  never a missed window.
- A `protect` or `survive` answer no longer stops for a combat between other
  players, or at an empty end step. Casting one there for its own sake is
  rare. Pin the step to do it.
- Spells gain nothing until they declare (above).
- At the Fast table pace the stack hold is 0 s (ADR 0143 §2.6), so an
  unmatched opponent item passes at once. Undo and Careful remain.
- `remove` keeps today's windows (Q1), so Pestilence with {B} open still
  stops you on every opponent item.

### Tests that prove it

**Server** (`internal/legal`, slice 2). Real games, as `threat_gate_test.go`
builds them:

- `TestAnswersNowByKind`, one table of kind against window:
  - an opponent's Murder on your creature, and on another player's;
  - a declared creature sweep, with and without a creature of yours;
  - an opponent's declared-`value` sorcery, and an undeclared one;
  - an opponent's creature spell, Fleshbag Marauder's edict and Pestilence's
    sweep;
  - your own item;
  - an attack on you, on your planeswalker, on a battle you protect, and on
    another player;
  - your blocker, and your attacker blocked and unblocked;
  - beginning of combat, the first-strike damage step and the regular one;
  - an opponent's upkeep, empty and with a trigger, their draw step and their
    end step;
  - an opponent's spell against an opponent's ability, for `restrict`.
- `TestAnswersNowAbsentKeepsTodaysClass`: an undeclared cast, a run-time
  row, a targeted row, a `remove` row and a `value` cast carry no verdict.
  Over the catalog's in-scope rows, their `interacts`, `combat_interacts`
  and `value_only` stay identical before and after.
- `TestAnswersNowSelfSacrifice`: Ranger-Captain matches at any step while an
  opponent's Swords to Plowshares targets it.
- `TestAnswersNowUnion`: a row declaring `survive` and `restrict` matches in
  either one's window.
- **The #4734 regression**, `TestHeldAnswersTable3376a15c`. It rebuilds the
  game's standing board: Adeline on the battlefield; Flawless Maneuver in
  hand with its free cost; Ranger-Captain of Eos and Mother of Runes on the
  battlefield; Plains, Mind Stone and Arcane Signet. It walks the window
  shapes the transcript holds:
  - an opponent's creature spell (Smaug the Magnificent, Ghalta);
  - Smaug's trigger aimed at another player during an attack on the seat;
  - Goreclaw's untargeted trigger during an attack on another player;
  - Smaug's trigger aimed at Mother of Runes;
  - Palantír of Orthanc's trigger aimed at the seat;
  - Generous Gift on another player's Ghalta;
  - an empty opponent end step;
  - empty declare-attackers and declare-blockers steps, in an attack on
    another player and in one on the seat;
  - an opponent's upkeep;
  - the seat's own main phase.

  Each case asserts both moves' `answers_now`. Its frames join the agreement
  fixture, so slices 3 and 4 assert the browser's and the seat's verdicts on
  them too. Back-out: with the readers removed, every opponent window stops,
  as it did on 2026-10-30.

**Browser** (slice 3):

- `responseWindow.test.ts`: true, false and absent, for each class and
  window. The categories still filter, and the sorcery window is still a
  play.
- `autopassDecision.test.ts`: rule 7's new clause at an opponent's upkeep
  and in the first-strike damage step, with no bluff there. Careful, Manual,
  Skip to my turn and the stack hold's pass are unchanged.
- `passReason.test.ts`: the new reason.
- `smartPassAgreement.test.ts`: the shared fixture.
- A branch E2E, because this changes when a player is stopped.

**Seat** (slice 4): new `smart_test.go` cases and the shared fixture.

**Measured** (slice 8): the arena replay, as #3555 ran it. Eight seeded
four-seat games on the curated decks, every priority window through
`autopassDecision` at the default settings, before and after. It reports
stops by window and how many of them the bot acted on. Acceptance: no stop
the change removes is one the bot acted on.

### Slices

| # | Slice | What | Model | Needs |
|---|---|---|---|---|
| 1 | This amendment | docs | — | — |
| 2 | **The server's match** | `answer_window.go`: `answers_now` on casts, activations and mana moves; a combat that involves you; an opponent's upkeep; an opponent's spell; the stack threat as narrowed above; owner answer 4. `Move.AnswersNow`, the cap key, `docs/protocol.md`, `protocol.ts`. The per-kind counts in the spell listing. The server tests, the #4734 regression and the agreement fixture's writer. No reader yet, so no player sees a change | Opus (tier 3) | 1, with Q1, Q3 and Q4 answered |
| 3 | **The browser** | `classifyMove` and `hasPlay` read the verdict; rule 7's new clause; the pass reason; one sentence of help under "What counts as a response"; ADR 0143 §2.5 and the `autopassDecision.ts` header; the agreement test. Branch E2E | Sonnet (tier 2) | 2 |
| 4 | **The seat** | `moveIsResponse` and `smartPassApplies`; the agreement fixture read in `smart_test.go`; `docs/mcp-seat.md` and a pointer in ADR 0122 | Sonnet (tier 2) | 2 |
| 5 | **Declarations, first batch** | Flawless Maneuver and the curated decks' other 55 undeclared instant-speed cards, each from its effect with ADR 0142's table and rulings, quoting the listing's counts before and after. Inert until slice 2, so it can land first | Sonnet (tier 1) | — |
| 6 | **The spell sweep** | ADR 0142's #2877 sweep of the remaining ~1,077 instant-speed cards, in batches of about 80, protective kinds first | Sonnet (tier 1), about 14 PRs | 5 |
| 7 | **Targeted rows declare** (Q2 (a)) | the guard accepts `Answers` on a targeted activated row without requiring it; declare the protective ones first (Mother of Runes, Arwen, Liliana the Faultless, Plaza of Heroes, Skrelv, Zack Fair), then targeted pumps and prevents | Opus for the guard, Sonnet for the batches | 2 |
| 8 | **Measured** | the arena replay above; optionally, the bot's Layer A reads `answers_now` behind a `Config` switch, measured by ADR 0052's report | Opus (tier 3) | 3, 5 |
| 9 | **`graveyard` splits from `restrict`** (Q5 (a)) | a new kind after `restrict` on the wire, as `survive` split from `protect`: the guard, the boardtext label, and the retag of the eight graveyard-exile rows (Soul-Guide Lantern, Scavenger Grounds, Silent Gravestone, Stonespeaker Crystal, and `ExileFromASingleGraveyardAbility`'s Famished Ghoul, Rag Dealer, Unlicensed Hearse and Carrion Beetles) | Sonnet (tier 2) | 2 |

Slices 3 and 4 can run in parallel, and so can 5, 7 and 9. A player sees
the change only with slice 3; the seat, with slice 4. Ranger-Captain is
quiet from slice 3, and Flawless Maneuver from slices 3 and 5.

### Back-out

Reading the verdict is the whole change for a player and for the seat. Drop
it in `classifyMove` and `moveIsResponse`, along with rule 7's clause, and
every move is classed as today. The server field is additive and can stay.

### Open questions for the owner (#4734)

The decision does not settle these. Each lists the recommended option first,
the most rules-faithful one, and the text above follows it.

1. **Q1. The kinds the decision does not name:** `pump`, `prevent`,
   `sac_outlet`, `remove` and the three combat kinds.
   - **(a) Recommended:** match each by what it can change.
     - `pump`, `prevent` and `sac_outlet` join `protect` and `survive`. Each
       saves or trades away something of yours, as #3555 already groups them.
       #3555 left combat alone for them, and this narrows that: they count
       in a combat only when it involves you.
     - The combat kinds count only in a combat that involves you, and only in
       steps where they can still change it: no new blocker after blocks are
       declared.
     - `remove` keeps today's windows. It acts on an opponent's board, so
       any opponent item, a declared attack or an opponent's end step can be
       the moment for it.
   - **(b)** Change only the three kinds the decision names. `pump`, `prevent`
     and `sac_outlet` keep #3555's gate (a stack threat, or any combat
     window). `remove` and the combat kinds stay as they are.
   - **(c)** As (a), and narrow `remove` too: an opponent's item on the
     stack, or a combat that involves you, but not an opponent's end step.
2. **Q2. Targeted moves.** ADR 0142 §1 does not let a targeted activated row
   declare. Its move carries `has_targets` and stops you at every key window.
   That is 638 rows, Mother of Runes among them: the classic answer held all
   game.
   - **(a) Recommended:** let a targeted activated row declare `Answers`,
     without requiring it (no ratchet), and declare the protective ones
     (slice 7). A declared targeted move is matched by its kind. An
     undeclared one stops as today.
   - **(b)** Read the announcement instead: a targeted move whose every
     target is the seat or something it controls is matched like `protect`
     and `survive`, with no declaration. Smaller, but it guesses: "untap
     target creature you control" would be read as a save.
   - **(c)** Leave targeted moves as they are.
3. **Q3. Where the match is computed.**
   - **(a) Recommended:** on the server, as `answers_now` on each move,
     which the browser and the seat read. The agreement fixture pins the
     seat's port of the windows to the browser's.
   - **(b)** The server ships the inputs as facts on the view (a threat, a
     combat that involves you, an opponent's upkeep, an opponent's spell),
     and each client matches the kinds from `purpose.answers` itself.
   - **(c)** Two ports of the whole rule, in the client and in the seat,
     pinned only by the agreement fixture.
4. **Q4. How exact "threatens something of yours" is.**
   - **(a) Recommended:** per kind, as the tables above. A target on your
     permanent, an untargeted spell, a declared sweep or edict, an attack on
     you or yours, and your creature in a block count for every kind that
     saves. A target on you counts only for `protect`, `prevent` and
     `sac_outlet`, and a target on your spell only for `protect`. In the
     declare-blockers step, an unblocked attacker at you or yours counts only
     for `protect` and `prevent`: indestructible does not stop damage to a
     player (CR 702.12b). In the #4734 game, (a) has 5 fewer stops than (b).
   - **(b)** One test for every kind: #3555's `stackThreatens` as it stands,
     plus, in declare attackers, declare blockers or a first-strike damage
     step, an attack on you or yours, or a creature of yours attacking or
     blocking.
5. **Q5. Graveyard exile and an opponent's ability.** Eight `restrict` rows
   exile from graveyards (ADR 0142 sweep ruling 9). Three do it with no
   target: Soul-Guide Lantern, Scavenger Grounds and Silent Gravestone. The
   other five choose targets, and their targeted moves stop as today
   (`has_targets`). A graveyard exile answers an opponent's reanimation or
   regrowth trigger, whose target is gone by the time it resolves
   (CR 608.2b): Eternal Witness's, in the #4734 game. The decided `restrict`
   windows match an opponent's spell, not an ability, so for those three the
   trigger passes.
   - **(a) Recommended:** split graveyard exile out of `restrict` as a new
     kind, `graveyard`, as `survive` split from `protect` on 2026-10-15
     (slice 9). It matches an opponent's spell on the stack, or an opponent's
     item that targets a card in a graveyard. `restrict` keeps the decided
     windows.
   - **(b)** Keep one kind, and add an opponent's ability that targets a card
     in a graveyard to `restrict`'s windows. Smaller, but Ranger-Captain
     would stop for those triggers too, with nothing to answer.
   - **(c)** As decided: an opponent's spell only, and a graveyard exile
     misses the trigger.

## Context

S13 shipped the per-step "stops" grid and a global `autoPassPriority`
toggle (ADR [0006](0006-priority-foundation.md)): auto-pass priority
through any step the viewer hasn't pinned, stop at every step they
have. S13.3 shipped `client/src/lib/timing.ts` — the legality
predicates (`canCastFromHand`, `canActivateAbility`, etc.) that
grey out hand cards the viewer can't play right now.

Four-player playtests exposed a rough edge the original design
didn't anticipate: **a stop means "don't skip this step," not "I
definitely have something to do here."** In practice the active
player gets stopped on their own upkeep with an empty hand, stopped
on combat steps with no creatures, and so on — a click per stop per
turn cycle, each of which adds up to a drumbeat of pointless
interactions in a 2–3 hour game.

## Decisions

### 1. The stops grid becomes an intent filter, not a gate

**Decision:** Add `gameplay.smartAutoPass: boolean` (schema v5;
default `true`). When on, the existing `autoPassPriority` effect in
`Game.svelte` folds in `hasAnyLegalResponse(snap, viewerID)`: if
the viewer holds priority on a stopped step but the legality
engine reports nothing legal, auto-pass anyway. The stops grid
still controls _where the player wants to consider stopping_; the
predicate decides whether there's actually anything to consider.

**Why a new setting instead of just changing behaviour:**

- Some players (especially sandbox testers) want every stop to
  hold open so they can think or fake plays that aren't in the
  catalog yet. Making this a toggle lets them opt out.
- Matches the pattern we used for the v2→v3 `autoPassPriority`
  re-default (ADR 0006 §3): ship the better default, give the
  holdouts a knob.

**Why default on:** the predicate is intentionally permissive
(see §3) so false-negative skips — the failure mode that would
eat a player's response — are rare. False-positive stops (the
flip side: stop when there was nothing to do) cost one extra
click and are self-explanatory. The ratio says on is the right
default.

### 2. The predicate lives in `client/src/lib/priority.ts`

**Decision:** `hasAnyLegalResponse(snap, viewerID, snapSeq?)` folds
the S13.3 timing helpers across the viewer's hand, command zone,
and viewer-controlled battlefield cards. Returns true as soon as
any predicate returns legal.

Memoised by `snap.seq` so repeat calls inside one snapshot window
(the autoPassPriority effect + any future UI consumer) don't
rescan the same state.

**Why a new module and not an extension of `timing.ts`:** scope
clarity. `timing.ts` is the per-action predicate layer
(card-in-hand vs. ability vs. pass). `priority.ts` is the
aggregate question "anything at all?" built on top. Keeping them
separate lets `timing.ts` stay focused on the grey-card UX and
`priority.ts` stay focused on the auto-pass UX without each
module's API growing tendrils into the other.

### 3. Predicate is conservative (false-positive-stop, not false-negative-skip)

**Decision:** Treat any viewer-controlled battlefield card as
"potentially has an activated ability" (via `canActivateAbility`,
which is the generic instant-speed gate). Treat any commander in
the command zone as castable via the same predicate as hand
cards.

**Why:**

- Skipping a priority window the player wanted is the bad
  outcome. The grey-card UX recovers from a false positive (you
  see the grey, you click pass); there's no recovery from a
  skipped window short of undo.
- We don't know per-card ability lists on the client — the
  server's effect catalog (S14) does, but the client doesn't
  carry that metadata on-wire. Treating every permanent as
  _potentially_ activatable keeps the predicate's false-negative
  rate near zero.

### 4. Mana affordability is out of scope (S13.6 ships without it)

**Decision:** `hasAnyLegalResponse` does **not** check whether the
viewer can afford the mana cost of any of their hand cards. A
player with a 7-drop and zero lands still counts as "has a legal
response" under this predicate.

**Why not:**

- Mana affordability requires the server's `/auto-tap-preview`
  endpoint (S15). Hitting it on every priority window for every
  card in hand, on every snapshot tick, is a perf cliff. The
  endpoint is designed for one-shot lookups, not hot-path
  scanning.
- The predicate's job is to skip _dead air_, not to divine
  intent. A player who can't afford their one sorcery usually
  still wants to see their upkeep land — if only to draw into
  something castable on the next priority window.

Future sprint (punted): a light affordance-precomputation on the
client that runs once per snapshot would let the predicate
tighten. Not shipping here.

### 5. Manual one-time stops (click-to-pin on phase icons)

**Decision:** `client/src/lib/priorityStops.ts` owns a Svelte store
of pinned `StepID`s. Clicking a priority-granting phase icon on the
PhaseDisplay track toggles that step in the set. When the auto-pass
effect sees a pinned step, it short-circuits before any other rule
(stops grid, smartAutoPass) — the cursor holds regardless. The pin
is consumed on the next snapshot step transition, so the next cycle
of that step is unpinned unless re-clicked.

**Precedence order (strongest first)** — amended by #526, which put the pin
above the autopass toggle of decision 6 as well:

1. Manual pin → hold (this decision).
2. `stepStops[step] === true` + `smartAutoPass=true` + predicate
   says "something to consider" → hold.
3. `stepStops[step] === true` + `smartAutoPass=false` → hold.
4. Everything else → auto-pass.

**Why manual stops override smartAutoPass:**
The whole point of a manual pin is "I want the cursor even though
the engine sees no reason for it." A player might want to think
about an upcoming line, fake a cast to bluff a counter, or
consider a play the catalog doesn't model yet (our coverage is
opt-in; lots of legal actions aren't predicate-visible). If
smartAutoPass could skip a pinned step, the affordance would be a
lie — so pins sit above the predicate in the hierarchy.

**Why priority-granting only:**
Untap and Cleanup don't grant priority (CR 502.4 / 514.3). Pinning
them would put a visible affordance on a step where the server
actively rejects `pass_priority`. `canManuallyStop(step)` gates
the toggle; the click handler no-ops on non-priority steps, and
the rendered icon shows a "no priority" tooltip so the missing
affordance is self-explanatory.

**Why one-time, not sticky:**
Sticky pins are the stops grid (persisted, persistent intent).
Manual pins are now-intent ("this cycle I want to see declare
attackers"). Sticky would duplicate the grid with worse ergonomics
(no Settings UI to unpin) and risk "I pinned this days ago and
forgot" stuck-cursor incidents. One-time consume on step
transition matches the feature as described and keeps state
disposable.

**Why not persisted across reloads:**
The store is process-scoped — a reload drops all pins. Consistent
with one-time semantics (a reload is a stronger signal than a step
transition), and anyone who reloads mid-game probably wants the
default stops behaviour to reassert.

### 6. Pared-down priority toolbar: just `next` + `autopass`

**Decision:** Replace the (pre-S13.6) six-button priority toolbar
(next step / pass priority / ⇥ pass step / → next stop / pass
until end of turn / pass turn) with two controls that live inside
the PhaseDisplay box itself, below the priority pills:

- **`next`** — fires a single `pass_priority`. Labelled "next"
  because the meaningful action is "move to the next priority
  window," not "pass [the thing I have]."
- **`autopass`** — a session toggle (not a one-shot). When on,
  the auto-pass `$effect` ignores every gate (stops grid,
  smartAutoPass predicate, ~~manual pins~~, `settings.autoPassPriority`)
  and fires `pass_priority` on every snapshot where the viewer
  holds priority. Persists until clicked off or reload.
  **Amended (#526): manual pins are no longer among the gates it
  ignores** — see the amendment near the top of this ADR.

`pass turn` (active-player-only whole-turn skip) and `undo` stay in
the Game.svelte toolbar — they're not priority passes, and they
belong with the other table-wide shortcuts.

**Why the toolbar got pared down:**
The intelligent auto-pass system (stops grid + smartAutoPass +
manual pins) makes the batch-pass buttons redundant — the cursor
advances to the next configured stop automatically. The only thing
a player needs a button for is the two extreme cases: "one more
step please" (`next`) and "I'm out, stop asking" (`autopass`).

**Why autopass lives inside PhaseDisplay (not the main toolbar):**
The priority pills + phase track + step label are the information
context; the two buttons are actions against that context. Keeping
them in the same visual box trims eye-travel and makes the widget
self-contained. Game.svelte's main toolbar stays focused on
one-shot game actions (draw, untap all, shuffle, life, pass turn,
undo) that aren't priority-specific.

**Why autopass must only fire when the viewer holds priority:**
The first iteration of autopass was a `passToEnd`-style loop that
bashed `pass_priority` until the step hit cleanup. That loop
didn't check `viewerHasPriority` between sends — during an
opponent's turn, priority rotates off the viewer, and every
subsequent send was rejected with `bad_request: you do not hold
priority`. The toggle form sidesteps this: the `$effect` fires
only when `viewerHasPriority` is true, so opponents' turns don't
generate rejection spam.

### 7. Autopass safety belt (`autopassPersistThroughTurns: false`)

**Decision:** The autopass toggle auto-clears the first time the
cursor enters the viewer's own `precombat_main` step — preventing
the nightmare case where a forgotten autopass skips your own turn.
A new gameplay setting `autopassPersistThroughTurns` (default
`false`, schema v6) is the opt-out: flip it on to keep autopass
engaged indefinitely. Labelled DANGER in the Settings UI with
explicit warning copy ("WARNING: ENABLING THIS SETTING MAY CAUSE
YOU TO SKIP YOUR OWN TURN").

**Why the safety is default on:**
The autopass toggle is tempting to leave on across turns ("I'll
turn it off when it matters"), but the failure mode — losing your
entire main phase to a one-click-forgot — is much worse than the
inconvenience of having to re-click autopass each rotation.
Default-on safety puts the sharp edge behind an explicit opt-in.

**Why precombat_main (not earlier):**
A paranoid reading would clear autopass on the viewer's untap
step, but that's premature: most players actively want autopass to
carry them through their own upkeep and draw steps when they're
tapped out and nothing is happening. The one step that almost
universally matters is precombat_main — that's where you cast
things. Disabling there is the minimum-surprise default.

**Why a danger label (not just a toggle):**
The opt-in doesn't merely change UX — it changes whether the
player can lose a turn to a forgotten UI state. Danger styling
(amber border, WARNING copy in caps) matches how other
minefield-adjacent settings are marked in similar projects. The
cost of over-warning a rare "I know what I'm doing" case is low;
the cost of a player flipping the toggle without noticing is a
missed turn.

## Consequences

### Good

- A four-player turn cycle where nobody has anything to respond
  with auto-walks from Untap to Untap with exactly the clicks
  players intentionally made — no drumbeat of pointless passes.
- The predicate surface is tiny (one function, one cache) and
  trivially extensible. When S13.5 visibility tightens, we can
  refine the opponent-legality questions; when S15-derived mana
  affordability is worth precomputing, we can fold it in.

### Tradeoffs

- False-positive-stop: the viewer controls _any_ permanent →
  `hasAnyLegalResponse` returns true. In Commander that's almost
  every mid-game priority window. Practically this means
  smart-auto-pass shines in empty-board and empty-hand scenarios
  (turn 1, board-wiped mid-game, etc.) and is a gentle no-op
  once players are established. That's fine — the frustration it
  addresses is precisely the empty-state drumbeat.
- Server-side authority: smart-auto-pass dispatches real
  `pass_priority` actions. If the client's predicate diverges
  from the server's legality model (e.g. an unshipped split-
  second effect that the server enforces but the client doesn't
  surface), the client might auto-pass through a priority window
  the server would have given the viewer anyway. Mitigation: the
  server's action dispatcher is authoritative — the worst case is
  the exact same "viewer had the chance, nothing happened" state
  you'd get by clicking pass by hand.

### Not done

- ~~**Opponent "thinking" indicator.** Accurate telegraph of
  "opponent can respond" requires their full hand / board
  visibility, which S13.5 doesn't grant across seats. Future
  sprint.~~ Resolved by #1307: the chip shows a hold, not
  whether the hold is real. See "'Considering a response…'" in the
  #1307 amendment.
- **Triggered-ability responses.** S19 will auto-fire catalog
  triggers; predicate doesn't consider them since they aren't
  the viewer's action to dispatch.
- **Mana affordability.** Out of scope per §4.
- **Per-player granularity on the smartAutoPass toggle.** Global
  for now; split if users ask.
