# ADR 0009 — Smart priority auto-pass (S13.6)

**Status:** Accepted · 2026-04-23 · Sprint S13.6
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
    taps or untaps the enchanted creature, doubles damage, or makes a creature
    token that does not enter tapped (populate and amass included).

  It rides in `capLegalMoves`' key, now `(source, kind, targets_stack,
  has_targets, interacts, combat_interacts)`. The change to `interacts.go` is
  none: `abilityInteracts` is asked first, and a row it marks is never marked
  again.
- **The combat window** (`inCombatWindow`, `client/src/lib/responseWindow.ts`):
  beginning of combat, declare attackers and declare blockers, or a triggered
  or activated item on the stack whose text names an attack or a block. Combat
  damage and end of combat are too late to make a blocker. There,
  `classifyMove` classes a `combat_interacts` activation as `ability` (on by
  default); elsewhere it stays `untargeted` (off by default). So a crew or a
  manland stops you at an opponent's declare attackers or blockers, and on a
  ticked beginning of combat, and not for an opponent's main-phase sorcery.
- **The audit.** Of the catalog's 478 untargeted, instant-speed activated rows
  that `interacts` does not mark, the rules mark 110 as combat abilities: 27
  crew rows, the manlands and animated artifacts, the keyword grants, the
  extra-block rows and the untapped creature-token makers. The rest were read
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

- A token maker with `{T}` (Castle Ardenvale, Kjeldoran Outpost) now stops you
  at every opponent's declare attackers, whoever is attacked. A chump blocker
  is a real play, and the owner's goal weighs a missed block above a click.
- `combat_interacts` is read from printed text like `interacts`, with the same
  risk of a misread either way; the test lists are where one is fixed.
- A pump written without a `+N/+N` ("Double this creature's power") is still
  read as value by `interacts`; #2872's declared answers is the place to fix
  that, not this text reader.

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
