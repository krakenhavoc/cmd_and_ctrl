# ADR 0169 — The life that damage costs runs the life-loss window

**Status:** Accepted · 2026-10-13 · S58 — Deck requests (October batch)
**Issue:** [#3388](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3388) (Bloodletter of Aclazotz, deck requests #3248 and #3173).
**Numbering:** the AGENTS.md §4 sweep on 2026-10-13 (`origin/develop`, `origin/main` and the heads of every open PR) found **0168** the highest, so this is **0169**.
**Builds on:** #482 (every non-damage life change runs one CR 614 window, `game/life_tail.go`), #2105 (`Event.DamageNotLifeLoss`, the life a damage event cost), [ADR 0085](0085-life-total-cant-change.md) (the life-total lock) and the Angel's Grace floor (#3172, `game/life_floor.go`).

---

## Context

Bloodletter of Aclazotz: "If an opponent would lose life during your turn, they lose twice that much life instead. (Damage causes loss of life.)"

- **CR 120.3a.** Damage dealt to a player by a source without infect causes that player to lose that much life.
- **CR 119.4.** Paying life is losing life.
- **CR 614.1 / 616.1.** A replacement effect that watches "would lose life" applies to every way a player loses life, so it applies to the loss damage causes. When several apply, the affected player chooses the order.

Since #482 every non-damage loss and every life payment runs the CR 614 window as `RepEventLife`. Damage does not: `applyResolvedDamageToPlayerLocked` (`game/damage_tail.go`) lowered the total directly once the DAMAGE replacements had settled, and only fixed rules read the life a damage event costs (infect, the lock, the Angel's Grace floor). A doubler on `RepEventLife` therefore doubled a drain and a paid Thoughtseize but not a Lightning Bolt or combat damage, which is the case the card is played for.

The files said on purpose that a life replacement must not see damage. That was about a second bite at the DAMAGE (a prevention shield halving a Bolt twice) and about life GAIN replacements, which never match a loss. It does not apply to the life the damage costs, which is a separate result.

## Decisions

### 1. The damage result opens the `RepEventLife` window

Once the damage replacements have settled and the damage is split into its results (infect, toxic and the as-though statics, unchanged), the life it would cost runs the CR 614 life window as a new `RepEventLife` event: `LifePlayer` the damaged player, `LifeDelta` minus the loss, `Source` the damage source. It is a new event, not a nested one, as ADR 0056 Decision 3 already made the damage result's counters. Nothing about the damage rides it: the damage amount is settled.

No new event kind. A "would lose life" replacement is worded about life, and every one should see damage, drains and payments alike. Gain replacements (`LifeDelta > 0`) and Strong, the Brutish Thespian's radiation clause (`LifeFromRadiation`) cannot match it. A loss of zero (infect, fully prevented damage) opens no window.

### 2. It settles inline (`mustSettleNow`)

The damage tail cannot pause between the damage and its result: combat damage is dealt simultaneously (CR 510.2), and the event, the commander tally and lifelink all follow from it. So the event is `mustSettleNow`, like a life payment: a CR 616 ordering is applied in the gathered order and an optional replacement is skipped.

No card in the catalog makes that observable. The only life-loss replacements are Bloodletter, whose copies are the same modification (two copies quadruple in either order, #792), and the preemptive lock. A future card with a different loss replacement that asks a question would be the case to revisit.

### 3. Order: the window, then the floor, then the lock

- **The lock** ("your life total can't change", ADR 0085) is the window's preemptive built-in, so it cancels the event inside the window. The existing check in the damage tail stays and gives the same answer.
- **The Angel's Grace floor** caps the result AFTER the window. The floor is a replacement on the life the damage costs; under CR 616.1 the affected player picks the order, and capping last is never worse for them. So 5 damage at 3 life under a Bloodletter and a floor leaves 1.

### 4. A loss larger than the damage is a negative `DamageNotLifeLoss`

`Event.DamageNotLifeLoss` is the part of the damage that cost no life, and `DamageLifeLoss()` is `Amount − DamageNotLifeLoss`. A doubled loss is recorded as a negative value: 3 damage that cost 6 life carries `DamageNotLifeLoss = −3`.

- No snapshot change: the field is already an `int` in the v8 shape, and its zero value stays the ordinary case.
- `DamageLifeLoss()` needs no change, and a binary from before this ADR reads the doubled loss correctly too.
- Every life-loss reader sees the doubled amount: "whenever a player loses life" triggers (`s22PlayerLostLife`, `b04OpponentLostLife`, Exquisite Blood), the turn tally's `LifeLost`, Vilis and The Master of Lake-Town.

### 5. The damage is unchanged

`Amount` is still the damage dealt. Lifelink gains the damage amount (CR 702.15b), the CR 903.10a commander tally counts the damage, and "is dealt damage" triggers see the damage. Only the life total, and the life-loss readers above, see the doubled loss.

## Consequences

- Bloodletter of Aclazotz ships Full as one `ReplacementEffect` (`effects.OpponentsLoseTwiceThatMuchLifeDuringYourTurn`) with no damage-specific code.
- Any later "would lose life" replacement automatically sees damage. A replacement that must not see it has no flag to read yet. None is printed in the catalog.
- If a replacement in the window turned the loss into a gain, the damage would cost no life and gain nothing. No printed card does that to damage.
