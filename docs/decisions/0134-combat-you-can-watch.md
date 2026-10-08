# ADR 0134 — Combat you can watch: attackers lunge, hits land

**Status:** Accepted (owner answers 2026-10-07) · 2026-10-07 · S69 — Combat you can watch: table animations (issue [#2614](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2614))
**Issues:** [#2614](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2614) (this change).
**Owner decisions:** on 2026-10-07 the owner chose the recommended option, (a), on all eleven questions. The answers are listed under [Owner decisions](#owner-decisions-2026-10-07) and are binding. The options not chosen are kept under [Questions for the owner (answered)](#questions-for-the-owner-answered). Smaller calls the questions do not cover are under [Calls made here](#calls-made-here), so the owner can overturn any of them in review.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-07. I ran `git fetch --all --prune` and listed `docs/decisions/` on every remote head: `origin/develop`, `origin/main`, `origin/cost-ledger`, `origin/docs/issue-audit`, `origin/feat/750-conditional-block-restrictions`, `origin/feat/cost-tracker`, `origin/feat/playmats`, `origin/fix/2608-harvest-battlefield-mutation`, `origin/fix/caddy-reload-admin-off`, `origin/fix/themberchaud-sweep`, `origin/wip/836-one-click-default` and `pr/2326`. The highest number on any of them is 0133 (`0133-opening-hand-actions.md`, on `origin/develop`). None of the four open pull requests adds an ADR. This ADR takes **0134**.
**Amends:** [ADR 0053](0053-combat-damage-beats.md) in two small ways (§1). Combat damage with no first-strike step gets a beat with no label, so it has motion and still has no text cue. The same-frame pause between the two beats grows to fit a lunge when combat motion is on.
**Builds on:** [ADR 0053](0053-combat-damage-beats.md) (the beat tracker, priming, rewind, finishing the cues, and cached geometry for creatures that died), [ADR 0119](0119-a-stack-you-can-follow.md) §2 and §3 (the stack hold, and the linger's copy in a layer over the board), [ADR 0120](0120-expand-a-players-board.md) §3 (`boardAnchor` and the z ladder), [ADR 0121](0121-animated-dice.md) §7 (an animation that waits for nothing, with a per-effect toggle), and [ADR 0075](0075-table-settings-and-host-controls.md) §2.2 (bot pace).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

Combat damage reaches the table as a jump cut today. A frame arrives, the life totals and damage badges change, dead creatures vanish, and a log line says what happened. In a four-player game, nobody watching sees which creature hit what. #2614 asks for combat you can watch: an attacker lunges toward its blocker, or toward the defending player's avatar when unblocked, then snaps back, and the hit lands with a small shake next to the existing damage popup.

**Constraint (owner, on #2614):** the table is never played below tablet size, so every attacker, blocker and avatar is on screen. This ADR designs no phone-width fallback.

Every claim below was checked on `origin/develop` at `280e5309`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`).

### Owner decisions (2026-10-07)

The owner chose (a), the recommended option, on every question:

1. **Trigger:** the combat damage log entries in a new frame, through ADR 0053's beat tracker (§1).
2. **What moves:** an art-only copy in a board-level layer, with the live tile hidden during the flight (§2).
3. **Distance:** to contact, clamped to between ½ and 0.85 of the centre-to-centre distance (§2).
4. **Several blockers:** one lunge at the blockers' centroid, and they all shake together (§3).
5. **Trample:** the player's avatar shakes at the same contact, and PR 2 adds a streak to it (§3).
6. **Blockers:** a blocker lunges only in a beat where its attacker dealt it no damage, and otherwise braces (§3).
7. **Deaths:** the death plays after the hit. An attacker that died fades as it snaps back, and a blocker that died shakes, then fades (§3).
8. **Pacing:** display only, so nothing waits for the animation (§4).
9. **Setting:** a new `animations.combat` toggle, default on, under the master switch and reduced motion, scaled by speed (§5).
10. **Replays:** strikes play only on a one-frame step forward, and any other jump re-primes (§6).
11. **Sound:** in PR 3, `combat_resolve` moves to the first contact of each beat, with no new sound files (§7).

### The rules

- **CR 510.1:** the active player announces how each attacking creature assigns its combat damage, then the defending player does the same for each blocker. **CR 510.1c:** when two or more creatures block an attacker, its damage is "divided as its controller chooses among them". This edition has no damage assignment order, so there is no order in which an attacker hits its blockers.
- **CR 510.2:** "all combat damage that's been assigned is dealt simultaneously." So one step's hits are one moment, not a sequence.
- **CR 510.4** (and 702.7b): if a creature has first strike or double strike as the step begins, the phase gets a second combat damage step. A double striker deals damage in both.
- **CR 702.19b:** trample assigns lethal damage to the blockers, and the rest may go to the player, planeswalker or battle being attacked. It is dealt in the same moment as the rest (CR 510.2).
- **CR 704.3:** state-based actions are checked when a player would get priority, so a creature dealt lethal damage dies after the damage, not during it (CR 704.5g).
- **CR 506.2:** during combat, the active player is the attacking player.

### What exists

**The beat pipeline (ADR 0053).** `client/src/lib/combatBeats.ts` already turns the public log into combat damage cues. It covers almost everything an animation needs:
- `track` (`:136`) handles the hard cases. It primes on the first frame and after a reconnect or replay toggle, so nothing already on the wire plays. It treats a frame whose highest `seq` is lower as an undo rewind, so nothing plays then either. Only entries with a higher `seq` are new.
- `splitBeats` (`:208`) groups a frame's new combat damage entries into a first-strike beat and a regular beat for each combat, including beats split across frames by a damage-assignment or CR 616 prompt.
- `BeatSequencer` and `BeatDirector` (`:739`, `:800`) play the cues on timers. A later frame only adds cues, and a priming frame cancels them all.
- A same-frame pause, `BEAT_PAUSE_MS` = 400 ms × `animations.speed` (`:55`), separates beat 2 from beat 1 when both arrive in one frame.
- `arrowRefsFor` (`:373`) works out who blocked whom from the same combat's `block` entries.
- `cardCache` in `CombatArrows.svelte` keeps every battlefield tile's last measured centre, so a creature that died can still be pointed at.

One gap matters here. `combatTagOf` (`:199`) only accepts entries tagged `first_strike` or `regular`, and the tag is set only when the combat had a first-strike step. Ordinary combat, the common case, produces **no cue at all** (`combatBeats.test.ts`, "schedules nothing for untagged combat damage", and AC6 at `:612`).

**The damage entry carries what an animation needs.** `EventDealDamage` is projected in `server/internal/protocol/log.go:1210`. A combat `LogDamage` entry has:
- `card_id`: the source creature.
- `target`: a target permanent (a blocker, an attacker, a planeswalker or a battle), or `target_seat` for a player (`setTarget`).
- `amount`: always above 0. An entry for 0 damage is not emitted.
- `combat: true`.
- `combat_step`, on a combat with a first-strike step.
- `seat`: the dealing creature's controller, captured when the damage is dealt (`base.Seat = seatOf(ev.Actor)`, `:1038`; `damage_tail.go:957`).

`LogStep` entries carry the active seat (`game.go:1827`). So an entry is from the attacking side exactly when its `seat` equals its step entry's seat (CR 506.2), even when the `block` entries have fallen out of the 200-entry window. **No wire field is needed.** Had one been needed, it would have been a `combat_role` (`attacker` or `blocker`) on combat `LogDamage` entries. The step entry's seat makes it redundant.

**How frames arrive.** Every successful commit captures one frame (`ws/room.go:303`, `apply`), and the hub sends every frame to every client. It never drops one: a client whose send buffer is full is disconnected instead (`hub.go:728`), and its reconnect primes. A combat's damage arrives in **one frame**, except in three cases:
- With a first-strike step, the two steps are real steps since #717. They usually arrive in two frames, with a priority window between them. They arrive in one frame when the table auto-passes through.
- A CR 510.1c division prompt or a CR 616 ordering prompt holds part of a step's damage for a later frame (ADR 0053, "Beats across frames").
- ADR 0127's automatic answers capture a frame for each answer, but return only the last to the caller (`autoAnswerThenCaptureLocked`). One broadcast can therefore carry several commits.

Each of these is one more reason to trigger from the log rather than from a diff of two views.

**What the board looks like when a frame lands.** The board renders the frame's end state at once. A creature that died is already gone from its row, and the row has closed up. Its damage badge, the defending player's life total and the popup have already changed.

**Overlays already drawn across boards.**
- `CombatArrows.svelte`: SVG at z 35 and 36, positioned from `getBoundingClientRect` through `boardAnchor`.
- `TargetingArrows.svelte`.
- `StackLinger.svelte`: a copy of a card drawn where it was, then flown to its landing with `flyTo` (`animations.ts`), z 41.
- `DiceLayer.svelte`: z 41.

All four are `aria-hidden`, take no pointer events, and are mounted in `Board.svelte` (`:3183` onward).

**Why the attacker cannot move in place.** `Card.svelte` composes its transform from CSS variables (`:1288`, `rotate(var(--tap-rot)) … scale(var(--hover-scale))`), so a small effect such as a shake can be added beside tapping. A lunge across the table cannot use it:
- Each row scrolls (`BattlefieldRow.svelte:409`, `overflow: auto`).
- Each seat's panel clips (`PlayerPanel.svelte`, `overflow: hidden`) and is its own stacking context (`isolation: isolate`, `:991`).
- A tile moved three-quarters of the way to another seat is cut off at its own panel's edge, and paints under the other panels.
- A creature that died has no tile left to move.

The issue's note that a lunge "can sit alongside tapping" holds for the impact shake. It does not hold for the lunge itself.

**Animation plumbing.** `animations.ts` holds an `AnimationConfig` with a master `enabled`, a `speed` of 0.5, 1, 1.5 or 2, and one flag per effect. `gatedDuration` snaps an animation to 1 ms when it is off. The settings bridge pushes the `animations` group (`settings.ts:1143`), and `animations.enabled` starts off when the OS asks for reduced motion. `accessibility.reduceMotion` is a separate setting that the bridge does not push, so motion code reads it from the store, as `beatMode` does.

**Sound.** `sounds.ts` has `attack`, `block`, `combat_resolve` and `damage`.
- `combat_resolve` plays when the step changes to `first_strike_damage` or `combat_damage` (`Game.svelte:930`).
- `damage` and `heal` play with the life popup (`lifePopup.ts:241`, `PlayerIdentity.svelte:210`), at the moment the frame arrives.

**Pacing.**
- The bot presets (`aiseat/runner.go:143`) are fast {MinThink 0, StackHold 0}, normal {700 ms, 2 s} and slow {2 s, 3 s}. The runner re-reads the table's pace before every decision.
- ADR 0119 §2's stack hold delays only an automatic pass on another seat's top stack item.
- Nothing in the client or server waits for an animation. ADR 0053, ADR 0119 §3 and ADR 0121 §7 all chose display only.

---

## Decision

### 1. What starts it: the combat damage entries in a new frame

A strike is cued by the combat `LogDamage` entries that are new in a frame. It reuses ADR 0053's tracker, so priming, rewinds, reconnects and replay toggles behave exactly as the beat cues do (question 1).

Two changes to `combatBeats.ts`:

- **Ordinary combat gets a beat.** An untagged combat damage entry joins a beat tagged `regular`, marked `labelled: false`. `schedule` gives it `label: null`, so no text cue is drawn and nothing is announced. With `combat` motion off, the beat is dropped. So for a player with combat motion off, ADR 0053 Decision 3 and AC6 are unchanged: no text cue, no pulse. The tests that pin "schedules nothing for untagged combat damage" are rewritten to pin "schedules no label and no pulse".
- **The pause fits the lunge.** `beatPauseMs(combatMotion)` replaces the constant. It is 400 ms as today with combat motion off, and `STRIKE_MS + 60` = 560 ms with it on, both × `speed`. Beat 2 never starts before beat 1's attackers are back. `BEAT_EFFECT_MS` keeps its clamp against the smaller pause.

**One director, two layers.** The `BeatDirector` moves out of `CombatArrows.svelte` into a small shared object, `combatCues.svelte.ts`. `Board.svelte` creates it, and both `CombatArrows` and the new `CombatStrikes` layer subscribe to it, as `DiceQueue` is shared (ADR 0121). One tracker and one clock mean the beat label and the lunge it describes can never disagree.

**Rejected: diffing the view.** `damage_marked`, life totals and battlefield membership do not say who dealt the damage, or whether it was combat damage. A creature that died has no view to diff. A double strike to a player lands as one life change when both steps share a frame (the reason `lifePopup.ts` stopped diffing lengths). A frame that carries several commits merges them.

### 2. The motion

**A copy moves, in a board-level layer (question 2).** A new `CombatStrikes.svelte`, mounted in `Board.svelte` beside `CombatArrows`, draws each moving creature as an **art-only copy**:
- The copy uses the same image the tile shows: `cardImageURL`, or the card back for a face-down permanent (`showsCardBack`). It has the tile's size and its tap rotation, and is drawn where the tile is.
- The live tile is hidden for the copy's flight only. `Card.svelte` reads a `striking` set from the cue object and sets `visibility: hidden`.
- The layer sits at **z 37**: above the arrows (35, 36), below the pile (38), the attention strip (40), the linger and the dice (41) and the dock (55). A lunge never covers the stack, a die or a control.
- It is `aria-hidden` and takes no pointer events.
- Endpoints are found through `boardAnchor`, so with a board expanded (ADR 0120) the lunge starts and ends on the overlay's copies.

**Where it starts and aims.**
- A copy starts from the live tile's rect, or from the last rect cached for a creature that has died. The layer keeps the last frame's `CardView` and the last rect for every battlefield card, with a size and a rotation, which extends `cardCache`.
- A copy aims at its target's centre: a blocker's tile (live or cached), a planeswalker's or battle's tile, or a player's avatar. When one creature hits several targets in a beat (§3), it aims at the centroid of their centres.
- A cached rect from a board of a different size is never used (ADR 0053's reflow rule). That strike is skipped, and the log, the badges and the popup still carry the hit.

**Distance: to contact (question 3).** The copy travels until its leading edge meets the target's edge, plus 8 px of overlap so the two visibly touch. The travel is clamped to between ½ and 0.85 of the centre-to-centre distance. For each box, the half-extent along the direction of travel is `|dx|/d · w/2 + |dy|/d · h/2`. On a tablet table this lands inside the owner's suggested ½ to ¾ for most pairs. It never stops short of a far target and never buries a near one.

**Timing at speed 1** (`STRIKE_*` in `animations.ts`, each × `speed`):

| Phase | ms | Easing |
|---|---|---|
| Out | 180 | `power2.in`: it accelerates into the hit |
| Contact hold | 60 | none: the impact plays here |
| Back | 260 | `power3.out`: a quick snap that settles |
| **Total** (`STRIKE_MS`) | **500** | |

There is no wind-up. A wind-up adds time before the hit, and the life popup is already up at the frame's arrival (see [Calls made here](#calls-made-here)).

**The impact.**
- **A creature, planeswalker or battle that is hit:** the live tile shakes horizontally, ±4 px over about 160 ms (three half-cycles, `sine.inOut`). A brightness flash runs with it, 1.5 back to 1. This uses a new `--impact-x` variable composed into `Card.svelte`'s transform before the rotation, and a `--impact-glow` filter. Both default to no-ops, so a tile at rest renders exactly as it does today. The existing damage badge and its lethal styling carry the number.
- **A player that is hit:** the avatar disc in `PlayerIdentity.svelte` shakes the same way, through the same variables on the disc. The existing popup carries the number.
- **The attacker, when its blocker hits it back in the same beat:** its copy flashes at contact, with no shake, so the two motions do not fight.

### 3. Sequencing

| Case | What plays |
|---|---|
| **No first strike** | One beat: every attacker that dealt damage lunges at once (CR 510.2), and every target shakes at contact. |
| **First strike or double strike** | One beat per damage step (CR 510.4), from ADR 0053's beats. Steps in different frames play as their frames arrive. Steps in one frame are separated by `beatPauseMs`. A double striker lunges in both. |
| **Several blockers (question 4)** | The attacker lunges **once**, at the centroid of the blockers it damaged, and they all shake together. CR 510.1c has no order, and CR 510.2 deals the damage at once, so a lunge per blocker would show a sequence the rules do not have. |
| **Trample (question 5)** | The attacker lunges at its blockers. The defending player's avatar, planeswalker or battle shakes **at the same contact**, because it is the same moment. PR 2 adds a short streak from the blockers to the avatar, drawn during the contact hold. |
| **Planeswalkers and battles** | An unblocked attacker on a planeswalker or a battle lunges at that permanent's tile, from the `target` card ID, and the tile shakes. Its loyalty or defence change shows as today. |
| **Blockers (question 6)** | A blocker lunges only in a beat where the attacker it blocks dealt **it** no damage: a first-strike blocker, or a blocker facing a 0-power attacker. Otherwise it braces, shaking at the attacker's contact, while the attacker's copy flashes. |
| **Who is attacking** | A source is on the attacking side when its entry's `seat` equals the seat on its damage step's `LogStep` entry (CR 506.2). The `block` entries name what a blocker blocked. |
| **Deaths (question 7)** | A creature that dies plays its death after the hit (CR 704.3). An attacker that died fades out on its way back (`opacity` 1 → 0 over the back phase). A blocker that died shakes, then fades over 240 ms in its cached place. Both are drawn as copies from the cache, since their tiles are already gone. A creature counts as dead when the beat's own entries include its zone change off the battlefield (ADR 0053's beat membership already puts the SBA deaths in the beat). |
| **No damage, no lunge** | A creature that dealt no damage has no entry, so it does not move. That covers a blocked attacker with no blockers left (CR 510.1c), 0 power (CR 510.1a), and damage prevented to 0. The arrows already showed it attacking. |

**Many attackers.** CR 510.2 makes a beat simultaneous, so every lunge in a beat starts together. A beat moves at most **12 copies**, chosen in log order. Past that, the remaining attackers do not lunge, but their targets still shake. A token swarm in Commander can be thirty creatures, and thirty full-size images in flight is the wrong cost for the effect.

### 4. Pacing: display only

The animation holds nothing (question 8):
- the server never delays a broadcast;
- the bots do not wait for it;
- the client's auto-pass does not wait for it;
- input stays live under it.

This is the rule of ADR 0053, ADR 0119 §3 and ADR 0121 §7. A frame that moves the board on while strikes are playing does not cut them. They finish against cached geometry, as ADR 0053's cues do.

- **Bots.** At `normal` pace, a bot's next move waits at least MinThink, 700 ms, which is longer than a 500 ms strike. So at the default pace the next frame usually lands after the attackers are back. At `fast` the board may move on under the strike, and the strike finishes as an overlay. No `CombatHold` is added to the presets.
- **The stack hold (ADR 0119 §2).** It is unchanged and independent. "Deals combat damage" triggers go on the stack in the same frame as the damage. The pile shows them, and the hold delays auto-passing them, while the strikes play on the board. The two are in different places, and the hold (2 s by default) outlasts a strike.
- **A strike that would start late is dropped.**
  - If the page is hidden (`document.visibilityState !== "visible"`) when a frame arrives, that frame's strikes are not scheduled. GSAP runs on `requestAnimationFrame`, which a hidden tab does not run, so they would otherwise all play on return, against a board that has moved on.
  - A strike whose start comes more than 1 s after its frame arrived is dropped.

  ADR 0053's text cues are not affected by either rule.

### 5. Settings

A new per-effect toggle, **`animations.combat`** (question 9):
- Default on, under the master `enabled` switch.
- Labelled in Settings → Animations "Combat: attackers lunge and hits land".
- Synced like the other `animations` fields (`settings.ts:564`), and filled by the settings merge, so it needs no migration.
- Pushed to `animations.ts` by the bridge.

`combatMotion(s)` is true only when all of these hold: `animations.enabled`, `animations.combat`, and not `accessibility.reduceMotion`. The strike layer reads it again at cue time, so turning reduced motion on mid-combat starts no further tween, as `CombatArrows.playCue` does.

- **Speed:** `animations.speed` scales every `STRIKE_*` phase, the shake, the death fade and `beatPauseMs`. Nothing in this ADR is reading time, so nothing is exempt.
- **Off or reduced motion:** no copy, no lunge, no shake, no flash, and the beat pause stays at 400 ms. What remains is what exists today: the damage badges, the life popup, the log line, and ADR 0053's text cue on a combat with first strike. That already carries every number. Combat motion off does not turn off ADR 0053's arrow pulses, which stay under `damagePopups`.

### 6. Who sees it

- **Every seat and every spectator** sees the same strikes. They come from the public log, which every viewer receives. A face-down creature's copy shows the card back, as its tile does.
- **Bots and the MCP seat** see nothing; they have no board.
- **Reconnects and restores.** A reconnect, or a deploy restore followed by a reconnect, primes through `beatsPrimeKey`, so nothing stale plays.
- **Undo.** An undo rewinds and plays nothing. The redo's damage plays as new, which is what happened at the table.
- **Replays (question 10).** The dev scrubber (ADR 0023) plays strikes only when it steps **exactly one frame forward**. Any other jump re-primes, so scrubbing past three combats does not fire them all at once. Today the prime key changes only when the replay is toggled. PR 2 adds the replay index to it whenever a jump is not a single step forward. That also fixes the same burst for ADR 0053's cues, the dice and the linger, which share the key.

### 7. Sound: later, and no new files

PR 1 changes no sound (question 11).
- PR 3 moves the `combat_resolve` cue from the step change to the **first contact of each beat** when combat motion is on. With motion off it stays on the step change.
- The life popup's `damage` sound keeps firing with the popup.
- No new sound files are added: a new sample brings a licence question, and the existing two takes per event already cover a hit.

### 8. Tests and contracts

**Vitest (PR 1).**
- `combatBeats.test.ts`:
  - untagged combat damage gives one `regular` cue with `label: null` when combat motion is on, and none when it is off;
  - `beatPauseMs` is 400 or 560 × speed;
  - every existing prime, rewind, cross-frame and finish-the-cues case still holds.
- `combatStrikes.test.ts` (new, pure):
  - `strikesFor(cue, log)` puts movers and targets in the right sides, by the step entry's seat, with and without the `block` entries in the window;
  - several blockers give one mover and a centroid aim;
  - trample shakes the avatar at contact;
  - a first-strike blocker lunges, and a blocker facing a lunging attacker braces;
  - a planeswalker or battle target is a card;
  - deaths come from the beat's zone entries;
  - the 12-copy cap;
  - `lungeVector` clamps at ½ and 0.85 and gives contact plus 8 px in between;
  - `strikeTimeline(speed)`;
  - the hidden-page and late-start drops;
  - `combatMotion` against each of the three settings.
- `combatStrikes.render.test.ts` (new): a cue mounts a copy per mover and hides that live tile until the copy is gone. Reduced motion mounts nothing. A priming frame removes copies in flight and unhides tiles. A dead blocker's copy is drawn from the cache. A cached rect from another board size is skipped.
- `settings.test.ts`: `animations.combat` is classified as synced and filled by the merge.
- A Card render test: with no strike, the tile's computed transform is unchanged.

**Playwright (PR 1).** One new spec, `combat-strike-2614.spec.ts`, with two seats. The first attacks unblocked. During combat damage, a strike copy (`[data-strike-copy]`) is present, and within 2 s it is gone and the attacker's tile is visible again. With `animations.combat` off, no copy appears and the life total still changes. If the spec toggles the setting through the Settings row, that row's name goes in `labels.ts` (ADR 0125 §2). PR 1 touches `Card.svelte` and the attack and block display, so it runs the nightly E2E on its branch before merging.

**No server tests.** Nothing on the server changes.

---

## Delivery

Each PR goes into `develop`, Sprint S69, Issue #2614.

| PR | What | Size | Needs |
|---|---|---|---|
| — | **This ADR.** Docs only. | — | — |
| 1 | **The small version.** <br>• `combatBeats.ts`: an unlabelled beat for untagged combat damage, and `beatPauseMs`. <br>• `combatCues.svelte.ts`: the director, shared. <br>• `combatStrikes.ts` (new, pure): `strikesFor`, `lungeVector`, `strikeTimeline`, `combatMotion`. <br>• `animations.ts`: `STRIKE_*`, `lunge()`, `impactShake()`, `combat` in `AnimationConfig`. <br>• `settings.ts` and `Settings.svelte`: `animations.combat`. <br>• `CombatStrikes.svelte` (new), mounted in `Board.svelte`. <br>• `CombatArrows.svelte`: reads the shared director. <br>• `Card.svelte`: `--impact-x`, `--impact-glow`, `striking`. <br>• `PlayerIdentity.svelte`: the disc shake. <br>• The vitest suites and the Playwright spec of §8. <br>Covers the lunge, every impact, several blockers, trample's simultaneous shake, first and double strike, deaths, the setting and reduced motion. | 1–2 days | the owner's answers |
| 2 | **Polish.** <br>• A distinct lethal hit: a bigger shake and a red flash on a creature that dies, and on a player whose life reaches 0. <br>• The trample streak. <br>• The replay one-step rule (`Game.svelte`'s `beatsPrimeKey`). <br>• Tuning the copy cap and the timings from a cmd-dev playtest. <br>• Tests in the same files. | 1–2 days | 1 |
| 3 | **Sound.** `combat_resolve` moves to the first contact of a beat when combat motion is on (`Game.svelte`, `CombatStrikes.svelte`), with a vitest on when it fires. | ½ day | 1 |

After PR 1: a cmd-dev game with four seats, one combat with first strike and one with a double block, recorded in the PR. After PR 3: close #2614 with evidence.

## Consequences

- Everyone at the table sees who hit what, at the moment it happened. The log becomes the record, not the only way to follow combat.
- The rules' shape shows in the motion: one moment per damage step, simultaneous hits, no false order among blockers, and deaths after the damage.
- Nothing waits for the animation. A fast table stays fast, and a viewer's slow animation setting slows only their own screen.
- A creature that dealt no damage does not move, so a prevented hit looks like no hit. The log line says why.
- The board shows the end state before the strike plays: life totals, badges and empty slots change first, and the strike follows within about 200 ms at speed 1. Holding the board back would mean freezing frames, which ADR 0053 rejected.
- Combat with no first strike now has a beat in the tracker. Today it has none, so any later consumer of beats has to allow for unlabelled beats.

## Out of scope

- **Animating attacks and blocks as they are declared.** The arrows already draw on, and the `attack` and `block` sounds already play.
- **Non-combat damage** (Lightning Bolt, Sulfuric Vortex). It has no attacker to lunge, and the stack's target arrows and rings (ADR 0119 §4) already point at what it hits.
- **A phone-width layout** (owner constraint).
- **Freezing the board until the strike finishes.**
- **Bot pacing changes.**

## Calls made here

These are made by this ADR and can be overturned in review.

1. **The copy is art-only**, not a full `Card.svelte` with badges. The real tile's badges come back when it reappears 500 ms later.
2. **Z 37** for the layer (§2).
3. **Timing:** 180 / 60 / 260 ms, `power2.in` out, `power3.out` back, no wind-up (§2).
4. **Shake:** ±4 px, 160 ms, with a brightness flash (§2).
5. **The life popup stays at frame arrival.** It is not delayed to the impact. The life number under it has already changed, so delaying only the popup would make the two disagree. PR 2 may revisit this after a playtest.
6. **At most 12 copies in a beat** (§3).
7. **Strikes are dropped** if the page is hidden when the frame arrives, or if they would start more than 1 s late (§4).
8. **One director shared** between `CombatArrows` and `CombatStrikes` (§1).
9. **The attacking side is read from the step entry's seat** (CR 506.2), with no new wire field (Context).
10. **Combat motion off leaves ADR 0053's arrow pulses under `damagePopups`** (§5).

---

## Questions for the owner (answered)

Each question lists the recommended option first. On 2026-10-07 the owner chose (a) on all eleven (see [Owner decisions](#owner-decisions-2026-10-07)). The questions are kept here with the options that were not chosen.

1. **What starts the animation?**
   - (a) **The combat damage entries in a newly arrived frame, through ADR 0053's beat tracker** (§1). It is exact about source and target, and priming, undo and reconnects are already solved and tested.
   - (b) A second tracker over the same log entries, separate from the beats. It is just as exact, but duplicates the priming and rewind rules, and its clock can drift from the beat labels.
   - (c) Diffing the view. It cannot say who hit whom, loses creatures that died, and merges double strike.

2. **What moves?**
   - (a) **An art-only copy in a board-level layer, with the live tile hidden during the flight** (§2). It crosses panels, works for creatures that died, and follows an expanded board. Badges disappear for half a second.
   - (b) The live tile, through a new `Card.svelte` CSS variable. It is cheap, but is clipped at its own panel's edge, paints under the other seats, and cannot show a creature that died.
   - (c) Nothing moves: a streak or projectile runs along the combat arrow. It is the lightest option, but it reads as a spell, not a creature attacking.

3. **How far does the attacker lunge?**
   - (a) **Until it touches its target, clamped to between ½ and 0.85 of the distance** (§2). The contact sells the hit at any distance.
   - (b) A fixed ¾ of the centre-to-centre distance. It is the simplest, but near targets overlap heavily and far ones never touch.
   - (c) A fixed ½. It is calmer, but no hit ever visibly connects.

4. **Several blockers on one attacker.**
   - (a) **One lunge at the blockers' centroid, all of them shaking together** (§3). It matches CR 510.1c (no order) and CR 510.2 (simultaneous).
   - (b) One lunge per blocker, in log order. It is easier to follow, but shows an order the rules no longer have, and takes 500 ms per blocker.
   - (c) One lunge at the blocker that took the most damage, while the others shake. It is clear, but picks a favourite the rules don't.

5. **Trample damage to the player.**
   - (a) **The player's avatar shakes at the same contact, and PR 2 adds a streak from the blockers to it** (§3). Simultaneous, as CR 510.2 says.
   - (b) The lunge carries on through the blockers to the avatar, in two stages. It is vivid, but makes the player's damage look later than the blockers'.
   - (c) No extra motion: only the popup and the log. The least work, but trample is invisible on the board.

6. **Do blockers lunge?**
   - (a) **Only in a beat where their attacker dealt them no damage** (a first-strike blocker, or a 0-power attacker), and otherwise they brace (§3). The creature that struck first is the one that moves.
   - (b) Never. Their damage shows as a flash on the attacker. It is simpler, but a first-strike blocker killing the attacker looks passive.
   - (c) Always, meeting the attacker in the middle. It is symmetric, but busy, and every exchange looks the same.

7. **Creatures that die.**
   - (a) **The death plays after the hit** (CR 704.3): a dead attacker fades as it snaps back, and a dead blocker shakes, then fades in place (§3).
   - (b) Dead creatures don't move or shake. Only survivors animate, and the dead are gone at frame arrival as today. It is simpler, but the fight that killed them is never shown.
   - (c) As (a), plus a distinct death effect (a crack or crumble) in PR 1. It is richer, but doubles PR 1. Recommended for PR 2 as the "distinct lethal hit".

8. **Pacing.**
   - (a) **Display only: nothing waits for it** (§4). ADR 0053, 0119 §3 and 0121 §7 all chose this, and a slow-animation viewer slows only themselves.
   - (b) Bots wait: a `CombatHold` in each pace preset (fast 0, normal 600 ms, slow 1 s) after a combat damage frame. A fast bot table is easier to watch, but the normal pace already covers the strike.
   - (c) Every client's auto-pass also waits for its own strike to finish. The board never moves under a strike, but each viewer's speed setting changes when the table moves on.

9. **The setting.**
   - (a) **A new `animations.combat` toggle, default on, under the master switch and reduced motion, scaled by speed** (§5). It matches `dice` (ADR 0121).
   - (b) Reuse `damagePopups`, as ADR 0053's beats do. One fewer setting, but turning off popups would also turn off lunges.
   - (c) No toggle: only the master switch and reduced motion. The simplest, but someone who likes every other motion cannot turn combat off.

10. **Replays.**
    - (a) **Strikes play only on a one-frame step forward, and any other jump re-primes** (§6). Scrubbing stays readable, and the same fix covers the beats, the dice and the linger.
    - (b) Never in replay: the toggle primes and nothing plays. Simplest, but you cannot watch a combat back.
    - (c) As today: any forward jump plays everything it skipped. No work, but a jump over three combats plays them all at once.

11. **Sound.**
    - (a) **Later (PR 3): move `combat_resolve` to the first contact of each beat when combat motion is on, with no new files** (§7). The sound lands with the hit.
    - (b) A new hit sample for each impact. Louder feedback, but a new asset, and many hits at once stack (the 40 ms rate limit only dedupes the same name).
    - (c) Leave sound as it is. No work, but the sound comes about 200 ms before the hit it describes.

---

## Amendments

Dated notes. The decisions above stand as written. This section records where the code departed from them, and why.

### 2026-10-08 · What PR 1 and PR 2 built (#2624, PR 2)

**PR 1 (#2624) departed from the text in four small ways.**

1. **A late start is measured from the strike's scheduled start**, not from the frame's arrival (§4). The scheduled start is the frame's arrival plus the cue's own offset. At speed 2, the same-frame pause before beat 2 is 1120 ms (560 × 2). Measuring from the frame would drop every second beat at that speed, although nothing about it is late. `strikeLate(frameAt, atMs, now)` drops a strike only when it starts more than 1 s after `frameAt + atMs`.
2. **The impact flash is a white wash, not a brightness filter** (§2). The flash is a pseudo-element over the tile (`Card.svelte`) and over the avatar disc (`PlayerIdentity.svelte`), at `--impact-glow` × 0.5 opacity. A `filter` on the tile would fight the hover and phased-out filters. The strike copy's face has no such filters, so it keeps the brightness filter.
3. **The avatar disc shakes through the `translate` property**, not a transform. With `--impact-x` unset, which is every moment but a shake, `translate` is `none`. The disc is then not a stacking context, and the damage popup layers as it did before.
4. **Tiles are measured on every frame during combat**, as the clock folds the frame in, as well as on animation frames. Board's frame effect runs after the DOM has the frame. So a creature that dies in a later frame has already been measured in this one, even when no animation frame ran in between.

**PR 2 built the polish in Delivery, with these calls.**

5. **The lethal hit.** A creature that dies in the beat, and a player the beat eliminates, take a bigger shake: ±8 px over 240 ms instead of ±4 px over 160 ms. The flash is red (`--impact-wash`) instead of white. A mover that a hit back killed gets the red flash on its copy at contact. **A player is lethal when the beat holds their `eliminated` entry** (any cause but `concede`), not when their life reaches 0, as Delivery puts it. The state-based actions that follow the damage (CR 704.3) are in the beat already. So this covers 21 commander damage and ten poison counters, and it does not show a killing blow to a player who can't lose at 0 life.
6. **The crumble replaces the fade** (question 7's option (c), which the ADR deferred to PR 2). A dead target takes the lethal shake at contact, then breaks into 3 × 4 shards that fall and fade over 420 ms. Before, it faded over 240 ms (`STRIKE_DEATH_FADE_MS` is gone). A dead attacker no longer fades on its way back. It flies home, then crumbles there. At speed 1 a dead target's copy lasts 840 ms after its cue and a dead attacker's 920 ms. Both still play after the hit (CR 704.3). The shards are planned in `combatStrikes.ts` (`crumbleShards`, seeded by the card, so a card always breaks the same way) and made inside the copy, so they go when it does.
7. **Trample's streak** runs from the edge of the blockers' aim box (their centroid) to the edge of the player, planeswalker or battle that took the excess, one per such target. It draws during the 60 ms contact hold and fades over 220 ms. It is not drawn for a gap under 16 px. It is also not drawn when the block entries have left the 200-entry window: with no block entries, the excess cannot be told apart from a blocker.
8. **The replay one-step rule** is a jump counter in the prime key: `beatsPrimeKey(connected, replaying, epoch)` in `combatBeats.ts`. `replayJumpEpoch` keeps the epoch only for a one-frame step forward. A jump ahead, a step back, the same frame again, and entering or leaving the replay all bump it, so the next frame primes. The beats, the strikes, the dice and the linger share the key, so all of them follow the rule.
9. **Tuning is not done.** Delivery asks for the copy cap and the timings to be tuned from a cmd-dev playtest, and no playtest was possible for this PR. The cap stays at 12 and the timings stay as in §2. They are named constants (`STRIKE_*`, `IMPACT_*`, `STREAK_*` in `animations.ts`), so a playtest can change them in one place.
10. **The after-PR-1 four-seat check** was not run on cmd-dev. In its place, `combat-strike-2614.spec.ts` covers one combat on a two-seat table: a first striker lunges before the regular beat, a trampler double-blocked by two 2/2s lunges once, the excess draws a streak, and both blockers crumble. Its screenshots are in the branch run's Playwright report. Still not seen in a browser: four seats, an expanded board (ADR 0120), a planeswalker or battle target, and a lethal hit on a player.
