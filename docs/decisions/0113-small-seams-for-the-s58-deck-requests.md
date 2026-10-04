# ADR 0113 — Small seams for the S58 deck requests

**Status:** Proposed · 2026-10-03 · S58 — Deck requests, October batch (tracker [#2077](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2077))
**Owner decisions:** pending. Two questions are open; see [Open questions for the owner](#open-questions-for-the-owner). Everything else below is settled by the rules or by an earlier owner decision, and is written as decided.
**Issues:** [#2072](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2072) (a spell that reads what its sacrifice cost took), [#2073](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2073) (annihilator), [#2074](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2074) (statics that set or change a maximum hand size), [#2075](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2075) (undying and persist).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-03. I ran `git fetch --all --prune` and read the `docs/decisions/` file names on all 36 remote heads (`origin/develop`, `origin/main` and 34 chore, docs, feat, fix, repro and wip branches). The highest number anywhere is 0112. This one takes **0113**.
**Amends:** [ADR 0100](0100-delve-either-or-and-variable-sacrifice-costs.md) (owner decision 5: Corpse Cobble's list of sacrificed permanents, §1 here). A pointer line goes into it with the first implementation PR.
**Builds on:** [ADR 0014](0014-combat-keywords.md)'s 2026-09-24 amendment (prowess, the first keyword derived from the ability list), [ADR 0106](0106-five-small-seams-from-the-s50-rechecks.md) §3 (evolve, the second) and owner decision 6 (a PR lands every card its seam unblocks), [ADR 0056](0056-infect-wither-toxic.md) (toxic, the numbered keyword token), [ADR 0109](0109-rule-gates-land-types-mana-and-cost-components.md) §8 (a payment fact reaches the effect), [ADR 0041](0041-game-persistence.md) phase 3 (keyed bodies and restore points), [ADR 0093](0093-abilities-granted-to-other-permanents.md) (abilities granted by a layer-6 static) and [ADR 0033](0033-ai-bot-seat.md) §1 (the legal-move enumerator).

This ADR was written plan-first. No engine code changed with it. The engine and card changes land in the PRs listed under [Delivery](#delivery).

---

## Context

The S58 triage of the seven October deck requests found 48 class B cards: each needs one small or medium engine change. Four of those changes unlock 12 requested cards between them, and each is one rule the engine does not express yet. They are grouped here because they are small and independent.

The seams doc runs stale, so every claim below was checked in the code on `origin/develop` at `f6444416`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`, effective September 25, 2026). Rulings are Scryfall's, read on 2026-10-03. Card counts come from the local Scryfall dump, counted by oracle ID, Commander-legal only.

The triage report was right about the gaps and wrong about four details:

- **#2072 is not about variable counts.** Fling, Kazuul's Fury, Momentous Fall and Tend the Pests each sacrifice exactly one creature. Their gap is that a spell cannot read the creature its cost took. Only Corpse Cobble's count is variable. The registry row's slug, `variable-sacrifice-cast-cost`, stays, but its summary changes (§1).
- **Extus, Oriq Overlord is not unlocked by #2072.** ADR 0100 left it waiting on its magecraft ("cast or copy") and its Avatar token's attack trigger, not on this record. §1 lands it only if those turn out to exist; otherwise it moves to the row for its real blocker.
- **Unstoppable Slasher is not in the catalog.** The triage cited it as the existing user of `ReturnFromGraveyardWithCountersForEffect` and `LastKnownCountersForEffect`. The actual users are Perennation, Makeshift Mannequin, Resourceful Defense and the Ozolith-style "if it had counters" readers. The pair exists; the precedent doesn't.
- **The Eldrazi shuffle trigger does not exist yet.** Only its body does (`finaleShuffleGraveyardIntoLibrary`). No catalog card has "When this is put into a graveyard from anywhere". The closest helper is `WhenThisIsPutIntoYourGraveyardFromYourLibrary`. §2's PR adds the from-anywhere form.

### Sizing

| § | Seam | Engine change | Requested cards | Seam pool (Commander-legal) | PR size |
|---|---|---|---|---|---|
| 1 | #2072 a spell reads what its sacrifice took | one `PaidCost` field, filled at three payment sites, carried by two copy paths | Fling, Kazuul's Fury, Momentous Fall, Tend the Pests (+ Corpse Cobble, waiting since ADR 0100) | 34 spells that read their sacrificed permanent | S engine, M with the pool |
| 2 | #2073 annihilator | a numbered keyword token and one engine trigger | Kozilek, Butcher of Truth; Ulamog, the Defiler; Ulamog, the Infinite Gyre | 13 print annihilator, 5 grant it or have it by text | M |
| 3 | #2074 maximum hand size | one `Spec` slot, a timestamp-ordered fold, one `Player` field | Jin-Gitaxias, Core Augur; Null Profusion; Price of Knowledge | 19 set or change it, 4 give other players no maximum | S engine, M with the pool |
| 4 | #2075 undying and persist | two keyword tokens, the leaves-the-battlefield harvest learns keyword triggers | Gleeful Arsonist, Persistent Constrictor (+ Murderous Redcap's caveat) | 23 undying, 25 persist, 14 grant one | M |

---

## 1. A spell that reads what its sacrifice cost took (#2072)

### What exists, what is missing

- **The count exists.** `PaidCost.Sacrificed` (`game/paid_cost.go`, #1213) is how many permanents a sacrifice component took. It is filled at three sites: a spell's additional cost (`paidWithSacrifices`, `game/mutations.go`), an activated ability (`game/activated.go`) and a mana ability (`game/mutations.go`'s mana path). Readers use `ctx.Sacrificed()`.
- **Which permanents is not recorded.** Nothing on the record names the sacrificed objects. Activated abilities work around it: `b17PermanentSacrificedToPay` (`cards/effects/batch17_helpers.go`) scans the event log backwards from the announcement for an `EventSacrifice`. It finds one permanent, and only for an activated ability. A spell has no equivalent.
- **Last-known information exists.** `Game.lastKnownPermanents` (`game/permanent_lki.go`, #1379) keeps one `PermanentInfo` per departed permanent object for the rest of the turn, keyed by instance ID and told apart by `ObjectEpoch`. `PermanentForEffect(ObjectRef)` answers live while the object is on the battlefield and from the record once it has left. Clone and the snapshot carry it.
- **The record already has the precedent.** `PaidCost.TappedOthers` (ADR 0071) and `PaidCost.Delved` (ADR 0100) record objects a payment touched as identity plus epoch. `Exiled` and `Discarded` record cards by instance ID.
- **A spell copy drops the count.** `CopySpellForEffect` (`game/spell_copy.go`) carries `OptionalCosts`, `GiftOpponent`, `CostBranch` and `Discarded`, but not `Sacrificed`. A copied Vicious Betrayal reads zero. An ability copy (`game/ability_copy.go`) does carry `Sacrificed`.

### The rules

- **CR 601.2b / 118.8:** an additional cost is announced while the spell is cast and paid with the rest of the total cost (CR 601.2h).
- **CR 400.7j:** "If the cost of a spell or ability causes an object to move to a public zone, that spell or ability's effects can find that object."
- **CR 608.2h:** an effect that needs information from a specific object that has left its zone "uses the object's last known information". The rulings say the same for each card: Fling uses "the sacrificed creature's power as it last existed on the battlefield" (2024-04-12), and Momentous Fall checks "the sacrificed creature's last known existence on the battlefield … to determine its power and its toughness" (2010-06-15).
- **CR 707.10:** "If an effect of the copy refers to objects used to pay its costs, it uses the objects used to pay the costs of the original spell or ability." The Tend the Pests ruling (2021-04-16): copies "use the power of the creature sacrificed for the original spell".
- **CR 107.1b:** a calculation may use a negative value, but one "that would determine the result of an effect" and yields a negative number uses zero. So Corpse Cobble sums the powers, negatives included, and floors the total at zero. Fling with a −1 power creature deals no damage.

### Decision

1. **The record.** `PaidCost.SacrificedObjects []ObjectRef` (`json:"sacrificedObjects,omitempty"`): every permanent the payment sacrificed, as the battlefield object it was (instance ID and `ObjectEpoch`), in the order named. It records exactly the set `Sacrificed` counts, so `len(SacrificedObjects) == Sacrificed` on every new record. An alternative cost's sacrifice (`AltCostIDs`, flashback "sacrifice three creatures") is not counted by `Sacrificed` today and is not recorded here either: no printed card reads it.
2. **Where it is filled.** At all three payment sites, so spells, activated abilities and mana abilities have one record. The refs are read before `payCostSacrificesLocked` moves the permanents, the way `PaidTap` is read before the tap. `IsZero` and `clonePaidCost` learn the field.
3. **Copies.** `CopySpellForEffect` carries `SacrificedObjects` and, fixing the gap above, `Sacrificed` (CR 707.10). `ability_copy.go` carries `SacrificedObjects` beside the `Sacrificed` it already carries.
4. **The reader.** `ctx.SacrificedPermanents() []game.PermanentInfo`: each ref through `PermanentForEffect`, so each answer is the object as it last existed on the battlefield (CR 608.2h). A ref with no record (the object left by a route that bypasses the battlefield exit, CR 800.4a) is skipped, which errs weaker. Two conveniences sit on top: `ctx.SacrificedPower()` (the first one's power, floored at zero) and `ctx.SacrificedTotalPower()` (the sum, floored at zero after summing, CR 107.1b). Momentous Fall reads `Toughness` off the same `PermanentInfo`.
5. **The log scan.** `b17PermanentSacrificedToPay` reads the record first. It keeps the log scan only as the fallback for a stack item with `Sacrificed > 0` and no refs, which only a restore point written before this field can produce. The fallback's comment says it can be deleted once no such restore point can exist.
6. **Lifetime.** The record outlives nothing it refers to. `lastKnownPermanents` lasts the turn, and a spell or ability on the stack cannot outlive the turn, because a step ends only with an empty stack. A token's record survives the token ceasing to exist (CR 704.5d): only leaving the game forgets a record.

### Snapshot impact

Additive within schema v7. `sacrificedObjects` appears under `stackMeta[].paid`, `pendingTriggers[].paid` and `lastKnownStack[].item.paid`, recorded with `-update-shape`. Since #1497 a binary refuses a stack-item field it does not know, so a rollback refuses a file that has one: the rollback case, needing no bump. A new corpus board, `sacrifice_cost_objects.json` (Fling on the stack after its cost was paid), is written with `-write-corpus`, which adds a board without touching the existing files.

### Wire, client and bot

No wire change: the payment record is not on the wire. The enumerator already offers every candidate as its own move for a one-permanent sacrifice (`sacrificePayments`, N = 1), so a bot casting Fling already chooses which creature to throw. A variable clause keeps ADR 0100's "zero plus up to three counts, cheapest fuel first".

### Cards

- **Fling, Kazuul's Fury // Kazuul's Cliffs, Momentous Fall, Tend the Pests: Full**, each verified against its oracle text and rulings. Kazuul's Fury is a front face: its land back is already in `mdfc_lands.go`, so the face gets its own file in the shape of `agadeems_awakening.go`. Tend the Pests creates `PestToken`, which exists.
- **Corpse Cobble: Full** (owner decision 5 of ADR 0100, closed by this record). Its flashback still pays the additional cost (2021-09-24 ruling), and that path already works.
- **Extus, Oriq Overlord:** checked in the PR. It lands only if its magecraft on a copy and its Avatar token's attack trigger are expressible; otherwise it moves to the row for that blocker, with the reason.
- The rest of the pool follows ADR 0106 owner decision 6, subject to [question 2](#open-questions-for-the-owner): Thud, Eldritch Evolution, Neoform, Carrion, Life's Legacy and the rest of the 34 spells that read their sacrificed permanent.

### Tests

- Fling: an anthem's bonus and +1/+1 counters on the sacrificed creature count, and a −1 power creature deals 0. Nobody can respond between the choice and the sacrifice (the 2019-10-04 ruling), so the power is the one it had as the cost was paid.
- Momentous Fall reads both power and toughness from the same record.
- Corpse Cobble with zero, one and three creatures; with a −2 power creature among them; and cast with flashback.
- Copies: a copied Tend the Pests makes the original's number of Pests, and a copied Vicious Betrayal now reads the original's count.
- A sacrificed token and a sacrificed commander are both read.
- The activated path: Greater Good and Jarad read the record, and a hand-built item with `Sacrificed: 1` and no refs still finds its permanent through the fallback.
- A snapshot round trip with Fling on the stack.

### Registry

`variable-sacrifice-cast-cost` goes to **Implemented**. Its summary gains "and a spell can read the permanents it sacrificed, like Fling's power". `Missing` is emptied, Corpse Cobble leaves `Waiting`, Extus moves as above, `Rules` gains 608.2h, 400.7j and 707.10, `Issue` becomes 2072, and `Examples` gains Fling and Corpse Cobble.

---

## 2. Annihilator (#2073)

### What exists, what is missing

- **Nothing for annihilator.** It is not in `canonicalKeywords` (`game/keywords.go`), so the deck importer drops Scryfall's "Annihilator", and no trigger exists.
- **The models exist.** Prowess and evolve (`game/prowess.go`, `game/evolve.go`) are canonical tokens. `keywordTriggersFor` hands `TriggersForCard` one engine `TriggeredAbility` per token, with a keyed body, so a printed, token-carried or layer-6 granted instance works with no catalog entry. Toxic (`game/infect_wither_toxic.go`) is the numbered precedent: the family key `toxic`, tokens `toxic N` minted by `CanonicalToxicToken`, and the importer's oracle-line scan supplies the number Scryfall's `keywords` array leaves out.
- **The sacrifice exists.** `PermanentsPickedThenForEffect` with min = max = N, then `SacrificeAllThenForEffect`, is Lotus Field's "sacrifice two lands": one choice, one simultaneous exit, with CR 903.9 commander answers handled on the way.
- **The defending player exists.** `DefendingPlayerForAttackerForEffect` reads CR 508.5 for an attacker, including one whose planeswalker or battle has left combat (#1364).

### The rules

- **CR 702.86a:** "Annihilator is a triggered ability. 'Annihilator N' means 'Whenever this creature attacks, defending player sacrifices N permanents.'"
- **CR 702.86b:** "If a creature has multiple instances of annihilator, each triggers separately."
- **CR 508.3a:** "Whenever [a creature] attacks" triggers when it is declared as an attacker, and not when it is put onto the battlefield attacking.
- **CR 508.5 / 508.5a:** the defending player is the player the creature is attacking, the controller of the planeswalker it is attacking, or the protector of the battle. If it is no longer attacking, it is the one it was attacking "before it was removed from combat". In multiplayer it is determined per creature.
- **CR 701.21a:** a player sacrifices only permanents they control. Told to sacrifice four with three, they sacrifice three.
- Rulings (2010-06-15, repeated on every Eldrazi): the trigger resolves in the declare attackers step and the permanents are sacrificed before blockers are declared. A planeswalker being attacked may be sacrificed, and the attacker keeps attacking and deals no combat damage if unblocked. Ulamog, the Defiler (2024-06-07): use the +1/+1 counters on it "at the time its annihilator ability resolves".

### Decision

1. **The token.** `KeywordAnnihilator = "annihilator"` joins `canonicalKeywords` as a family key, exactly like toxic: the tokens are `annihilator N`, minted by `CanonicalAnnihilatorToken`, and `CanonicalKeywords` refuses a bare `annihilator`. It is **cumulative** (`KeywordIsCumulative`, CR 702.86b), so `AppendKeywordAbility` keeps a granted `annihilator 2` beside a printed `annihilator 4`.
2. **The importer.** `deck.printedKeywordsForFace` treats "Annihilator" as it treats "Toxic": the array marks it wanted and the oracle-line scan supplies the number ("Annihilator 4"). A card whose lines name no number stamps nothing, which errs weaker. Ulamog, the Defiler's Scryfall array has no "Annihilator", so the importer never stamps it.
3. **The trigger.** `keywordTriggersFor` returns one annihilator trigger per token, after prowess and evolve. It watches `EventAttack` for its own source declared as an attacker (CR 508.3a; the same predicate `ThisAttacked` uses card-side, so a creature put onto the battlefield attacking does not trigger). `Build` makes a keyed item, body `annihilator/sacrifice`, with `Params.Amount = N` and `Params.Player` = the defending player at the time it triggered.
4. **The resolution.** The body reads the defending player again: if the source is still the same object and still attacking, `DefendingPlayerForAttackerForEffect`; otherwise the recorded player (CR 508.5's "before it was removed from combat"). That player chooses min(N, permanents they control) permanents in one prompt, and they are sacrificed as one simultaneous exit. A player who has left the game is skipped (CR 800.4a). The source leaving the battlefield does not stop the ability: it has no "if" clause.
5. **Several instances.** Each instance is its own trigger (CR 702.86b). They are not marked `Commutes`: the attacking player orders them (CR 603.3b), because each sacrifice's dies triggers land between them. Two instances on one creature share a label, so the existing same-source, same-label rule skips a prompt that means nothing.
6. **Ulamog, the Defiler.** "Ulamog has annihilator X" has no fixed N, so it is not a token. Its catalog file declares a triggered row through `effects.AnnihilatorCounted(label, count)`, which uses the same `annihilator/sacrifice` body with the amount read at resolution: the +1/+1 counters on Ulamog, or on its last-known information if it has left (CR 608.2h).
7. **The Eldrazi shuffle.** `effects.WhenThisIsPutIntoAGraveyardFromAnywhere`, the library-mill helper's shape widened to every source zone (battlefield, hand, library, stack, exile), scoped to the graveyard and over `finaleShuffleGraveyardIntoLibrary`. "Its owner shuffles" is the owner of the card, which is who the graveyard belongs to.

### Snapshot impact

A new keyed body, `annihilator/sacrifice`: an on-disk identity, never renamed or reused. A binary from before this change refuses a restore point naming it (`ErrUnknownEffectKey`), which is the rollback case. No shape change: `Params.Player` and `Params.Amount` already exist. While the defending player is choosing, the table holds the same pick prompt Lotus Field's does, with the same restore-point class. A new corpus board, `annihilator_trigger_pending.json`.

### Wire, client and bot

The token reaches the card view as `annihilator 4`, and the client's keyword badge row shows its three-letter text badge (no icon, like toxic). The defending player's choice is the existing pick dialog. The enumerator already answers Lotus Field's choose-two prompt; a choose-four is the same prompt kind, so the legal-move list needs no change. The tests check that the move it offers is one set ordered cheapest first (ADR 0020 addendum §15), not every combination. The attack heuristic does not learn to value annihilator in this PR.

### Cards

- **Kozilek, Butcher of Truth and Ulamog, the Infinite Gyre: Full.** Cast trigger (Kozilek, the Great Distortion's shape; the Gyre's "destroy target permanent"), indestructible for the Gyre, the token, and the shuffle trigger.
- **Ulamog, the Defiler: Full.** Cast trigger (Peer into the Abyss's half, rounded up, on a target opponent), `WardSacrificeN(2)`, entry counters equal to the greatest mana value in exile, and `AnnihilatorCounted`.
- Creatures whose only text is annihilator and other canonical keywords need no card file (Ulamog's Crusher). The rest of the pool follows ADR 0106 owner decision 6, subject to [question 2](#open-questions-for-the-owner): Artisan of Kozilek, Pathrazer of Ulamog, It That Betrays, Eldrazi Conscription and the others.

### Tests

- One trigger per instance; a granted `annihilator 2` on a printed `annihilator 4` triggers twice.
- No trigger for a creature put onto the battlefield attacking.
- Four players, three attackers with annihilator at three different players: each trigger's defending player is its own attacker's.
- The defending player has fewer than N permanents; sacrifices a planeswalker that is being attacked; or loses the game first.
- The attacker is removed from combat, or leaves the battlefield, before the trigger resolves.
- Ulamog, the Defiler with a +1/+1 counter added in response.
- Kozilek milled, discarded, killed and countered: the graveyard is shuffled in each case.
- A snapshot round trip with the trigger on the stack and with the pick open.

### Registry

A new keyword row: `annihilator`, **Implemented**, `Rules` 702.86, 508.3a, 508.5, `Keywords` `[game.KeywordAnnihilator]`, `Probe` `hasKeyword(game.KeywordAnnihilator + " ")`, `Printed` `printedLine("Annihilator")`, `Issue` 2073, this ADR.

---

## 3. Statics that set or change a maximum hand size (#2074)

### What exists, what is missing

- **One shape exists.** `EffectiveMaxHandSizeLocked` (`game/effect_hooks.go`, #338) returns `NoMaxHandSize` if the player's own `Player.MaxHandSize` is `NoMaxHandSize`, or if they control a permanent whose catalog entry sets `Spec.NoMaxHandSize` (with its designation gate, `NoMaxHandSizeWhen`). Otherwise it returns `Player.MaxHandSize`. Nothing reads an opponent's permanent, nothing sets a number, nothing modifies one, and nothing is ordered.
- **The player-level grant has no timestamp.** `SetMaxHandSizeForEffect` writes `Player.MaxHandSize` for Finale of Revelation and Sea Gate Restoration ("for the rest of the game"), and the sandbox `set_max_hand_size` action writes the same field. The read checks it first, so it always wins.
- **The consumers are right already.** The cleanup discard (`populateDiscardPendingLocked`, `game/game.go`) asks for the ACTIVE player's effective maximum only, and the player view's `max_hand_size` is the same call. Both treat −1 as "no maximum".
- **Timestamps exist.** `Card.layerTimestamp()` is the CR 613.7 timestamp the layer engine sorts by.

### The rules

- **CR 402.2:** "Each player has a maximum hand size, which is normally seven cards."
- **CR 514.1:** "if the active player's hand contains more cards than their maximum hand size … they discard enough cards to reduce their hand size to that number." Only the active player, against their own maximum. The Jin-Gitaxias ruling (2017-11-17): the maximum "isn't checked at any time other than their own cleanup step", and Jin-Gitaxias "doesn't affect the maximum hand size of its controller".
- **CR 613.11:** effects that modify a player's maximum hand size "are applied after all other continuous effects", and "in timestamp order". The Null Profusion ruling (2009-10-01) gives the example: Spellbook, then Null Profusion, is a maximum of two; the other order is no maximum.
- **CR 613.7a / 613.7b:** a static ability's effect has its object's timestamp; a resolved spell's effect has the timestamp of its creation.
- **CR 107.1b:** a calculation uses a negative value if it needs to, and a result that "would determine the result of an effect" is zero if negative. So Jin-Gitaxias twice is 7 − 7 − 7 = −7, and that player discards their whole hand. The Jin-Gitaxias ruling: alone, it makes a maximum of zero.
- **CR 102.3:** opponents are all players not on your team. In free-for-all Commander that is every other player.

### Decision

1. **The declaration.** `Spec.HandSize []game.HandSizeStatic`, a new `Spec` slot (the four edits of docs/adding-cards.md "Adding a `Spec` slot"). `HandSizeStatic{Players, Kind, N, When}`:
   - `Players`: `HandSizeYou`, `HandSizeEachOpponent` or `HandSizeEachPlayer`, relative to the permanent's controller.
   - `Kind`: `HandSizeNoMaximum`, `HandSizeSet` (to N ≥ 0) or `HandSizeModify` (by N ≠ 0).
   - `When`: the designation gate, as `NoMaxHandSizeWhen` has now.

   `Spec.NoMaxHandSize` stays as the card-side shorthand for eight cards; `buildDef` folds it into the list as `{HandSizeYou, HandSizeNoMaximum, When: NoMaxHandSizeWhen}`. `Register` refuses a `Set` with N < 0 and a `Modify` with N = 0. `game.CatalogNoMaxHandSize` becomes `game.CatalogHandSize(key) []HandSizeStatic`.
2. **The fold.** `EffectiveMaxHandSizeLocked(p)` collects every entry that applies to `p`: each battlefield permanent's statics, read through `CatalogAbilityKey` (so a permanent that has lost its abilities gives nothing, CR 613.1f) with its gate, at the permanent's `layerTimestamp()`; and the player's own grant, when `Player.MaxHandSize` is not the default, at `Player.MaxHandSizeAt`. It sorts them by timestamp, with the layer engine's tie-break. It folds from 7: `NoMaximum` makes it unbounded, `Set` makes it N, and `Modify` adds N to a number and leaves "unbounded" unbounded. Intermediate values are not clamped (CR 107.1b). The result is `NoMaxHandSize` or max(0, value), so the −1 sentinel can never be produced by arithmetic.
3. **The player grant's timestamp.** `Player.MaxHandSizeAt int64` (`json:"maxHandSizeAt,omitempty"`), stamped by `SetMaxHandSizeForEffect` and by the sandbox action from the same clock as `EnteredBattlefieldAt`. Finale of Revelation's grant then sorts by when it resolved (CR 613.7b), so a Null Profusion that enters after it sets two. A restore point written before this field has no stamp, so its grant sorts first. That changes an answer only for a player who has a Finale grant and controls, or is the opponent of, a set or reduce permanent, which no known restore point holds.
4. **Which player discards.** Unchanged: the active player, against their own effective maximum (CR 514.1). Jin-Gitaxias's controller is never reduced.

### Snapshot impact

`players[].maxHandSizeAt` is additive within v7, recorded with `-update-shape`. The `Spec` slot is catalog data and is never serialised. Nothing new is reachable from `Game` that holds a func, so the closure ratchet does not move. A new corpus board, `max_hand_size_grant.json`: a player with a stamped Finale grant.

### Wire, client and bot

No protocol change: `max_hand_size` is still the effective value, −1 or a number from 0 up. The client shows it only in the discard prompt today; whether the seat panel should show a changed maximum is [question 1](#open-questions-for-the-owner). The bot answers the cleanup discard prompt as it does now.

### Cards

- **Jin-Gitaxias, Core Augur: Full.** Flash, the end-step draw seven, and `{HandSizeEachOpponent, Modify, −7}`.
- **Null Profusion: Full.** "Skip your draw step" (`skip_draw_step.go`), "whenever you play a card, draw a card" (a land played or a spell cast, never a copy: the 2007-02-01 ruling), and `{HandSizeYou, Set, 2}`. The PR verifies that the play-a-card trigger exists without Prosper's from-exile filter; if it doesn't, the PR adds it.
- **Price of Knowledge: Full.** `{HandSizeEachPlayer, NoMaximum}` and the each-opponent upkeep damage equal to that player's hand size, counted at resolution (2013-10-17 ruling).
- The rest of the pool follows ADR 0106 owner decision 6, subject to [question 2](#open-questions-for-the-owner): Gnat Miser, Locust Miser, Cursed Rack (a chosen opponent, set to four), Anvil of Bogardan and the others.

### Tests

- The Null Profusion ruling in both orders, with Spellbook.
- Jin-Gitaxias: each opponent at 0, the controller at 7; two Jin-Gitaxias give −7, which the view reports as 0 and the cleanup reads as "discard everything".
- Jin-Gitaxias then Reliquary Tower is no maximum; Reliquary Tower then Jin-Gitaxias is also no maximum.
- Price of Knowledge lifts every player's maximum, including an opponent under Jin-Gitaxias, in either order.
- Finale of Revelation's grant, then a Null Profusion: two.
- A Null Profusion that has lost its abilities gives nothing.
- A snapshot round trip with a stamped grant.

### Registry

A new seam row: `max-hand-size-statics`, "Maximum hand size", **Implemented**, `Rules` 402.2, 514.1, 613.11, `Issue` 2074, this ADR, `Examples` Jin-Gitaxias, Core Augur, Null Profusion and Price of Knowledge.

---

## 4. Undying and persist (#2075)

### What exists, what is missing

- **Neither keyword exists.** `persist.go` is the sorcery Persist. Murderous Redcap ships with the caveat "Persist is not implemented".
- **The leaves-the-battlefield harvest reads only the catalog.** `harvestLTB` (`game/triggers.go`) takes the triggers from `TriggersForKey(oracle, source)` and returns early when the key is empty. `keywordTriggersFor` is not asked, so a keyword trigger on a dying permanent cannot fire, and a token with no catalog key fires nothing at all.
- **Last-known counters exist.** `snapshotLKILocked` writes `lastKnownCounters` before every battlefield exit, and `harvestLTB` deletes it only after the harvest, so a trigger's `AppliesTo` can read what the permanent had. `PermanentInfo.Counters` keeps the same counters for the rest of the turn.
- **Return with counters exists.** `ReturnFromGraveyardWithCountersForEffect` puts the counters on the entry event, so a CR 614 counter replacement sees them (Perennation, Makeshift Mannequin). `returnThisCreatureFromGraveyard` (`duration_grants.go`) is the nearest body, but it does not check that the card in the graveyard is still the same object.
- **SBA 704.5q exists** (`game/mutations.go`).

### The rules

- **CR 702.93a:** "'Undying' means 'When this permanent is put into a graveyard from the battlefield, if it had no +1/+1 counters on it, return it to the battlefield under its owner's control with a +1/+1 counter on it.'"
- **CR 702.79a:** persist is the same with -1/-1 counters.
- **CR 603.10a:** leaves-the-battlefield abilities look back in time: whether the permanent had the ability, and what counters it had, is read as it last existed on the battlefield. The persist rulings: a creature with +1/+1 counters that gets enough -1/-1 counters to die still had -1/-1 counters "at that point", so persist doesn't trigger.
- **CR 603.4:** the intervening "if" is checked when the event occurs and again on resolution. Here both checks read the same last-known information, so the second can't fail.
- **CR 603.3a:** the trigger is controlled by whoever controlled the permanent when it left. The card returns under its OWNER's control.
- **CR 400.7e:** an ability that triggers on a zone change "can find the new object that it became in the zone it moved to when the ability triggered". A card that has left the graveyard since is a new object, and the rulings say it won't be returned.
- **CR 113.2c:** "If an object has multiple instances of the same ability, each instance functions independently." The Redcap ruling: each triggers, and once one has returned the card the rest do nothing.
- **CR 111.7 / 704.5d:** a token in the graveyard ceases to exist. The ability triggers, but there is nothing to return (Redcap ruling, 2013-06-07).
- **CR 122.6:** counters an object is given as it enters count as being put on it, so Hardened Scales and Vizier of Remedies apply to the returning creature.
- **CR 704.5q:** if a permanent has both +1/+1 and -1/-1 counters, N of each are removed. A creature that came back with undying's +1/+1 counter and then got a -1/-1 counter has neither, so undying can return it again. This is the Mikaeus, the Unhallowed and Triskelion loop, and the persist ruling (2024-09-20) says the same of persist.

### Options

- **A. Engine keyword tokens, like prowess and evolve (chosen).** `undying` and `persist` join `canonicalKeywords`. The leaves-the-battlefield harvest derives one trigger per token from the permanent's LAST-KNOWN ability list. A printed instance (stamped by the importer from Scryfall's array, no catalog entry needed), a token template's and a layer-6 grant (Mikaeus, the Unhallowed; Cauldron Haze; Rhys, the Evermore) all work. A Clone that copied a persist creature has the token in its last-known list (layer 1), so it returns, as itself, and copies again.
- **B. A catalog constructor, `effects.Persist()`.** Every card needs a file, and the 14 cards that grant one of the two keywords reach nothing. It is less faithful to how the cards are printed and granted, for the reason ADR 0106 §3 gives for evolve.

### Decision (A)

1. **The tokens.** `KeywordUndying = "undying"` and `KeywordPersist = "persist"` join `canonicalKeywords`, both cumulative (CR 113.2c), so a granted instance on a printed one is two triggers.
2. **The harvest.** `harvestLTB` asks `ltbKeywordTriggersFor(lki)` for the keyword triggers in the last-known characteristic's ability list, before the empty-key early return, and appends the catalog's triggers when there is a key. It reads the LKI list, never the card, so a creature that died having lost all abilities (CR 613.1f) has none, and one that died wearing a grant has it (CR 603.10a).
3. **The condition.** `AppliesTo` holds when the event put the card into a graveyard (`NewZone == ZoneGraveyard`; a creature exiled instead by Rest in Peace triggers nothing) and `lastKnownCounters[card]` has no `+1/+1` counters (undying) or no `-1/-1` counters (persist). The permanent need not be a creature (Redcap ruling: persist works on a permanent that stopped being one).
4. **The item.** A keyed item, body `undying/return` or `persist/return`, built for the LKI controller (CR 603.3a). Its `SourceObject` is the card's new graveyard object, the object CR 400.7e lets it find.
5. **The resolution.** The body re-reads the condition off `PermanentInfo.Counters` for the departed object (CR 603.4; the same answer). Then, if the card is in a graveyard and its `ObjectEpoch` still matches `SourceObject`, it returns the card under its owner's control with one counter of the keyword's kind on the entry event. Otherwise it does nothing, and `ErrCardNotFound` is swallowed, so a token or a moved card is not an `EventEffectError`.
6. **Ordering.** The triggers are not marked `Commutes`. Several creatures returning in different orders get different timestamps and different entry triggers, so CR 603.3b's choice is a real one. Two instances on one creature share a source and a label, so the existing rule skips that prompt.
7. **Commanders.** A commander with persist that dies triggers. If its owner moves it to the command zone (CR 903.9a), the graveyard object is gone and the trigger does nothing. The engine's existing commander answers decide which happens; this seam does not change them.

### Snapshot impact

Two new keyed bodies, `undying/return` and `persist/return`: on-disk identities, refused by an older binary (the rollback case). No shape change. A new corpus board, `persist_trigger_pending.json`.

### Wire, client and bot

The tokens reach the card view, and the badge row shows the three-letter text fallback. No prompt is added: the return has no choice. No legal-move change. Teaching the bot's sacrifice-fuel order to prefer a persist creature with no -1/-1 counter is out of scope.

### Cards

- **Gleeful Arsonist: Full.** The opponent-casts-a-noncreature-spell trigger is Kambal, Consul of Allocation's shape; the damage is its power at resolution, or its last-known power (2024-09-20 ruling). Undying is `PrintedKeywords`.
- **Persistent Constrictor: Full.** At each opponent's upkeep, that player loses 1 life and up to one target creature they control gets a -1/-1 counter. Per the 2024-09-20 ruling, an illegal target means no life is lost either. Persist is `PrintedKeywords`.
- **Murderous Redcap: Caveats → Full.** The caveat goes, and `PrintedKeywords` gains `persist`.
- Creatures whose only text is undying or persist and other canonical keywords need no file (Young Wolf, Butcher Ghoul, Safehold Elite). The rest of the pool follows ADR 0106 owner decision 6, subject to [question 2](#open-questions-for-the-owner): Kitchen Finks, Glen Elendra Archmage, Strangleroot Geist, Geralf's Messenger, Mikaeus, the Unhallowed, Cauldron Haze and the others.

### Tests

- Undying returns a creature with a +1/+1 counter; persist with a -1/-1 counter; neither triggers when the creature had that counter.
- 704.5q: an undying creature that came back and then got a -1/-1 counter has no counters, dies again and returns again. A persist creature with a +1/+1 counter that dies to -1/-1 counters does not return (the ruling).
- A token with persist triggers and returns nothing, with no `EventEffectError`.
- A nontoken Clone copying Kitchen Finks returns as a Clone and chooses again.
- The card is exiled from the graveyard in response: nothing returns. It is exiled and put back: nothing returns (CR 400.7).
- A stolen Redcap returns under its owner's control; the trigger was the thief's.
- Hardened Scales: undying returns it with two +1/+1 counters.
- Two instances: the second does nothing.
- A board wipe with persist creatures on two sides: APNAP order.
- A snapshot round trip with a persist trigger on the stack.

### Registry

Two new keyword rows, `undying` (Rules 702.93) and `persist` (Rules 702.79), both **Implemented**, both with 603.10a and 400.7e, `Keywords`, `Probe` `hasKeyword(…)`, `Printed` `printedKeyword(…)`, `Issue` 2075, this ADR. Murderous Redcap's `Caveats` entry goes.

---

## Delivery

One PR per seam. Each lands its engine change, its registry rows, its corpus board and its cards together, and each closes its issue.

1. **#2072, the sacrificed-objects record.** `PaidCost.SacrificedObjects`, the three fill sites, the two copy paths (with the `Sacrificed` copy fix), the readers, the `b17PermanentSacrificedToPay` migration, Fling, Kazuul's Fury, Momentous Fall, Tend the Pests and Corpse Cobble. The pointer line in ADR 0100.
2. **#2074, maximum hand size.** `HandSizeStatic`, the fold, `Player.MaxHandSizeAt`, Jin-Gitaxias, Null Profusion and Price of Knowledge.
3. **#2073, annihilator.** The token, the importer, the trigger and its body, `AnnihilatorCounted`, the from-anywhere shuffle trigger, Kozilek, both Ulamogs.
4. **#2075, undying and persist.** The tokens, the harvest change, the two bodies, Gleeful Arsonist, Persistent Constrictor, Murderous Redcap.

PRs 1 and 2 touch nothing the others touch, so they can go in parallel with everything. PRs 3 and 4 both add to `canonicalKeywords`, `keywords.go`'s cumulative list and the registry, so whichever lands second merges `develop` in first. Each PR follows AGENTS.md §7: oracle fixtures for its own cards only, `go test ./internal/roadmap/ -update` for the docs, and every card it skips goes on its row's `Waiting` list with the reason.

## Consequences

- A payment record now names every permanent its sacrifice took. Any new "the sacrificed …" reader uses `ctx.SacrificedPermanents()`, never the event log.
- Annihilator, undying and persist are the third, fourth and fifth keywords derived from the ability list. The leaves-the-battlefield harvest now reads keyword triggers too, so a later dies keyword (modular's second half, for one) has a place to go.
- A maximum hand size is a timestamp-ordered fold, not a flag. A future "your maximum hand size is increased by one" is one more `HandSizeStatic`.
- Spell copies now carry the sacrifice count, which changes a copied Vicious Betrayal, Devouring Greed and Devouring Rage from zero to the original's count (CR 707.10).

## Out of scope

- Phyrexian Obliterator's "that many permanents" is still asked as N one-at-a-time prompts (`b17PlayerSacrificesN`). §2 uses one choice and one simultaneous exit. Moving the older helpers onto §2's shape is a separate change.
- An alternative cost's sacrificed permanents (§1 decision 1), and a permanent that reads what the spell that became it sacrificed (CR 400.7d). No printed card in either pool needs them.
- Emerge, devour and the other costs that sacrifice and read a creature as part of a keyword. Devour has its own row.
- Teaching the attack heuristic to value annihilator, or the sacrifice-fuel order to value persist and undying.

---

## Open questions for the owner

1. **Showing a changed maximum hand size (§3).** Today a player sees their maximum only in the cleanup discard prompt. With Jin-Gitaxias out, every opponent's maximum is zero.
   - **(a) Recommended:** the seat panel shows a small badge when a player's effective maximum is not seven ("Hand max 0", "No hand max"), from the existing `max_hand_size`. No protocol change.
   - (b) No client change. Players learn it at cleanup.
2. **How many cards each PR lands.** ADR 0106 owner decision 6 is "the named waiting cards plus every other card of the seam that needs no other missing primitive". Here the pools are 34 sacrifice-reading spells (§1), 18 annihilator cards (§2), 23 hand-size cards (§3) and 62 undying or persist cards (§4), though the keyword-only creatures in §2 and §4 need no file at all.
   - **(a) Recommended:** apply decision 6 as before. Each PR lands its named cards plus every pool card that needs nothing else, each verified against its full oracle text; anything that needs more goes on a `Waiting` list with the reason.
   - (b) Each PR lands only the named cards (and Corpse Cobble and the Redcap's caveat). The rest of each pool follows in one card batch after the four seams have merged.
