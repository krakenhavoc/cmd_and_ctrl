# ADR 0133 — Opening-hand actions: beginning the game with a card on the battlefield

**Status:** Accepted · 2026-10-07 · S58 — Deck requests, October batch
**Issues:** [#2190](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2190) (the Sami Whammy deck request: Gemstone Caverns). The same gap is the caveat on Leyline of Anticipation, of Lifeforce, of Mutation, of Punishment, of Sanctity and of the Void, and on Leyline Axe.
**Numbering:** the AGENTS.md §4 sweep on 2026-10-07 found 0132 (`0132-day-and-night.md`, branch `feat/2561-day-night`) as the highest number on any remote head, and no open claim on 0133. This ADR takes **0133**.
**Builds on:** [ADR 0118](0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md) (its out-of-scope list names "the opening-hand action (CR 103.6a) for any Leyline"), [ADR 0121](0121-animated-dice.md) (the start of the game: the roll, then the deal), the mulligan window of `game/mulligan_order.go` (#2237), and the chained-choice prompts of `game/chained_choice.go`.

---

## Context

Eight catalog cards print a sentence about the **opening hand** that is not a spell or an ability of a permanent:

> If this card is in your opening hand, you may begin the game with it on the battlefield.

Seven of them are exactly that (the six Leylines and Leyline Axe). Gemstone Caverns adds three riders: "and you're not the starting player", "with a luck counter on it", and "If you do, exile a card from your hand."

The engine had no place to put the sentence. Deck setup puts every card in the library, the mulligan window (`MulligansOpen`) only asks keep or mulligan, and nothing between "every seat has kept" and "the first step begins" was a moment a card could speak at. So each of the eight shipped weaker than printed with a caveat, which is the right direction for a simplification (#259), and which ADR 0118 left for later.

### The rules

- **CR 103.6a** (quoted in ADR 0118): "If a card allows a player to begin the game with that card on the battlefield, the player taking this action puts that card onto the battlefield."
- **CR 103.6** puts these actions after the starting hands are final and takes them in turn order beginning with the starting player. The cards' own comments cite it as "CR 103.6" and "CR 103.6a". The pinned Comprehensive Rules text was not reachable from the session that wrote this ADR, so those two citations rest on ADR 0118's quotation and on the existing card comments; check the turn-order wording of 103.6 against the TXT before relying on it.
- **CR 110.4 / 614:** the card arrives as an ordinary entry. Nothing here invents a second way onto the battlefield.

### What exists

- **The window closes in one place.** `settleMulliganLocked` sets `MulligansOpen` false when every live seat has kept, and then runs the first step's entry hooks. Departures and keeps both go through it.
- **The door onto the battlefield from a hand.** `PutFromHandOntoBattlefieldThenForEffect` runs the CR 614 window, the layer timestamp and the card's enters hooks, and tells its continuation which permanent arrived.
- **Prompts that chain.** `PendingChoiceConfirm` ("do A, or do B") and `PendingChoiceChooseCards` ("choose N of these") exist for exactly this shape, are enumerated for bots (`legal/choices.go`) and render in the client.

---

## Decisions

### 1. The card declares the action; the engine asks when the window closes

`Spec.OpeningHand *game.OpeningHandAction`, read through `CatalogOpeningHand` (a `CardDef` slot, #622). The action is **plain data**:

```go
type OpeningHandAction struct {
    NotStartingPlayer  bool           // Gemstone Caverns: not the seat that takes the first turn
    EntersWithCounters map[string]int // "with a luck counter on it"
    ExileFromHand      int            // "if you do, exile a card from your hand"
}
```

Cards build it with `effects.BeginTheGameOnTheBattlefield(opts...)` and the options `NotTheStartingPlayer()`, `WithCounter(kind, n)` and `ThenExileFromHand(n)`. Data, not closures, so the ADR 0041 closure ratchet gains no route and a catalog entry stays restorable. A card that prints a fourth rider adds a fourth field.

`settleMulliganLocked` calls `offerOpeningHandActionsLocked` once, as the window closes and before the first step's entry hooks. For each seat in turn order from the starting seat, and each card in that hand in hand order, whose catalog entry carries an action (and which the `NotStartingPlayer` rider does not exclude), it queues a `PendingChoiceConfirm` to the owner: "Begin the game with *X* on the battlefield?" with the answers "Begin the game with it" and "Keep it in my hand".

### 2. Yes puts the card in through the ordinary hand door, then does the riders

The accept branch checks the card is still in that seat's hand, then calls `PutFromHandOntoBattlefieldThenForEffect`. In the continuation: the counters are placed, and the exile pick is queued (a `PendingChoiceChooseCards`, exactly one card, from what is left in the hand, nothing asked of an empty hand). The pick is a real pick: it cannot be skipped and cannot name the land that just left. The card is exiled through `ExileCardForEffect`, so it sees the CR 614 window like every other exile.

### 3. The table waits on the prompts; the first step is not held back

The prompts are queued together and the first step's entry hooks run as they did before. The table is parked on them the same way as on any open choice (`ChoiceBlocksTable` is true for a confirm): `pass_priority` and the gated verbs refuse, and `legal.EnumerateFor` offers the seat that owes one its answers and every other seat nothing.

This is a **narrowing of "before the game begins"**, chosen on purpose. Holding the entry hooks until the last answer would need new game state (an "offers outstanding" flag that restore, undo and a departing seat each have to keep honest), and a prompt dropped on a seat's departure would then have to re-run the hooks or the table would wedge. As built, the continuation is the prompt's own branch, a dropped prompt simply leaves its card in hand, and there is no flag. What the narrowing costs is nothing a card can see: turn one's untap and upkeep entry have no permanent to act on, and a permanent that arrives afterwards enters untapped like any other.

### 4. Counters are placed after the card has entered

`ZoneEntryOptions` has no "enters with counters" field, and the entry replacement path is for effects that watch other cards' events. Nothing can respond to an opening-hand action and no replacement effect is on a battlefield that is still empty, so placing the counter immediately after entry is indistinguishable from entering with it. Gemstone Caverns' mana ability reads the counter at activation, so it is right from the first tap.

### 5. What the table can see

The question names a card in a hidden zone, and a player who declines has revealed nothing in paper. The prompt is marked `PendingChoice.private` and the view carries `private_text`. The chooser gets the source card, the question and the branch labels. Every other viewer, spectators and admins included, gets the prompt, its kind and its chooser, with a neutral reason ("Deciding on an opening-hand action") and no source. The table still sees that a seat owes an answer, which it has to, because the game waits on it. It cannot see which card, or whether the card is a Leyline or a Gemstone Caverns.

### 6. The bot

No new prompt kind. A confirm and a choose-cards are already enumerated, and the decline is the confirm's always-legal answer (#544). The heuristic takes the offer by default, and picks the card to exile by its ordinary choose-cards scoring. `aiseat/opening_hand_test.go` plays the opening to the end with the heuristic in both seats.

---

## Calls made here (for owner review)

These are choices the ADR makes without a recorded owner answer. Each is reversible without a wire break.

1. **Prompts are asked together at window close, not one after another.** CR 103.6 orders them by seat. The prompts are *queued* in that order but may be answered in any order, and nothing observable depends on it today. If two opening-hand actions ever interact (a card that cares what another seat began with), the offers would need to chain.
2. **The first step is not held while prompts are open** (decision 3).
3. **The existence of an opening-hand prompt is visible to the table**, though not its card (decision 5). The alternative, asking every seat a blind question, would also hide that a seat holds such a card at all, at the cost of a prompt per seat per game for a mechanic eight cards use.

---

## Out of scope

- **"You may reveal this card from your opening hand. If you do, …"** (the Chancellors, Providence, Sphinx of Foresight, Devourer of Destiny) is a different shape: a reveal that arms a delayed trigger at the first upkeep or the first main phase. It needs the same moment and a different slot. None of them is catalogued, and none is on a deck request that names it.
- **Quicksilver, Brash Blur** ("begin the game with him on the battlefield") is the same shape as a Leyline and would take `BeginTheGameOnTheBattlefield()` as is; it has other unbuilt text and is not catalogued.
- **The uncatalogued Leylines** (Abundance, Combustion, Hope, Lightning, Resonance, Singularity, Transformation, Vitality, the Guildpact, the Meek) wait on their own static abilities, not on this seam.

## Delivery

One PR: the engine (`game/opening_hand.go`, the `CardDef` and `Spec` slots, the hook in `settleMulliganLocked`, `PendingChoice.private` and the view's `private_text`), the eight cards as proof, the tests, and this ADR. The client needs no change beyond the `private_text` field on the type: the confirm and choose-cards modals already render.

## Snapshot impact

None. `OpeningHandAction` is catalog data. The prompts carry continuation frames, which the restore-point census already counts, so a table is not restorable only while an opening-hand answer is outstanding; `PendingChoice.private` is unexported and never written.
