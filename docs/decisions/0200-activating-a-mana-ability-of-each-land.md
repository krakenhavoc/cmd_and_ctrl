# ADR 0200 — Activating a mana ability of each land

**Status:** Accepted · 2026-10-16 · S76
**Issue:** [#3952](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3952) (split from [#3531](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3531))
**Numbering:** every remote head was listed (`develop`, `main`, `cost-ledger`, `docs/issue-audit`, `feat/3874-restricted-autopay`, `feat/3955-cast-from-outside`, `feat/3953-hand-revealed`, `feat/750-conditional-block-restrictions`, `feat/playmats`, `fix/2947-resolution-copy-casts`, `fix/2950-darkness-replacement-continuation`, `fix/3940-different-names-sacrifice`, `fix/caddy-reload-admin-off`, `screenshots/3934-playmat`, `wip/836-one-click-default`, `pr/2326`). `feat/3953-hand-revealed` (#3975) holds **0199**, so this is **0200**.
**Builds on:** the mana ability activation path (CR 605), [#1443](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1443)'s colours named up front, the `option_pick` prompt (#568), [ADR 0187](0187-any-player-pays-per-object-graveyard-and-either-payments.md) §4 (the CR 510.1a gate) and the delayed triggers of [ADR 0188](0188-skipped-turns-and-turn-bound-effects.md) Decision 5 (bound to one phase).

---

## Context

Two cards make a player activate a mana ability of each land they control and then lose all unspent mana:

- Drain Power: "Target player activates a mana ability of each land they control. Then that player loses all unspent mana and you add the mana lost this way."
- Pygmy Hippo: "Whenever this creature attacks and isn't blocked, you may have defending player activate a mana ability of each land they control and lose all unspent mana. If you do, this creature assigns no combat damage this turn and at the beginning of your next main phase this turn, you add an amount of {C} equal to the amount of mana that player lost this way."

The engine activated mana abilities when a player chose to (a click, the auto-tapper paying a cost), never because an effect told a player to. Nothing could count the mana a player loses.

CR 605.3a lets a player activate a mana ability "whenever a rule or effect asks for a mana payment, even if it's in the middle of casting or resolving a spell or activating or resolving an ability". Drain Power's rulings (2004-10-04) say the player is forced to draw mana from their untapped lands, but chooses how when a land has several mana abilities or an ability with choices, and chooses how much when the amount varies. CR 106.4 is where unspent mana is lost.

## Decision

### 1. The player activates one mana ability of each land, through the ordinary path

`effects.ActivateAManaAbilityOfEachLand` takes the lands the player controls as the instruction begins, in battlefield order. For each land, `game.LandManaOptionsForEffect` lists the ways the player can activate one of its mana abilities now. It uses the gates the activation itself checks first: the {T} can be paid, the ability isn't exhausted, its condition and the activation gate pass, and its life, energy, mill and mana components can be paid from the pool. `game.ActivateLandManaForEffect` then activates the chosen way through `activateManaAbilityLocked`, so costs, riders (City of Brass's damage), triggered mana abilities and the activation records are the same as for a click. It is not a priority action (CR 117.3c).

A tapped land, or one with no mana ability the player can activate, adds nothing (CR 101.3). An ability whose cost needs another choice (sacrifice, tap, discard or exile another object) is not offered: no land the two cards care about prints one, and asking it would need a second prompt for each land.

### 2. The player chooses, colours included, in one question per land

A picking slot's colours are named up front (`ManaAbilityParams.Colors`, #1443), so each (ability, colours) pair is one option: City of Brass offers five, and Graven Cairns offers {C}, {B}{B}, {B}{R} and {R}{R}. A slot that must add different colours never offers a repeat. A land with one option activates it without asking, and a land with several asks the player one `option_pick`. No `mana_pick` follows the activation. If an option is refused at activation because the board moved, the question is asked again without it.

A land with an ability whose mana cost the pool can't pay yet (a filter land) is asked after the other lands, so the mana they add can pay for it if the player chooses that ability. That is the one ordering the cards' text leaves to the player that matters.

### 3. "Loses all unspent mana" empties the whole pool

`game.LoseAllUnspentManaForEffect` empties the player's pool, including mana that was in it before the effect began, and returns the mana lost. Effects that keep mana as steps end (`mana_keep.go`) don't apply: this is an instruction to lose the mana, not CR 106.4's emptying between steps.

- Drain Power adds mana of the same types and amounts to its caster. That mana is new mana Drain Power makes, so restrictions and spend riders that came from the ability that first made the lost mana don't carry over.
- Pygmy Hippo schedules a delayed trigger (CR 603.7) for the next main phase still to begin this turn (`game.NextMainPhaseThisTurnForEffect`), bound to that phase's id, so a turn that ends early never carries it into a later turn. The body is `pygmy-hippo/add-colorless`, which adds {C} for the amount. A turn with no main phase left adds nothing. Saying yes marks the Hippo (CR 510.1a) before the lands are asked.

### 4. Snapshot

Nothing new is stored. While the questions are open, the continuation is a closure on an `option_pick`, like the other cards in #3531, so that table is not a restore point until the walk ends. The walk is a value that each continuation copies, so an undo that brings a question back also brings back the walk as it stood. The delayed trigger is data (a ledger body and an amount).

## Consequences

- Drain Power and Pygmy Hippo ship with no simplifications.
- A later card that makes a player tap their lands for mana reuses the primitive.

## Out of scope

- A land whose mana ability needs a cost choice beyond the ones above, should a card ever make that matter.
- A prompt queued by a triggered mana ability during the walk (none of the catalogued ones asks) is answered after the pool is lost.
