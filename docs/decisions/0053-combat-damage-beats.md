# ADR 0053 — Combat damage beats: first strike and regular damage as two visible moments

**Status:** Accepted · 2026-09-16 · S18 polish, tracked on [#187](https://github.com/krakenhavoc/cmd_and_ctrl/issues/187)
**Numbering:** 0052 is skipped on purpose. The owner assigned it to the
emblems ADR in the "Decisions 2026-09-16" comment on
[#623](https://github.com/krakenhavoc/cmd_and_ctrl/issues/623), and that
ADR has not been written yet. On 2026-09-16 every remote branch was
checked with the AGENTS.md §4 loop, twice (once when drafting, again
after `git fetch --all --prune` on acceptance), and none has a
`docs/decisions/0052-*` or `0053-*` file (the highest on any branch is 0051).
**Builds on:** [ADR 0014](0014-combat-keywords.md) §3 (the two-pass
combat damage rewrite, PR #181, follow-ups in #189) and
[ADR 0033](0033-ai-bot-seat.md) §4 (the public game log, `GameView.log`).
It also borrows the S22 reveal-window pattern (`client/src/lib/reveals.ts`).
**Decided by the owner (2026-09-16):** the wire shape (Decision 1), the
reduced-motion rule (Decision 2), the no-empty-beat rule (Decision 3),
and putting the per-pair visual on `CombatArrows.svelte` (Decision 4).
**Review questions, answered by the owner the same day:** review found
that Decision 4 as first written could show nothing for a creature that
died in the frame, which is the main case for blocked first strike, and
raised four questions. All four are decided:

1. #187 ships **presentation-only**. The real second combat damage step
   is [#717](https://github.com/krakenhavoc/cmd_and_ctrl/issues/717)
   ([Decision 5](#decision-5--presentation-only-the-second-combat-damage-step-is-717)).
2. The pause is about 400 ms **scaled by `animations.speed`**, and
   `"still"` mode holds beat 1's text cue for the pause without motion
   ([Decision 2](#decision-2--reduced-motion-keeps-a-motionless-cue)).
3. A creature that died gets a **ghost arrow from cached geometry**
   (option a). Beats reuse the `damagePopups` toggle, with no new
   setting ([Decision 4](#decision-4--beats-cue-the-existing-combat-arrows)).
4. A frame that leaves the step mid-sequence **does not cut** the
   remaining cues ([Beats across frames](#beats-across-frames)).

Decision 3's rule stands, but one of its examples only holds once an
engine bug is fixed (bug A below, [#715](https://github.com/krakenhavoc/cmd_and_ctrl/issues/715)).

Line references are to `origin/develop` at `7f7427d`, except where a
reference names PR #709 (head `605458c`).

## Context

### The engine already does two passes

The #187 issue text says there is no engine work to do, and that part is
right. `resolveCombatDamageLocked` (`server/internal/game/mutations.go:4415-4447`)
runs combat damage in two passes:

1. A first-strike pass, only when `hasAnyFirstStrikeCombatants` (`:4453`)
   finds an attacker or blocker with first strike or double strike.
2. A state-based-action and trigger drain (`runStateChecksLocked`, `:4430`)
   and a layer recompute.
3. A regular pass (`:4439`), which always runs, followed by another SBA drain.

`participatesInSubstep` (`:4470`) decides which creatures deal damage in
which pass. This ADR calls the two passes "substeps" to match the code.
The Comprehensive Rules call them two combat damage steps (see
[Decision 5](#decision-5--presentation-only-the-second-combat-damage-step-is-717)).

**How many frames.** When nothing pauses, both passes run inside one
locked mutation, so every client receives one frame holding the end
state of both. Two things already pause combat damage part-way and
split it across frames:

- **A CR 510.1c damage-assignment prompt.** An attacker blocked by two or
  more creatures queues a prompt (`mutations.go:4624`), and its damage
  lands only when the prompt is answered (`ResolveDamageAssignment`,
  `pending_choice.go:1326`). That is a later frame, possibly seconds
  later while a human decides.
- **A CR 616 replacement ordering prompt.** Each combat damage helper
  returns early on `errReplacementPending` (`mutations.go:4703`, `:4770`,
  `:4820`, `:4934`), and that damage lands when the ordering is answered.

[#702](https://github.com/krakenhavoc/cmd_and_ctrl/issues/702) makes the
split permanent for one common case (see [Beats across frames](#beats-across-frames)).

**Citation note.** The #187 text, ADR 0014 §3 and the code comments
cite CR 510.2 and 510.3 for the two substeps. Those citations are wrong
in the pinned edition (Comprehensive Rules effective August 7, 2026).
There, 510.2 is "all assigned combat damage is dealt simultaneously",
510.3 is "the active player gets priority", and the first-strike split
is **CR 510.4** (repeated in 702.4b and 702.7b, and summarised in 506.1).
[#693](https://github.com/krakenhavoc/cmd_and_ctrl/issues/693) Part 2
tracks fixing the 13 wrong sites. This ADR cites 510.4 throughout.

### Bug A: the engine forgets that an attacker was blocked

The rules keep an attacker blocked once blockers are declared. CR 509.1h:
"A creature remains blocked even if all the creatures blocking it are
removed from combat." A creature leaving the battlefield is removed from
combat (CR 506.4). A blocked creature with no blockers left assigns no
combat damage (CR 510.1c). The one exception is trample: its damage goes
to the player or planeswalker it attacks, as though every blocker had
been assigned lethal damage (CR 702.19d).

The engine keeps no record that an attacker was blocked. Each pass
rebuilds `blockersByAttacker` from the live battlefield
(`mutations.go:4507-4513`), and an attacker with no live blockers
(`:4571-4580`) is treated as unblocked and hits its attack target
(`:4586-4589`). The menace close-out (`:4514-4526`) runs on the same
live map, so it also reverts a block that was legal when declared.

Three throwaway probe tests were run against `7f7427d` on 2026-09-16,
built on the `combat_test.go` helpers and deleted afterwards:

| scenario | rules | engine |
|---|---|---|
| 1/1 double strike blocked by a 1/1; the blocker dies in the first-strike pass | 1 combat damage event; defender stays at 40 | 2 combat damage events; defender 40 → 39 |
| 2/2 blocked by a 1/1; the blocker is destroyed before combat damage | no damage; defender stays at 40 | defender 40 → 38 |
| 3/3 menace blocked by two creatures; one is destroyed before combat damage | the attacker assigns 3 to the remaining blocker | the remaining block is reverted; defender 40 → 37 and the blocker takes 0 |

So the bug is not specific to first strike. Any removal of a blocker
before damage, including an instant during the declare blockers step,
turns a blocked attacker into an unblocked one. No open issue covers it
(open issues searched for blocked, blocker removed, menace and first
strike). [#672](https://github.com/krakenhavoc/cmd_and_ctrl/issues/672)
(a "remove from combat" verb) asks how the engine should represent
"stays blocked" but does not cover damage. The fix is a "blocked" flag
on the attacker, set when blockers are declared and kept until it is
removed from combat or combat ends (CR 509.1h, 506.4). **Bug A is filed
as [#715](https://github.com/krakenhavoc/cmd_and_ctrl/issues/715).**

### What the wire carries today

Every combat damage event already reaches the client in order, but
nothing marks which substep it belongs to:

- `game.Event` has a `Combat bool` (`server/internal/game/events.go:535-544`),
  and has no substep field.
- The public log projects each `EventDealDamage` into a `LogDamage`
  entry with `seq`, source `card_id`, target, `amount` and `combat`
  (`server/internal/protocol/log.go:417-428`, client mirror
  `client/src/lib/protocol.ts:376-405`). `GameView.log` is a projection
  of `Game.Events` capped at `PublicLogMax = 200` (`log.go:76`). It is
  not a stored buffer, so it survives undo, snapshot restore and deploy.
- `LogStep` entries mark step boundaries, and `LogBlock` entries record
  each block with the blocker as `card_id` and the attacker as `target`
  (`log.go:439-443`).
- Creatures that die in the SBA pass between substeps show up as
  `LogZone` entries between the two groups of damage entries.

A client could *guess* where the boundary falls: a source that deals
damage twice must have double strike, and a zone move in the middle of
the damage means SBA ran. The guess fails when a first-strike attacker
hits a blocker that survives, because nothing dies in between.

### What the client shows today

Nothing on the client shows a creature taking damage over time.
`damage_marked` is a static badge on `Card.svelte`. No out-animation
exists for a permanent leaving the battlefield (the comment is at
`client/src/lib/animations.ts:152-154`). The only combat-damage signal
is `play("combat_resolve")` when the step changes to `combat_damage`
(`client/src/routes/Game.svelte:495-496`). The life popup
(`PlayerIdentity.svelte:139-165`) shows only the last life change per
frame ([#703](https://github.com/krakenhavoc/cmd_and_ctrl/issues/703)).

Two pieces already exist that this design can build on:

- **`CombatArrows.svelte`** draws an arrow for each attack pair
  (attacker to player) and each block pair (blocker to attacker). The
  engine deliberately keeps `AttackingTarget` and `BlockingTarget`
  through the whole `combat_damage` step so these arrows stay drawn
  (`mutations.go:4385-4390`, cleared by `clearCombatLocked` on the way
  to `end_combat`). What it can draw is limited by how it is built:
  - Pairs come from the **live** `view.battlefield.cards` (`:38-59`). A
    block pair also needs its attacker on the battlefield (`:51`). A
    creature that has left takes its pairs with it.
  - Each endpoint is measured from a mounted DOM node
    (`[data-instance-id]` or `[data-seat-id]`, `:130-139`), and a pair
    with an unmounted endpoint is skipped (`:153-175`).
  - An attack arrow always looks for a player header (`data-seat-id`,
    `:160-161`). An attack on a planeswalker or battle has a card ID in
    `attacking_target` (the kind is `attacking_target_kind`,
    `protocol.ts:1066`), so no arrow is ever drawn for it.
- **The S22 reveal pattern** (`client/src/lib/reveals.ts:1-27`). The
  client cues an entry the first time it sees its `seq`, and never
  again. On the first frame after joining it *primes*: everything
  already on the wire is marked seen and not shown. That pattern already
  copes with dropped frames and reconnects.

### Reduced motion is plumbed in two separate places

A beat sequencer cannot rely on the existing gate, because that gate is split:

- `animations.ts` `gatedDuration` (`:57-64`) honours
  `animations.enabled` and the per-effect toggles. The settings bridge
  (`client/src/lib/settings.ts:535-546`) pushes **only** `animations.*`
  into it. `accessibility.reduceMotion` never reaches it.
- `accessibility.reduceMotion` works through a global CSS rule
  (`client/src/app.css:122-126`, `data-reduce-motion` set in
  `App.svelte:126`). That rule shortens CSS animations and transitions.
  It does **not** stop a GSAP tween or a `setTimeout`.
- A change in the OS media query flips both `reduceMotion` and
  `animations.enabled` (`settings.ts:563-571`). Toggling reduceMotion
  in the app flips only `reduceMotion`.
- Components that care check both themselves. The S11.5 precedent is
  `PlayerIdentity.svelte:115-121`:
  `$settings.animations.enabled && !$settings.accessibility.reduceMotion`,
  with the house rule "the information survives, the motion does not".
- Some animations read no settings at all. The `CombatArrows` draw-on
  (`:235-247`) is a raw `gsap.fromTo` with a fixed 0.34 s duration. It
  ignores `animations.enabled`, `animations.speed` and `reduceMotion`
  alike.

### The acceptance criteria in #187 are wrong

The issue's AC1 expects Fencing Ace against Grizzly Bears to show "bear
takes 1 → bear dies → pause → no further damage, Fencing Ace survives".
Neither the rules nor the engine produce that. Fencing Ace is a 1/1
with double strike (`fencing_ace.go:6`). Grizzly Bears is a 2/2. In the
first-strike substep Ace deals 1 and the bear **survives**. In the
regular substep Ace deals 1 more while the bear deals 2 back, and
**both die**. `TestCombatDoubleStrikeHitsTwice`
(`server/internal/game/combat_test.go:168-187`) pins exactly this
outcome. The issue's other forcing scenarios have the same arithmetic
errors:

- Unblocked, Ace deals 2, not 1.
- Ace and Youthful Knight (2/1 first strike, `youthful_knight.go:6`)
  each mark 2 on a Colossal Dreadmaw, not 2 and 1.
- Against Bears the end states already differ: Ace dies and Knight lives.

The acceptance criteria below replace the issue's.

## Decision 1 — The substep is a tag on the durable event stream

The engine stamps each combat `EventDealDamage` with the substep that
dealt it. The public log projection carries the stamp onto `LogEvent`.
There is **no** ephemeral `CombatDamageBeats` sidecar on `GameView`.

**Shape.** One new field on each struct, `omitempty` on all of them:

| struct | field | JSON | values |
|---|---|---|---|
| `game.Event` | `CombatStep string` | `combat_step` | `"first_strike"`, `"regular"` |
| `protocol.LogEvent` | `CombatStep string` | `combat_step` | same |
| `client/src/lib/protocol.ts` `LogEvent` | `combat_step?: "first_strike" \| "regular"` | | |
| `game.DamageAssignmentFrame` | `CombatStep string` | (server-side only; not added to `DamageAssignmentView`) | `""`, `"first_strike"`, `"regular"` |

**When the field is set.** Both substeps are tagged **only when the
first-strike substep ran** (`hasAnyFirstStrikeCombatants` returned
true). A combat with no first strike or double strike anywhere has
untagged combat damage, exactly as today, including damage that lands
later from a prompt. The wire then says directly that there is nothing
to sequence. Snapshots written before the field existed decode untagged
too, which is the same answer.

**Why the frame needs its own field.** `DamageAssignmentFrame.FirstStrike`
(`pending_choice.go:632-636`) only says "first-strike pass" or "regular
pass". A regular-pass prompt has `FirstStrike = false` whether or not
the first-strike pass ran. If the resume paths tagged from it, a vanilla
5/5 blocked by two 2/2s would come back tagged `regular` in a combat
with no first strike anywhere, which breaks the rule above and
acceptance criterion 6. So `queueDamageAssignmentPromptLocked`
(`mutations.go:4658-4678`) stores the pass's `CombatStep` value on the
frame (`""` when the first-strike pass did not run), and the resume
paths read that. `FirstStrike` stays as it is for #702's use. The new
field's zero value is `""`, meaning untagged, so a snapshot with a
pending prompt from before the field existed resumes untagged. The cue
is lost and the board is right. Under the rule at
`server/internal/game/snapshot.go:70-86` that needs no
`SnapshotSchemaVersion` bump.

**Where it is stamped.** Sub-PR 1 lands **after**
[PR #709](https://github.com/krakenhavoc/cmd_and_ctrl/pull/709)
(the fix for [#694](https://github.com/krakenhavoc/cmd_and_ctrl/issues/694)).
#709 gives `ReplacementEvent` an unexported `damageTail` snapshot and
routes every damage entry point, and the CR 616 resume, through one
`applyResolvedDamageLocked`, with one emit site, `emitDealDamageLocked`
(`damage_tail.go` on the PR). The tag rides that tail:

- `damageTail` gains `combatStep string`, and `emitDealDamageLocked`
  copies it onto `Event.CombatStep`. That is the only place the event
  field is written.
- **Direct paths.** `combatDamageTailLocked` (called from
  `markCombatDamageOnCardLocked` and `markCombatDamageToPlayerLocked`)
  takes the step as an explicit argument. `resolveCombatDamageLocked`
  passes it down through `assignAndDealCombatDamageLocked`. It is an
  argument, not a field on `Game`, so no stale value can outlive a pass.
- **Prompt-resume paths.** `damageTailFromFrame` copies
  `DamageAssignmentFrame.CombatStep`.
- **CR 616 resume.** The tail is on the event, so the tag comes back with
  it. There is no separate branch to tag.

Before #709, the same job means four separate tails plus the
`RepEventDamage` resume branch (`pending_choice.go:1201-1219`), which
drops `Combat` and `Actor` (#694) and which #709 deletes. Stamping there
first would be throwaway work.

**Log text.** The server-rendered `Text` for a tagged `LogDamage`
entry says which step dealt it, for example "Fencing Ace dealt 1 combat
damage to Grizzly Bears (first strike)". The log panel, bug reports and
the model-backed bot tiers then all see the distinction without any
client logic. Wording is settled in review.

**Why the log and not a sidecar:**

- **It survives what the sidecar would not.** The log rides
  `Game.Events`, so a dropped frame, a reconnect or a deploy restore
  still carries the tags. The issue's "cleared the instant the client
  acks" needs an ack signal that protocol v0 does not have, plus new
  per-viewer filtering plumbing for a second channel.
- **It follows an existing pattern.** It is the S22 reveals shape (cue
  by `seq`, prime on first frame) applied to entries already on the
  wire. It adds a few bytes per tagged entry, and none for combat
  without first strike.
- **Visibility is already handled.** `LogEvent` goes through
  `FilterViewFor`'s knower predicate. Combat damage is between
  battlefield permanents and players, so it is public anyway.

### Rejected: tag every combat, and let the client ignore regular-only steps

Tagging `regular` unconditionally would avoid the frame field. The
client would treat a `combat_damage` step with no `first_strike` entry
as nothing to sequence. Rejected: the log text would then say "(regular
damage)" in every combat, and a tag whose meaning depends on what else
is in the step is client-side inference of exactly the kind this
decision exists to avoid.

### Rejected: a client-only heuristic

The client already holds an earlier frame from `declare_blockers`
or a priority pass during combat. In it, every combatant's
`CardView.abilities` lists its effective keywords, including granted
ones. So the client *could* classify each `LogDamage` entry by looking
up its source's keywords. That would handle most of the cases where the
boundary can't be found from the log alone.

It is rejected for three reasons:

1. **House rule.** `protocol.ts:1193-1196` says "READ IT, DON'T DERIVE
   IT", and #429 deleted a pile of client-side rules re-derivation.
   Deciding which creature deals damage in which combat damage step is
   rules logic (CR 510.4, 702.4c-d, 702.7c).
2. **Wrong input.** The earlier frame's keywords are not the ones
   CR 510.4 reads. Keywords can change between that frame and the
   moment each pass begins (a lord granting first strike dies in the
   first pass), and the client never receives the state in between.
3. **Unreliable input.** The earlier frame may never have arrived
   (reconnect straight into `combat_damage`), and a card that died
   takes its `CardView` with it.

## Decision 2 — Reduced motion keeps a motionless cue

With `animations.enabled = false` or `accessibility.reduceMotion = true`,
a combat that had a first-strike substep still says so. The
information survives, the motion does not (S11.5,
`PlayerIdentity.svelte:115-121`).

- The sequencer has two modes, `"full"` and `"still"`. The mode is a
  pure function in `combatBeats.ts`,
  `beatMode(animationsEnabled, damagePopups, reduceMotion)`. It returns
  `"full"` only when `animationsEnabled && damagePopups && !reduceMotion`,
  which is `PlayerIdentity`'s predicate plus the per-effect toggle, and
  it is unit-tested. The Svelte shell passes it the values from the
  **settings store**. It does **not** read `animations.ts`'s module
  config, because the settings bridge (`settings.ts:535-546`) never
  passes `reduceMotion` to it, and the CSS rule in `app.css` cannot stop
  a `setTimeout` or a GSAP tween.
- **No new settings toggle (owner, 2026-09-16).** Beats reuse the
  existing `animations.damagePopups` per-effect toggle, the one that
  already gates the life popup. Turning it off gives `"still"` mode,
  not silence, because the information survives.
- In `"still"` mode the sequencer starts no tweens and no pulses. Each
  beat that happened is shown as a **text** cue: a "First strike" label
  (and "Regular damage" when both beats happened) next to the damage it
  describes. The cue must be real text in the accessibility tree. The
  arrow overlay is `aria-hidden="true"`, so a label inside that SVG
  does not count.

### Timing (owner, 2026-09-16)

- **The pause.** When both beats arrive in the same frame, beat 2 starts
  `BEAT_PAUSE_MS × animations.speed` after beat 1, with
  `BEAT_PAUSE_MS = 400`. That is the scaling `gatedDuration`
  (`animations.ts:57-64`) applies to other effects: speed 2 is twice as
  slow, 0.5 twice as fast. The schedule function takes `speed` as an
  input and does the multiplication itself. It does **not** call
  `gatedDuration`, which snaps to 1 ms when animations or the per-effect
  flag are off and would erase the `"still"` hold below.
- **`"still"` mode holds beat 1.** Beat 1's text cue shows alone for the
  same scaled pause, with no motion, and then beat 2's cue joins it. The
  cues are never shown all at once, so the order of the two steps still
  reads. Only motion is removed.
- Each pulse and each ghost-arrow fade is its own short effect, and its
  duration goes through the same speed scaling.

## Decision 3 — A substep that dealt no damage gets no beat

A beat exists only if its `combat_damage` step has at least one combat
`LogDamage` entry with that `combat_step` (see
[Beats across frames](#beats-across-frames) for what "its step" means).
There is no empty beat and no pause for one. This covers:

- **A plain first striker.** The regular substep still runs
  (`mutations.go:4439` runs unconditionally) and deals nothing.
- **Prevention.** A Fog-class effect that prevents every point of
  first-strike damage. A prevented event is never emitted, so it gets
  no entry.
- **A double striker whose only blocker died in the first substep, once
  bug A is fixed.** By the rules it is still blocked and assigns no
  damage in the second step (CR 509.1h, 510.1c), so it shows one beat,
  as a first striker would. **Today the engine gets this wrong**: it
  deals the regular damage to the defending player, that event is
  tagged `regular`, and it is a real second beat. The beats show what
  the engine did. Acceptance criterion 2 depends on bug A.
- **Not covered: trample.** A double striker with trample whose blockers
  are all gone assigns all its regular damage to the player or
  planeswalker it attacks (CR 702.19d). That is a real second beat, both
  today and after bug A is fixed.

The rule is about what was dealt, not about which keywords a creature
has. Where a creature has double strike, the card's keyword badge says so.

## Decision 4 — Beats cue the existing combat arrows

**Owner decision.** The per-pair visual lives on the arrows
`CombatArrows.svelte` already draws. No new overlay is added. For a
creature that has already left, the arrow is drawn from cached geometry
(decided 2026-09-16, [below](#creatures-that-died-cached-arrow-geometry)).

**Which arrow an entry names.** For each combat `LogDamage` entry in a
beat, the arrow ID comes from the log alone. Roles come from this
turn's `LogBlock` entries (the latest one per blocker), not from live
battlefield state:

| entry | arrow |
|---|---|
| target is a player (`target_seat` set) | `atk-<source>` |
| target is a card, and a `LogBlock` says the target blocks the source | `blk-<target>` |
| target is a card, and a `LogBlock` says the source blocks the target | `blk-<source>` |
| target is a planeswalker or battle the source attacks | none: `CombatArrows` never draws an attack arrow to a card (`:160-161`) |
| no role found (the `LogBlock` fell out of the 200-entry window) | none |

An entry with no arrow still gets the text cue.

**Why live arrows are not enough.** The frame that carries the beats is
already the end state. Every creature that died in it is gone from
`view.battlefield`, so its pairs are gone (`:38-59`), and its DOM node is
gone, so `recomputeArrows` could not measure it anyway (`:153-175`).
Taking the pair list "at sequence start", as the first draft said, does
not help: at sequence start that list has already lost those pairs.
Blocked first strike is mostly used to kill the blocker, so with live
arrows only, the owner-chosen visual would show nothing in its main case.

### Creatures that died: cached arrow geometry

**Decision (owner, 2026-09-16, option a).** `CombatArrows` keeps each
arrow's last measured geometry, keyed by arrow ID. When a beat names an
arrow whose endpoint is gone, it draws a fading **ghost arrow** from the
cache instead of skipping it.

- **What is cached.** On every `recomputeArrows` during combat, each
  attack and block arrow whose endpoints were both measured writes its
  `from`, `to` and control point to the cache. The coordinates are
  already board-relative (`rectIn`, `:130-139`), so page scrolling does
  not move them. An arrow ID that stops being measured keeps its last
  entry.
- **How long.** The cache fills from `declare_attackers` on. An entry is
  kept through the `combat_damage` step and until that step's last
  scheduled cue has played, even if the game has moved on (see
  [Beats across frames](#beats-across-frames)), and is then dropped. A
  later measurement of the same arrow ID (an extra combat) overwrites
  it with fresher geometry. The cache is also cleared when the client
  primes, so a reconnect never draws geometry it did not measure itself.
- **Live, ghost or nothing.** For each arrow ID in a beat: if both
  endpoints are mounted, pulse the live arrow. If an endpoint is gone
  and the cache has the ID, draw a ghost. Otherwise there is no arrow,
  and the text cue carries the beat. The choice is a pure function in
  `combatBeats.ts` (`arrowRender(arrowID, liveIDs, cachedIDs)` →
  `"live" | "ghost" | "none"`), so it is unit-tested and the shell only
  draws.
- **Mounted endpoints are re-measured.** A ghost uses cached coordinates
  only for the endpoint that is gone. An endpoint still on the board is
  measured again at cue time.

**Risk: reflow.** When a card leaves, `BattlefieldRow` reflows, and the
cards beside it slide into the gap. A cached endpoint then points at
where the dead card used to be, and possibly at a different, living card
that now sits there. A ghost could read as "this damage went to that
card". Mitigations, all inside `CombatArrows`:

- **Look different.** A ghost is drawn in a separate style (dashed or
  low-opacity stroke, no endpoint glow). It never looks like a live arrow.
- **Short and fading.** A ghost fades in and out within its own beat
  (scaled by `animations.speed`) and is never left on screen, so it
  reads as "was here", not as a lasting link to whatever is there now.
- **Drop, don't guess.** If the board's size changed since the cached
  measurement (a window resize, or a row wrapping onto another line),
  the ghost is not drawn and the text cue carries the beat. Only the
  overlay is lost, and the information survives in the text.

**No per-creature damage number.** Beats reuse `damagePopups` (see
[Decision 2](#decision-2--reduced-motion-keeps-a-motionless-cue)), but
today that toggle only drives the player life popup (`floatUp` and
`fadeOut` in `animations.ts`, used by `PlayerIdentity`). No floating
damage number on a creature tile follows from it, so this ADR adds none.
See [Deferred](#deferred).

With cached geometry, this is what each acceptance criterion gets:

| AC | arrows the beats name | cue |
|---|---|---|
| 1. Ace vs Bears, both die in beat 2 | `blk-Bears`, both endpoints gone | ghost arrow on both beats, plus text |
| 2. Ace vs a 1/1 that dies in beat 1 | `blk-<blocker>`, blocker gone | ghost arrow, plus text |
| 3. Knight vs Bears, the Bears die | `blk-Bears`, Bears gone | ghost arrow, plus text |
| 4. Ace unblocked | `atk-Ace`, both mounted | live pulse on both beats, plus text |
| 5. Mixed combat | a mix | live pulse for survivors, ghost for the rest |
| 9. Prevention | depends on who dies | live pulse for survivors, ghost for the rest |

In `"still"` mode none of these draw. The text cue carries every beat.

## Decision 5 — Presentation-only; the second combat damage step is #717

**Decision (owner, 2026-09-16).** #187 does **not** model the second
combat damage step. It ships presentation-only, as Decisions 1–4
describe. Modelling the real step is filed as
[#717](https://github.com/krakenhavoc/cmd_and_ctrl/issues/717).

**What #717 is about.** CR 510.4 (Aug 7, 2026 edition, restated in
702.4b and 702.7b, with 506.1 saying "there are two combat damage
steps") says that after the first-strike step "the phase gets a second
combat damage step", and CR 510.3 / 510.3a give the active player
priority after each step, with damage triggers put on the stack first.
The engine has one `StepCombatDamage` (`turn.go:31`) and no priority
window between passes. So a "whenever this deals combat damage" trigger
from first-strike damage resolves after regular damage, and nobody can
respond between the steps. Modelling it is a turn-structure change: the
step list in `turn.ts`, `PhaseIcon`, auto-pass and stops, the legal-move
enumerator (`server/internal/legal`), bot policy, and snapshots. That is
too large to hold #187 behind. (The 510.2 / 510.3 citation fix is **not**
part of this. #693 tracks it.)

**What happens when #717 lands.** The two beats become two real server
frames with a priority window between them, so the time between frames
is the pause. The client beat sequencer in sub-PR 2 (step keying, beat
splitting, the same-frame pause and the `"still"` hold) may then become
**largely unnecessary**, and #717 should say which parts it removes.
Sub-PR 1's `combat_step` tag still applies: it would name the step each
damage event was dealt in. The text cue and the cached-geometry ghost
arrows are still useful, because a creature that died in the first step
is still gone by the time its frame renders.

**#717 landed (2026-09-18). What it changed here, and what it did
not.** Nothing is removed. The two beats usually do arrive in two
frames now, with a real priority window between them, so the same-frame
pause fires less often — but it still has to exist: a table on autopass,
or a seat that clicks `advance_step` twice, delivers both steps in one
frame, and the pause is what keeps them readable. Three things changed,
all of them small:

- `splitBeats` keys a combat's beats on the **first** damage step entry
  it sees (`first_strike_damage` when there was one, `combat_damage`
  otherwise) instead of on the `combat_damage` entry alone. Without that
  the two steps are two `stepSeq`s, the pause never pairs and the text
  cue splits into two boxes.
- `keepArrowCache` keeps the geometry through `first_strike_damage`
  as well, so the ghost arrows a first-strike death needs survive.
- Decision 1's `combat_step` tag is unchanged and still earns its keep:
  it is what tells a first-strike entry from a regular one INSIDE a
  frame, and it still appears only when the combat had a first-strike
  step. It is now redundant with the step entry for most entries and
  not for damage that lands late from a prompt, which is the case a
  position-based reading was always wrong about.

Decisions 2, 3 and 4 (reduced motion, no empty beat, the arrow cue) are
untouched, and the animation is not rebuilt.

## Beats across frames

A beat is a set of log entries. These rules say which entries, and how
beats that span frames are played.

**The step.** A `combat_damage` step is identified by the `seq` of its
`LogStep` entry (`step == "combat_damage"`). Every entry after it, up to
the next `LogStep`, belongs to that step. A turn with an extra combat
phase has two such steps, and their beats never mix.

**Membership.**

- A tagged combat `LogDamage` entry belongs to the beat its tag names,
  in its step.
- A non-damage entry (a `LogZone` death, a `LogLife` from lifelink)
  belongs to the beat of the nearest earlier tagged damage entry **among
  the same frame's new entries**. So pass 1's SBA deaths belong to
  beat 1, and pass 2's deaths, which come after the last regular damage
  entry (the "both die" in acceptance criterion 1), belong to beat 2.
- New entries with no tagged damage entry before them in the same frame
  belong to no beat. A spell cast in the priority window of the combat
  damage step is not part of a beat.

**Spanning frames.** A beat can arrive in more than one frame, and the
two beats can arrive in different frames:

- Today, a damage-assignment prompt or a CR 616 prompt lands part of a
  beat's damage in a later frame (see [Context](#the-engine-already-does-two-passes)).
- After [#702](https://github.com/krakenhavoc/cmd_and_ctrl/issues/702)
  is fixed, a first-strike attacker blocked by two creatures always
  splits: frame N has the blockers' first-strike damage and the prompt;
  frame N+1 has the attacker's assigned damage, the SBA pass and the
  whole regular pass. Beat 1 is in both frames, and beat 2 is in the
  second, possibly seconds later.

The sequencer keeps state per step: which beats it has already cued.

- Entries for a beat already cued in an earlier frame are shown at
  once as a continuation of that beat. No pause goes before them and
  the "First strike" label is not repeated.
- The pause between beats is added only when beat 1 entries and beat 2
  entries arrive in the **same** frame. When beat 2 starts in a later
  frame than beat 1, the time between frames is the pause, and nothing
  is added.
- A `first_strike` entry that arrives after beat 2 has started (today's
  order for #702's case) is shown at once as a continuation of beat 1.
  It is not reordered.

**Leaving the step mid-sequence (owner, 2026-09-16): finish the cues.**
A bot or an auto-pass can move the game to `end_combat`, which retracts
the arrows (`clearCombatLocked`), within a few hundred milliseconds of
the frame that carried the beats. The sequencer does **not** cut the
remaining cues. They play on schedule, and any arrow they name whose
pair is no longer live is drawn as a ghost from the cached geometry
([Decision 4](#creatures-that-died-cached-arrow-geometry)), which is kept
until the last cue has played. This is safe because the cues are
overlays. The board already shows the new frame, and input stays live
against it, so finishing the cues never holds the board. A new
`combat_damage` step (an extra combat) starts its own sequence and does
not cancel the old one's remaining cues.

**Cue once, prime, and undo.** Entries are cued once, keyed by `seq`,
and a client primes on its first frame, as `reveals.ts` does. Undo
(`Game.RestoreFrom`, `server/internal/game/clone.go:450`) restores
`eventSeq` (`:477`), so events replayed after an undo reuse `seq`
numbers the client has already seen. The tracker therefore treats a
frame whose highest log `seq` is lower than the highest it has seen as a
rewind: it forgets the seen `seq`s above the frame's highest, and cues
nothing from that frame. Limit: if an undo and the replay both happen
between two frames this client receives, the replayed entries collide
with seen `seq`s and their cues are dropped. Only the overlay cue is
lost; the board is right. (`reveals.ts`'s seen set has the same exposure.
That is out of scope here.)

## How the pieces fit

```
engine   resolveCombatDamageLocked
           pass 1 → damageTail{combatStep:"first_strike"} → EventDealDamage{Combat, CombatStep} …
           SBA    → EventZoneChange (deaths)
           pass 2 → damageTail{combatStep:"regular"} → EventDealDamage …
           SBA    → EventZoneChange (deaths)
         prompt resume → damageTailFromFrame(frame.CombatStep) → EventDealDamage (later frame)
protocol publicLogOf → LogEvent{seq, kind:"damage", combat, combat_step, text}
client   combatBeats.ts  (pure)
           track(state, log) → new entries; prime on first frame; rewind on undo
           splitBeats(state, newEntries) → per step: [] | [beat] | [beat1, beat2], with continuations
           arrowIDsFor(beat, log)          — roles from LogBlock, never live state
           arrowRender(arrowID, liveIDs, cachedIDs) → "live" | "ghost" | "none"
           beatMode(animationsEnabled, damagePopups, reduceMotion) → "full" | "still"
           schedule(beats, mode, speed) → [{atMs, beat, arrowIDs, cueText}]   (pause = 400 ms × speed)
         Board.svelte / CombatArrows.svelte (thin shell)
           passes settings in, runs the schedule to the end even after the step changes,
           caches arrow geometry, pulses live arrows, draws ghost arrows, renders cue text
```

## Acceptance criteria (replacing the ones on #187)

1. **Double strike, blocker survives beat 1.** Fencing Ace (1/1 double
   strike) attacks and is blocked by Grizzly Bears (2/2). Beat 1: Ace
   deals 1 to the Bears. The Bears survive with 1 damage, and a "First
   strike" cue shows. Beat 2: Ace deals 1 more and the Bears deal 2 to
   Ace, and **both die**. The log has two tagged damage entries from
   Ace (`first_strike`, `regular`) and one from the Bears (`regular`).
   Both `LogZone` deaths belong to beat 2.
2. **Double strike, blocker dies in beat 1. Depends on bug A.** Fencing
   Ace is blocked by a vanilla 1/1. Beat 1: Ace deals 1 and the blocker
   dies. Beat 2 deals nothing, so there is **one** beat and no pause.
   Ace survives with 0 damage, and the defending player takes nothing.
   Until bug A is fixed the engine instead deals 1 regular damage to the
   defending player (40 → 39), which is two beats. Sub-PR 1 does not
   pin that behaviour in a test. Sub-PR 2 tests the one-beat rule with
   log data as input, so the test does not depend on the engine.
3. **First strike.** Youthful Knight (2/1 first strike) is blocked by
   Grizzly Bears. One beat: Knight deals 2, the Bears die, Knight
   survives with 0 damage. No second beat.
4. **Double strike to a player.** Fencing Ace attacks unblocked. Two
   beats: two `LogDamage` entries to the defending seat, 1 each, tagged
   `first_strike` then `regular`, and life drops from 40 to 38. The
   life popup is not part of this criterion. `life_history` has no key
   that joins a life change to a beat, and #703's fix may sum the
   changes, so the popup is #703's acceptance, not this ADR's.
5. **Mixed combat.** Double-strike, first-strike and vanilla attackers,
   each unblocked or blocked by one creature that is still there at
   damage, play two beats in `seq` order. Every first-strike entry
   comes before every regular entry. Multi-blocker first strike and
   double strike are **excluded** until
   [#702](https://github.com/krakenhavoc/cmd_and_ctrl/issues/702) is
   fixed. Blockers removed before damage are excluded until bug A is fixed.
6. **No first strike anywhere.** Combat damage entries are untagged,
   including a vanilla 5/5 blocked by two 2/2s whose damage lands when
   the assignment prompt is answered. The helper schedules nothing, and
   the frame renders with no added delay (asserted in the pure-helper
   tests).
7. **Reduced motion.** Helper half: `beatMode` returns `"still"` when
   `reduceMotion` is on, animations are off, or `damagePopups` is off.
   A still schedule has no pulse or ghost steps, still carries the
   "First strike" cue text, and when both beats are in one frame puts
   beat 2's cue one scaled pause after beat 1's (beat 1 is held, not
   collapsed). Shell half (no tween starts; the cue text is in the
   accessibility tree): checked by hand until
   [#689](https://github.com/krakenhavoc/cmd_and_ctrl/issues/689) gives
   the client a DOM test environment.
8. **Reconnect.** A client whose first frame lands mid-`combat_damage`
   primes: no beats replay.
9. **Prevention.** When every point of first-strike damage is
   prevented, beat 1 is skipped (Decision 3).
10. **Beats across frames.** Given two frames from one `combat_damage`
    step, where frame N has `first_strike` entries and frame N+1 has
    more `first_strike` entries followed by `regular` entries: the
    frame N+1 `first_strike` entries play as a continuation with no
    pause and no repeated label, then the pause, then beat 2. Given a
    frame N+1 with only `regular` entries: beat 2 plays with no added
    pause.
11. **Undo.** After a frame whose highest log `seq` is lower than the
    highest seen, entries replayed at the reused `seq` numbers are cued.
12. **Timing.** With both beats in one frame, beat 2 is scheduled at
    400 ms at `speed = 1`, 800 ms at `speed = 2` and 200 ms at
    `speed = 0.5` after beat 1, in both `"full"` and `"still"` mode.
13. **Ghost arrows.** `arrowRender` returns `"live"` when the arrow ID
    is live, `"ghost"` when it is not live but cached, and `"none"`
    when it is neither. For criteria 1–3, the arrow IDs the beats name
    come back `"ghost"` given the previous frame's cache and the end
    state's live set. Drawing the ghost, and dropping it after a board
    resize, is checked by hand until #689.
14. **Leaving the step mid-sequence.** A frame whose step is `end_combat`,
    arriving before beat 2's scheduled time, does not remove beat 2
    from the schedule.

## Sub-PR plan

**Sub-PR 1 — server: tag, projection, tests. After #709 merges.**
- Add `game.Event.CombatStep`, `damageTail.combatStep` and
  `DamageAssignmentFrame.CombatStep`. Stamp in `emitDealDamageLocked`,
  and fill the tail in `combatDamageTailLocked` (from an argument) and
  `damageTailFromFrame` (from the frame). Tag only when pass 1 ran.
- Add `protocol.LogEvent.CombatStep`, the projection in
  `log.go:417-428`, and the `Text` suffix.
- Update `docs/protocol.md` (the LogEvent line at `:419`) and the
  `protocol.ts` mirror type (type only, no behaviour).
- Tests:
  - Game-package tests for the tags on each path: first strike vs
    Bears, double strike vs Bears, double strike unblocked, vanilla-only
    (untagged).
  - Prompt-resume tagging through `DamageAssignmentFrame.CombatStep`:
    a regular-pass prompt with the first-strike pass run (`regular`),
    and a vanilla 5/5 blocked by two 2/2s (untagged). A first-strike
    pass prompt is tested only for its tag, not for its order, until
    #702 lands.
  - A CR 616 ordering prompt in the first-strike pass comes back tagged
    `first_strike`.
  - Log projection and text.
  - JSON round trip of an old snapshot event with no tag, and of an old
    pending damage-assignment prompt with no `CombatStep`.
- Checks: `go test` for `./internal/game/...` and `./internal/protocol/...`,
  plus `golangci-lint`.

**Sub-PR 2 — client: pure beat helper and a thin shell.**
- `client/src/lib/combatBeats.ts`: new-entry tracking with priming and
  rewind, step keying, beat membership and continuations, arrow IDs from
  the log, `arrowRender`, `beatMode` (including `damagePopups`), and a
  schedule that is a function of mode and `animations.speed`
  (`BEAT_PAUSE_MS = 400`).
- `combatBeats.test.ts` (vitest) covers acceptance criteria 1–6, 8–12
  and 14 as data-in/data-out cases, plus the helper halves of 7 and 13.
- The shell: `CombatArrows.svelte` gains the arrow geometry cache and a
  ghost arrow style, takes the arrow IDs to cue, pulses live arrows and
  draws ghosts as `arrowRender` says. It keeps running the schedule
  after the step changes. The board renders the text cue.
- No DOM render test until #689. Until then the shell stays thin enough
  that the helper tests carry the logic, and the shell halves of
  acceptance criteria 7 and 13 are checked by hand.
- Check: `npm run test`, `npm run check`.

**No sub-PR 3.** The real second combat damage step is not part of #187
([Decision 5](#decision-5--presentation-only-the-second-combat-damage-step-is-717)).
It is [#717](https://github.com/krakenhavoc/cmd_and_ctrl/issues/717) and
needs its own ADR before any code.

`docs/sprints.md`'s S18 line for #187, which repeated the issue's
premise and the wrong 510.2 / 510.3 citation, was updated to point here
in the commit that accepted this ADR.

## Addendum (sub-PR 2): client implementation notes

Sub-PR 2 follows the plan above. `client/src/lib/combatBeats.ts` holds
every rule (`track`, `splitBeats`, `arrowRefsFor` / `arrowIDsFor`,
`arrowRender` and `resolveArrow`, `beatMode`, `schedule`, plus
`planFrame`, which composes them, a `BeatSequencer` timer class and
the `BeatDirector` that owns both), and `CombatArrows.svelte` is the
shell. Choices the ADR left open, and the places the code goes past it:

- **The shell is `CombatArrows.svelte` alone.** It also renders the
  text cue, as HTML outside its `aria-hidden` SVG, in a
  `role="status" aria-live="polite"` region that is always mounted.
  There is one cue per `combat_damage` step, placed at the curve
  midpoint of the arrows its first beat drew (or the board centre when
  none could be drawn), and "Regular damage" joins "First strike" in
  the same cue. Every layer is `pointer-events: none`. `Board.svelte`
  only forwards a prop.
- **The announcement is terse.** The live region announces the label
  and a count ("First strike, 2 hits"), not the beat's log lines. It
  speaks on every beat of every combat, and the log panel already has
  the sentences.
- **Seen `seq`s are a watermark, not a set.** The log is appended in
  `seq` order, so "every `seq` at or below the highest handled" is the
  same set. A rewind lowers the watermark, and it also forgets a cued
  beat whose first entry is above the new watermark, so an undo inside
  the step cues the replay with its label.
- **Re-priming, not just first-frame priming.** An automatic reconnect
  does not unmount the board (`ws.ts` keeps the stale snapshot
  rendered), and neither does the dev replay scrubber. Game.svelte
  passes a `beatsPrimeKey` (connection status and replay toggle) down
  through `Board`, and a change primes the next frame. Without it, the
  combat damage missed during a disconnect, or everything between a
  replay frame and the live game, would be cued as live.
- **A priming frame cancels pending cues.** This is the one exception to
  "a later frame never cancels a cue". A frame that primes (the first,
  or a re-prime) makes `BeatDirector` dispose the sequencer, start a
  new one, and clear the on-screen cues, effects and caches. Those
  cues belong to frames the client is no longer showing. On a replay
  scrubber jump mid-sequence they would otherwise fire against the new
  frame and flash a stale label.
- **Two beats that are both new in one frame, with regular first**
  (#702's order today, when both land in one frame) are both cued at
  once. Nothing is reordered, and no pause is added because beat 1 did
  not come first.
- **Durations.** A pulse or ghost is `BEAT_EFFECT_MS = 360` ×
  `animations.speed`, fading in and out. It is clamped below
  `BEAT_PAUSE_MS`, and a unit test pins that margin, so the effect is
  over before the next beat. The text cue stays up
  `BEAT_CUE_HOLD_MS = 1800` after a frame's last beat for that step.
  That hold is **not** speed-scaled: it is text to read, not motion,
  and speed 0.5 must not make it unreadable.
- **Motion is checked twice.** Once when the frame is planned, and
  again when each cue plays (from the settings store), so turning
  reduce motion on mid-sequence starts no further tween.
- **Ghost endpoints.** "Still on the board" means the card is in
  `view.battlefield.cards`. A dead creature's instance can still be
  drawn in a graveyard pile, and a ghost must not point there. "The
  board's size changed" allows 1 px of tolerance. The ghost style is a
  thin dashed stroke with no glow and no arrowhead. A pulse is a wide,
  soft stroke over the live arrow.
- **Card-tile fallback for arrows that were never drawn.** Arrows are
  measured in a `requestAnimationFrame`. A browser check on a local
  stack found the declare-blockers frame and the combat damage frame
  reaching the defender 4–5 ms apart (3 of 3 runs, and the attacker
  in one run). Both landed inside one animation frame, so the block
  arrow was never measured, the arrow cache had nothing for it, and
  only the text cue played. That is blocked first strike, the main
  case.
  - **The fix.** `CombatArrows` also caches every battlefield tile's
    board-relative centre, by instance ID, on every measurement. A tile
    is measured on every frame it is on the board, long before combat.
    When no arrow geometry is usable, `resolveArrow` builds the ghost
    from the two endpoints. Each endpoint is freshly measured if it is
    still on the battlefield, and otherwise taken from its cached tile
    (a seat header is always mounted).
  - **Reflow rule.** A cached tile measured on a board of another size
    is never used, so the ghost is dropped and the text cue carries the
    beat.
  - **Why not only `$effect.pre`.** Measuring the outgoing DOM
    synchronously before each update was rejected as the sole fix: two
    frames can still be batched before any render, and the cache covers
    that case too.
- **Cache lifetime.** Kept while the step is `declare_attackers`,
  `declare_blockers` or `combat_damage`, or while any cue is pending
  (`keepArrowCache`). Cleared otherwise, and on every prime. For tiles,
  "cleared" means the tiles of cards no longer on the battlefield; tiles
  still there are re-measured on every frame (`pruneCardCache`).
  - **What "pending" means.** A cue counts as pending while it is
    **playing**, not only while it is waiting. The sequencer releases
    its timer after `onCue` returns, and the shell prunes the caches
    in a microtask after the cue. A browser re-check found the
    ordering bug this avoids. In the auto-pass flow, the last cue plays
    after the step has reached `postcombat_main`. Measuring at cue
    time pruned the dead creatures' tiles in the same millisecond,
    before the ghost was resolved, so beat 2 drew nothing (7 of 8
    runs).

## Dependencies

- **[#709](https://github.com/krakenhavoc/cmd_and_ctrl/pull/709)** (open,
  fixes [#694](https://github.com/krakenhavoc/cmd_and_ctrl/issues/694)):
  one damage tail and one emit site. Sub-PR 1 builds on it (Decision 1).
- **[#702](https://github.com/krakenhavoc/cmd_and_ctrl/issues/702)**:
  the regular substep resolves while a first-strike multi-blocker
  assignment prompt is still pending (`mutations.go:4439` runs after
  `:4624` queues the prompt). Today the tags on those events are
  correct but their `seq` order is wrong: first-strike damage lands
  after regular damage. #702's fix direction ("suspend combat damage
  resolution and resume substep 2 from the prompt's continuation") also
  changes **how the beats split across frames**, not only their order:
  beat 1 and beat 2 then routinely arrive in different frames. The
  [Beats across frames](#beats-across-frames) rules cover that, so
  extending acceptance criterion 5 to multi-blocker combat needs no
  sequencer redesign.
- **Bug A, [#715](https://github.com/krakenhavoc/cmd_and_ctrl/issues/715)**: an attacker whose blockers are all gone is
  treated as unblocked, and a menace block is reverted when one of two
  blockers is gone ([Context](#bug-a-the-engine-forgets-that-an-attacker-was-blocked)).
  Acceptance criterion 2 depends on it, and criterion 5 excludes it.
  Related: #672.
- **Bug B, [#716](https://github.com/krakenhavoc/cmd_and_ctrl/issues/716)**: see the last bullet under Consequences.
  Nothing here depends on it.
- **[#703](https://github.com/krakenhavoc/cmd_and_ctrl/issues/703)**:
  the life popup shows only the last delta per frame, and nothing once
  `life_history` reaches `MaxLifeHistoryEntries` (`player.go:20, 244-248`).
  Related to acceptance criterion 4, but not a dependency of it.
- **[#717](https://github.com/krakenhavoc/cmd_and_ctrl/issues/717)**:
  modelling the real second combat damage step, with priority. A
  follow-up, not a dependency: #187 does not wait for it (Decision 5),
  and it may make most of sub-PR 2's sequencer unnecessary.
- **[#689](https://github.com/krakenhavoc/cmd_and_ctrl/issues/689)**:
  the Svelte DOM test environment. It does not block this work. It is
  why sub-PR 2 is a pure helper plus a thin shell, and why the shell
  half of acceptance criterion 7 is a manual check.
- **[#693](https://github.com/krakenhavoc/cmd_and_ctrl/issues/693)**
  Part 2 (the 510.2 / 510.3 citation fix). It is independent, but any
  comment sub-PR 1 touches in `resolveCombatDamageLocked` should cite
  510.4.

## Consequences

- The substep becomes a fact the server states. The log panel, bug
  reports, replays and the bot tiers all get it without new client
  logic.
- `game.Event` and `DamageAssignmentFrame` each gain a persisted field.
  Both are additive, and their zero value means "untagged", which is
  correct for older files, so under the rule at
  `server/internal/game/snapshot.go:70-86` there is no
  `SnapshotSchemaVersion` bump.
- The board is never held. Each frame renders its end state at once, as
  today, and input stays live against it. Only the overlay cues (arrow
  pulses and the text cue) are sequenced, and they trail the board by up
  to one pause. The dying creature is already gone when its cue plays,
  so its arrow is a ghost drawn from cached geometry (Decision 4).
  Cues also finish after the game leaves the step, so they can trail
  the board past `end_combat`.
- A ghost arrow can briefly point at a spot another card has slid into
  after the reflow. Decision 4's mitigations (a distinct style, a short
  fade, dropping the ghost after a resize) limit this, and the text cue
  always carries the beat.
- #187 is not the rules fix for two combat damage steps. Until
  [#717](https://github.com/krakenhavoc/cmd_and_ctrl/issues/717) lands,
  first-strike damage triggers still resolve after regular damage, and
  no one can respond between the steps. When #717 lands, much of the
  client sequencer may be removed.
- Bug B, found by reading the code and not yet tested:
  `participatesInSubstep` reads keywords **as they are now**, after the
  layer recompute between passes. CR 510.4 and 702.7c say the regular
  step includes creatures that had neither first strike nor double
  strike *as the first combat damage step began*. A creature whose
  granted first strike leaves between passes (the lord died in pass 1)
  would deal damage a second time. The tag would faithfully label that
  wrong event `regular`. It is separate from #187. Filed as
  [#716](https://github.com/krakenhavoc/cmd_and_ctrl/issues/716).

## Deferred

- Making the `CombatArrows` draw-on honour settings. It is a raw
  `gsap.fromTo` with a fixed duration (`:235-247`) and reads no settings
  at all, so it would have to call `gatedDuration` (or check the
  settings store) itself. Separately, bridging
  `accessibility.reduceMotion` into `animations.ts` would make every
  existing `gatedDuration` caller honour it. Both change animations
  outside #187, and each is worth an issue of its own.
- Death out-animations in general (`animations.ts:152-154`). This ADR
  only needs the cached-geometry ghost arrows for combat (Decision 4).
- A damage cue on the creature itself: a floating damage number on the
  tile, or a flash of the `damage_marked` badge. `damagePopups` today
  drives only the player life popup, so nothing here follows from it.
  If it is wanted, it should reuse `damagePopups` rather than add a
  toggle, and it needs the same answer for a creature that has left
  (a tile to float over).
- Damage-assignment prompt animation (out of scope in #187).

## Alternatives considered

- **Ephemeral `GameView.CombatDamageBeats` sidecar** (the #187
  proposal). Rejected under Decision 1: it is lost on reconnect and
  dropped frames, needs an ack signal the protocol lacks, and needs a
  second filtered channel.
- **A marker event between the passes** (`EventCombatDamageStep`) instead
  of a field on each damage event. The client would still have to
  infer each entry's substep from its position. A marker adds a log
  entry kind that every log consumer must skip. And prompt-resume
  damage lands in a later frame, after the marker for the next pass,
  so position would be wrong there too (#702). A per-event field has
  none of these problems.
- **Tag every combat and let the client ignore regular-only steps.**
  Rejected under Decision 1.
- **Client-only keyword heuristic.** Rejected under Decision 1.
- **Collapse to the end state under reduced motion.** Rejected by the
  owner (Decision 2): the information must survive.
- **Show both cues at once in `"still"` mode.** Rejected by the owner
  (Decision 2, Timing): holding beat 1's cue keeps the order of the two
  steps readable without motion.
- **A fixed pause that ignores `animations.speed`.** Rejected by the
  owner (Decision 2, Timing): every other effect scales with speed.
- **A new `animations.combatBeats` toggle.** Rejected by the owner
  (Decision 2): beats reuse `damagePopups`.
- **Ghost tiles** (option b for a creature that died). Keep a dying
  card's tile mounted until its beat ends, so its arrow has a real DOM
  node. Rejected for cached geometry (Decision 4): it touches
  `BattlefieldRow`, overlaps the deferred death out-animation, and a
  ghost tile would have to refuse input (targeting, context menu, drag)
  for a card no longer on the battlefield.
- **Text only for a creature that died** (option c). No cost, but
  blocked first strike that kills its blocker, the main case, would get
  no per-pair visual. Rejected for cached geometry (Decision 4).
- **Cut the remaining cues when the game leaves the step.** Rejected by
  the owner (Beats across frames): the cues are overlays that never
  hold the board, and cached geometry lets them finish.
- **Model the second combat damage step inside #187.** Rejected by the
  owner (Decision 5): it is a turn-structure change, filed as #717.
