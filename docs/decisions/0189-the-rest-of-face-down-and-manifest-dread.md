# ADR 0189 — The rest of face-down cards and manifest dread: manifesting from a hand, both ways back up, survival, and a graveyard-exile cast cost

Date: 2026-10-14
Status: Accepted (owner decisions 2026-10-14, recorded below)
Sprint: none (engine seam ADR, Backlog)
Issues: #3523 (this seam: what is left of the face-down and manifest dread rows), #1948 (exile a card from your graveyard as an additional cost). History: #886, #95, #658, #1194, #2570, #2590 (all closed).
Builds on: [ADR 0069](0069-face-down-objects.md) (the face-down object model), [ADR 0082](0082-casting-face-down-and-turning-face-up.md) (the turn-face-up offer, decision 5, and the manifest dread amendment), [ADR 0059](0059-turn-machinery.md) Decision 8 (`Turn.PhaseOrdinal`), [ADR 0021](0021-additional-costs.md) and #3402 (a spell's additional costs, and the shape #3412 used to put an ability's cost component on a cast), [ADR 0062](0062-abilities-and-special-actions-from-the-hand.md) Decision 4 (one special-action verb).
Numbering: 0189 was reserved for this draft. Before pushing, every remote branch (`origin/develop`, `origin/main`, and the 11 feature branches plus `pr/2326`) was listed with `git ls-tree --name-only -r <ref> docs/decisions/` after `git fetch --all --prune`, along with the files of every open PR. The highest number in use is 0185. 0186–0188 are reserved for sibling drafts.

## Context

Two roadmap rows are still partial, and both of their tracking issues are closed:

| Row (`registry.go`) | Waiting | Tracked |
|---|---|---|
| Face-down cards (`face-down-objects`) | Kozilek, the Broken Reality | #886 / #95 / #658, closed |
| Manifest dread (`manifest-dread`) | Defiant Survivor, Abhorrent Oculus | #2570, closed |

Each row lists only these cards. I swept every Commander-legal card whose oracle text manifests, cloaks, manifests from a hand, triggers at "your second main phase", or exiles cards from a graveyard as an additional cost. That sweep turned up the cards named under each decision below as later users. None of them is on another row.

### The cards (Scryfall oracle text)

**Kozilek, the Broken Reality** {9}, Legendary Creature — Eldrazi 9/9:
"When you cast this spell, up to two target players each manifest two cards from their hands. For each card manifested this way, you draw a card. (To manifest a card, put it onto the battlefield face down as a 2/2 creature. Turn it face up any time for its mana cost if it's a creature card.) Other colorless creatures you control get +3/+2."

Rulings (2024-06-07): "If a targeted player has only one card in hand as Kozilek's triggered ability resolves, they manifest that card." "If a manifested creature would have morph or disguise if it were face up, you may also turn it face up by paying its morph cost or disguise cost." The trigger resolves even if Kozilek is countered.

**Defiant Survivor** {2}{G}, Creature — Human Survivor 3/2:
"Survival — At the beginning of your second main phase, if this creature is tapped, manifest dread."

Rulings (2024-09-20): "Survival abilities (and other abilities that trigger at the beginning of your second main phase) will trigger at the beginning of the second main phase you take in a turn. They won't trigger during your third, fourth, or other additional main phases." "If a creature's survival ability triggers but the creature leaves the battlefield before the ability resolves, use its tapped or untapped status as it last existed on the battlefield."

**Abhorrent Oculus** {2}{U}, Creature — Eye 5/5:
"As an additional cost to cast this spell, exile six cards from your graveyard. Flying. At the beginning of each opponent's upkeep, manifest dread."

Ruling (2024-09-20): "Abhorrent Oculus's additional cost must be paid even if it's cast 'without paying its mana cost' or for any alternative cost."

### The rules (CR effective 2026-09-25)

- **701.40a** To manifest a card, turn it face down, then put it onto the battlefield as a 2/2 face-down creature card with no text. **701.40b** A manifested *creature card* can be turned face up for its mana cost. **701.40c** "If a card with morph is manifested, its controller may turn that card face up using either the procedure described in rule 702.37e … or the procedure described above." **701.40d** says the same for disguise (702.168d). **701.40e** "If an effect instructs a player to manifest multiple cards **from their library**, those cards are manifested one at a time." **701.40g** A manifested instant or sorcery is revealed and stays face down.
- **701.58c/d** are the same as 701.40c/d, for cloak.
- **701.62a** Manifest dread.
- **505.1a** "Only the first main phase of the turn is a precombat main phase. All other main phases are postcombat main phases." **505.1b** "Phrases such as 'first main phase,' 'second main phase,' and so on count the number of main phases that have occurred only in the current turn."
- **603.4** The intervening "if" is checked when the trigger event happens and again on resolution.
- **601.2a, 601.2f–h, 118.3** Additional costs are paid with the spell already on the stack. A cost can't be paid without the resources to pay it fully.

### What already exists (develop at f943f6a4d)

- **Face-down objects and every face-down kind** (ADR 0069, ADR 0082, #1194, #1209, #1270, #1271): morph, megamorph, disguise, manifest, cloak and "turned". `ZoneEntryOptions.FaceDown` (`game/battlefield_put.go`) lifts the CR 110.4 refusal for a nonpermanent and stamps the face-down state and its viewers. This works for any source zone the entry batch takes, the hand included.
- **Manifest from a library**: `Game.ManifestForEffect`. Its doc comment still says "MANIFEST THE MECHANIC IS NOT SHIPPED", which is stale.
- **Manifest dread**: `Game.ManifestDreadThenForEffect` (`game/manifest_dread.go`), and `effects.ManifestDread{Then}`.
- **Turning face up as an effect**: `Game.TurnFaceUpForEffect` (#2590).
- **The turn-face-up offer**: `game.TurnFaceUpOffer` (`game/face_down.go`) returns one offer per permanent. For a manifested or cloaked permanent it offers only the mana-cost route, and its doc comment says CR 701.40c is out of scope (ADR 0082 decision 5).
- **Picking one of two offers of the same kind**: `SpecialActionParams.Cost` (#1391, Fblthp's plot) and `specialActionOfferLocked`'s cost filter. The enumerator sends `cost` only when `sa.Zone != ""` (`legal/special_actions.go`), and the client's menu id is `special-<kind>[-<door>]` (`contextMenu.logic.ts specialActionItems`).
- **Hidden per-seat choices**: `Game.ChooseCardsRunThenForEffect` (`game/choose_cards_run.go`, Explore the Vastlands) asks one `choose_cards` prompt per seat as one run and continues once.
- **Simultaneous entry from hands**: `Game.PutCardsFromHandOntoBattlefieldThenForEffect` (ADR 0146 amendment, Selvala's Stampede). With `Controller` unset, each card enters under its owner.
- **Cast triggers**: `effects.WhenYouCastThisSpell` (`annihilator.go`). "Up to two target players" is `TargetPlayer(…).WithCount(0, 2)` (Court of Cunning uses `(0, 0)`, Axis of Mortality `(2, 2)`). There is a colorless-creature anthem (Forsaken Monument).
- **Second-main-phase triggers**: `EventStepBegan` with `StepPostcombatMain` (`effects.StepBegan`, `AtYourPostcombatMain`). `Turn.PhaseOrdinal` counts the phases of a family begun this turn, with precombat and postcombat main as one family (`game/turn_plan.go`). It is set by `noteStepBegunLocked` before `EventStepBegan` is emitted (`game/game.go`). The manifest dread row's note ("`EventBeginPrecombatMain` is the only main-phase event") is stale. Scheming Silvertongue already prints "your second main phase" and is built on `StepBegan(StepPostcombatMain, true)`, so it also fires in a third main phase.
- **"If this permanent is tapped" at resolution, with last-known information**: `ctx.SourcePermanent()` (`effects/context.go`, #1418) returns `PermanentInfo.Tapped`, as it last existed once the source has left.
- **An "Exile N cards from your graveyard" cost component**: `game.ExileCost` (`game/exile_cost.go`), on activated and mana abilities only (`AbilityCost.ExileCards`, `effects.ExileCardsFromGraveyard`). Its candidate walk `ExileCostOptionsForEffect`, validator `validateExileCardsCostLocked`, payer `payExileCardsCostLocked`, the view's `exile_cost_n/label/options/zone`, and the client picker (`client/src/lib/exileCost.ts`) all exist. `PaidCost.Exiled` and `ctx.Exiled()` exist too. #3412 put an ability's component (`ExilePermanentsCost`) on a spell's mandatory slot, and this design follows the same route.

### The gaps

1. **Manifest from a hand** (Kozilek): no door or effect turns "each of these players chooses N cards from their hand" into one face-down entry.
2. **Both ways back up** (CR 701.40c/d, 701.58c/d): a manifested or cloaked card with morph or disguise can't use its keyword cost. This is weaker than printed for every manifest, manifest dread and cloak card already shipped. Kozilek makes it common, because it manifests other players' hands, and morph and disguise decks hold their morphs in hand.
3. **"Your second main phase"**: no condition counts main phases (Defiant Survivor's survival, and Scheming Silvertongue).
4. **A graveyard-exile cast cost** (Abhorrent Oculus, plus #1948's Draconic Intervention and Relentless Skaabs): `game.AdditionalCost` has no `ExileCost` component.

## Decisions

### 1. "Your second main phase" is a condition on the step event; survival is a wrapper

- `effects.YourSecondMainPhaseBegan` is a `When`. It holds when `StepBegan(game.StepPostcombatMain, true)` holds and `g.Turn.PhaseOrdinal == 2`. In a turn whose combat was skipped, the postcombat main is still main phase 2 (CR 505.1a). An added main phase is ordinal 3 or later and does not fire it, which is the ruling.
- `effects.AtYourSecondMainPhase(label, effect)` is `On(game.EventStepBegan, YourSecondMainPhaseBegan, label, effect)`.
- `effects.Survival(label, effect)` is "At the beginning of your second main phase, if this creature is tapped, …". Its `AppliesTo` is `YourSecondMainPhaseBegan` together with `ThisIsTapped`, which is the first CR 603.4 check. Its effect wrapper reads `ctx.SourcePermanent()` and does nothing unless `ok && info.Tapped`. That is the second check, and once the source has left it uses last-known information, which is the fourth ruling. The effect runs only after that check passes.
- Scheming Silvertongue moves from `StepBegan(StepPostcombatMain, true)` to `YourSecondMainPhaseBegan`. That is a correctness fix for turns with an extra main phase. Cards that print "each of your postcombat main phases" (Neheb, the Eternal, Sorin of House Markov, Megatron, Sphinx of the Second Sun) keep `StepPostcombatMain` and are correct as they are.

**Rejected:** a new `EventBeginSecondMain` kind. `EventStepBegan` is already the vocabulary for "at the beginning of <step>", and an event kind per ordinal multiplies kinds that every harvester has to know. **Rejected:** keeping `AtYourPostcombatMain`. It fires in a third main phase, which the ruling rules out.

### 2. "Exile N cards from your graveyard" as a spell's additional cost: the ability's `ExileCost` on the mandatory slot

This is #3412's shape, one component over:

- **Engine**: `game.AdditionalCost.ExileCards *ExileCost`. It is graveyard only (`From == ZoneGraveyard`) with a fixed `N`. `effects.Register` refuses the hand form, a non-positive N, an optional cost and an either/or branch, so nothing ships that no waiting card prints. The constructor is `effects.ExileFromGraveyardCost(n, label, match)`, which returns `*game.AdditionalCost{ExileCards: ExileCardsFromGraveyard(n, label, match), Label: "Exile " + label + " from your graveyard"}`.
- **Announce**: the caster names the cards on `CastSpellParams.ExileIDs`, with wire name `exile_ids` (the name activations already use). A new file, `game/exile_card_additional_cost.go`, holds `planExileCards(plan)` and `validateCastExileCardsLocked`. The latter calls `validateExileCardsCostLocked` with `alsoSpent` = every card another component of the same cast spends (`DiscardIDs`, `DelveIDs`, `AltCostIDs`). One card pays one component (CR 118.3). The spell itself is never a candidate, because CR 601.2a has already moved it to the stack. That covers Skaab Ruinator cast from its own graveyard.
- **Pay**: `payExileCardsCostLocked`, with the spell already on the stack, beside the sacrifices (CR 601.2h). The cards stay exiled if the spell is countered. A graveyard-leaves trigger lands above the spell.
- **Record**: `PaidCost.Exiled`, so `ctx.Exiled()` reads it at resolution. Draconic Intervention's "the exiled card's mana value" and Corpse Lunge's "power" read the cards in exile. A CR 707.10 copy keeps the record, as `spell_copy.go` already does for `Exiled`.
- **Free and alternative casts**: the mandatory slot is charged under every alternative cost and every "without paying its mana cost", which matches the Oculus ruling. No new code is needed. This is a test assertion.

**Rejected:** reusing `ExilePermanentsCost.FromGraveyard`. That is craft's "this artifact and N other permanents and/or graveyard cards" form, which mixes zones, and `Register` already refuses it on a spell. **Rejected:** the alternative-cost `ExileFromGraveyard` `TargetSpec`. It replaces the mana cost (escape), and this cost is added to it. **Rejected:** exiling in `OnResolve`. That is the wrong time (CR 601.2h), and a countered spell would keep its fuel.

Out of scope, and left on #1948's row: the X forms (Skeletal Scrying, Harvest Pyre, Chill Haunting, Haunting Misery), and Soaring Stoneglider's either/or branch ("exile two cards from your graveyard or pay {1}{W}").

### 3. Manifest from a hand: one door, one effect, one hidden run

- **Engine**: `Game.ManifestFromHandsThenForEffect(ids []uuid.UUID, then func(g *Game, manifested []uuid.UUID) error) error` (`game/battlefield_put.go`, beside `ManifestForEffect`). Every ID must be in its owner's hand, or the whole call is refused with `ErrCardNotFound` and nothing moves. It is one `PutCardsFromHandOntoBattlefieldThenForEffect` with `HandEntryOptions{FaceDown: FaceDownManifested}` and `Controller` unset, so each card enters under its owner, the player who manifests it. `then` is told which cards entered and runs exactly once. CR 701.40e's "one at a time" covers only a library. Cards from a hand enter together, as Selvala's Stampede's do, so a CR 614 window sees the board as it was before any of them entered. The stale "NOT SHIPPED" paragraph on `ManifestForEffect` is rewritten in the same change.
- **Effect**: `effects.EachManifestsFromHand{Players func(*Context) []uuid.UUID; N int; Then func(*Context, []uuid.UUID) error}` (`effects/manifest_from_hand.go`). For each player, in APNAP order, it builds one `ChooseCardsPrompt` leg: chooser = that player, `Cards` = their hand, `Min = Max = min(N, hand size)`, `Zone = ZoneHand`, `Destination = game.ChooseManifest`, question "Choose N cards from your hand to manifest". A player whose hand holds N cards or fewer is not asked: they manifest all of them (the Kozilek ruling). The other legs go up together through `ChooseCardsRunThenForEffect`, so no seat learns another's picks first (CR 101.4a). The run's continuation calls `ManifestFromHandsThenForEffect` with every pick, then `Then`.
- `game.ChooseManifest ChooseDestination = "manifest"` joins the open vocabulary (`game/chained_choice.go`). The client shows the verb "Manifest". The heuristic scores it as giving a card up (`valueKeptInHand`, so the worst cards are named), not as `valuePutOntoBattlefield`. A face-down 2/2 is not the card.
- **Kozilek, the Broken Reality**: `WhenYouCastThisSpell` with `TargetPlayer("up to two target players").WithCount(0, 2)`. The effect is `EachManifestsFromHand{Players: still-legal targets, N: 2, Then: the trigger's controller draws len(manifested)}`, so a card that a replacement kept off the battlefield is not counted. The static is a layer-7c +3/+2 anthem over other colorless creatures you control, Forsaken Monument's filter minus the source.
- **Scroll of Fate** ("{T}: Manifest a card from your hand.") is the second user: `EachManifestsFromHand{Players: you, N: 1}`.

**Rejected:** asking each player in turn. The picks would become public one by one and would leak to the later choosers. **Rejected:** choosing for the player. The player chooses which cards (CR 701.40a, and the ruling's "they manifest"). **Rejected:** a hand variant of `ManifestForEffect`, one card at a time. It would give a later card's CR 614 window a different board than an earlier card's, with no rule asking for that.

### 4. Both ways back up: `TurnFaceUpOffers` returns every route

- `game.TurnFaceUpOffers(c Card) []SpecialAction` (`game/face_down.go`). For `FaceDownManifested` and `FaceDownCloaked` it returns:
  1. The **mana-cost route** when `c.PrintedIsCreature()`, exactly as today (CR 701.40b / 701.58b).
  2. The **keyword route** when the card declares a morph or disguise face-down cast. This is the same `FaceDownCastFor(faceUpCatalogKey(c))` lookup the `FaceDownTurned` arm already makes, without the kind-equality guard, for that arm's reason: CR 702.37e and 702.168d key on the card having the ability. It carries `FaceUpCounter` for megamorph.
- **Every route is offered, even at the same cost** (owner decision 2). Each turn-face-up offer carries a route: `SpecialAction.Route`, of type `game.TurnFaceUpRoute`, with the values `"mana_cost"`, `"morph"` (megamorph included) and `"disguise"`. Every arm of `TurnFaceUpOffers` stamps it: morphed, disguised and turned permanents get their keyword's route, and manifested and cloaked ones get `"mana_cost"` and, when the card has one, the keyword route too. A Den Protector manifested face down offers both {1}{G} routes, and only the megamorph one adds the +1/+1 counter.
- **Order**: lower mana value first. On a tie, the keyword route comes first.
- **Labels**: when a permanent has two offers, each label names its route, so a player can tell them apart: "Turn face up {1}{G} (megamorph, +1/+1 counter)", "Turn face up {2}{U} (morph)", "Turn face up {2}{U} (disguise)", "Turn face up {1}{G} (mana cost)". With one offer the label is unchanged, so no existing log line or test changes.
- **Choosing a route**: `SpecialActionParams.Route` (`game/special_action.go`). `specialActionOfferLocked` (`game/special_action_grant.go`) filters by it as it filters by `Door`. It is refused on any kind other than `turn_face_up`, an unknown value is refused by the action parser, and an empty route takes the first offer, so every existing caller and saved client keeps working. The performer `turnFaceUpLocked` already reads `sa.FaceUpCounter` off the chosen offer, so turning face up itself needs no change.
- `TurnFaceUpOffer(c)` stays and returns `TurnFaceUpOffers(c)[0]` (or nil), so `SpecialActionOffered` and every other single-offer reader is untouched. `SpecialActionsOfferedByCard` appends all of them.
- **Wire**: `SpecialActionView.route` (`route,omitempty`, `protocol/view.go`), and `special_action.route` on the action (`actions/actions.go`, beside `door`). Both are documented in `docs/protocol.md`.
- **Enumerator** (`legal/special_actions.go`): when a card offers `turn_face_up` more than once, each move's params carry its `route`. Its label is the offer's label, which names the route.
- **Heuristic** (`aiseat/heuristic/moves.go`): the turn-face-up arm keeps the flat `SpecialActionValue`, adds a small bonus for an offer with `FaceUpCounter` (a megamorph route), and so takes the counter when the two cost the same and the cheaper offer otherwise.
- **Client** (`contextMenu.logic.ts specialActionItems`, `protocol.ts`): the id becomes `special-<kind>-<route>` when a route is set, and the payload carries `route`. Both rows show at once, each with its own label and cost note, so the two choices are visibly different.
- ADR 0082 decision 5's "CR 701.40c is out of scope" is superseded by this section.

**Rejected:** collapsing equal-cost routes into the keyword route and telling offers apart by `cost` alone. That was the draft's recommendation. The owner chose to show both, and `cost` cannot tell two equal-cost offers apart, so a route parameter is needed. **Rejected:** leaving it out. The Kozilek and manifest dread rulings both state the double offer, and today the engine is weaker than printed.

### 5. Snapshot, restore, undo and the closure ratchet

- **No new persisted field.** `AdditionalCost` and `ExileCost` are catalog data. `CastSpellParams` is not in the snapshot. The cast cost reuses `PaidCost.Exiled`, which is already in the v8 shape. `ChooseManifest` is a new *value* of the persisted `ChooseDestination` string, an open vocabulary that "restores as itself" (`chained_choice.go`). `testdata/snapshot_shape/v8.txt` does not change, and no fixture is touched. No version bump.
- **Restore**: a game captured while a Kozilek or Scroll of Fate pick is open is in the same position as one captured during manifest dread's or Explore the Vastlands' prompt. The continuation rides `promptRun.then` and `chooseCardsFrame.then`, which are already `census:ChoiceResumeFrames`. `Turn.PhaseOrdinal` is saved with the turn, and a file from before #753 has it derived by `derivePrePlanOrdinalsLocked` (`turn_plan.go`), so survival reads the right ordinal after a restore.
- **Undo**: nothing new. Turning face up by either route is the existing special action. A hand manifest reveals nothing to another seat (CR 708.5): the cards were in the chooser's own hand.
- **Closure ratchet**: no new route from `Game` to a func. Every continuation uses an existing door. `closure_fields.txt` and its ceilings do not change. The slices assert this by running `TestClosureFieldsReachableFromGame` unchanged.

### 6. Wire, enumerator, heuristic bot, MCP seat and client

| | Survival (1) | Cast cost (2) | Hand manifest (3) | Both routes (4) |
|---|---|---|---|---|
| Wire | none | `AdditionalCostView.exile_cost_n/label/options/zone` (the ability view's names); `cast_spell.exile_ids` | `choose_cards` with `destination: "manifest"` | duplicated `special_actions` rows, each with `route` and a label naming it; `special_action.route` |
| Enumerator | none | one payment per cast: the policy's cheapest fuel, named in the move label (#3412) | existing `choose_cards` enumeration | one move per offer, with `route` |
| Heuristic | none | the existing fuel pricer for graveyard cards | `ChooseManifest` scored as cards given up | flat value, plus a small bonus for a megamorph route; otherwise the cheaper offer |
| MCP seat | none | move label names the exiled cards | prompt question names the count | two numbered moves with distinct labels |
| Client | none | reuse `exileCost.ts` and the graveyard picker in the cast flow; the cast gate (`timing.ts`) greys a spell with too few graveyard cards | verb "Manifest" on the picker | one menu row per route, id `special-<kind>-<route>`, payload `route` |

`docs/protocol.md` gets the new `cast_spell.exile_ids`, the `"manifest"` destination, and `route` on `special_actions` and `special_action`.

## Delivery plan

Slices 1, 2, 4 and 5 are independent. Slice 3 depends on slice 2. Each slice updates `server/internal/roadmap/registry.go` for its cards and regenerates `docs/engine-seams.md` (`scripts/go-docker.sh go test ./internal/roadmap/ -update`). **Row rule:** the slice that empties a row's `Waiting` and `Missing` flips its `Status` to `StatusImplemented` and moves its history to `docs/engine-seams/closed/`, following the sibling entries there. The face-down row closes with slices 4 and 5. The manifest dread row closes with slices 1 and 2. #1948's row closes its fixed-count form with slice 3; #1948 itself stays open as the tracker for slices 2 and 3. Every card ships with `Completeness` declared truthfully and its oracle fixture generated alone (`-update-oracle -oracle-ids=<id>`).

### Slice 1 — `feat(cards): survival and your second main phase, and defiant survivor` · tier:1-mechanical

- **Scope:** `server/internal/cards/effects/triggers_common.go` (`YourSecondMainPhaseBegan`, `AtYourSecondMainPhase`); new `server/internal/cards/effects/survival.go` (`Survival`); new `defiant_survivor.go` (`Survival("Defiant Survivor — manifest dread", Do(ManifestDread{}))`); `scheming_silvertongue.go` switches to `YourSecondMainPhaseBegan`. Registry: remove Defiant Survivor from `manifest-dread`'s `Waiting` and its survival clause from `Missing`/`EngineNotes`.
- **Acceptance (tests):**
  - A tapped Defiant Survivor at the start of its controller's postcombat main triggers, and manifest dread runs.
  - An untapped one does not trigger.
  - Tapped at the trigger but untapped in response: resolves and does nothing.
  - Tapped, then sacrificed in response: still manifests dread (last-known information).
  - On a turn with an added combat and main phase (Relentless Assault cast in the second main phase), the third main phase does not trigger it.
  - A turn whose combat was skipped still triggers in main phase 2.
  - An opponent's second main phase does not trigger it.
  - Scheming Silvertongue does not become prepared in a third main phase.
- **Depends on:** nothing.

### Slice 2 — `feat(server): exile cards from your graveyard as a spell's additional cost` · tier:2-standard

- **Scope:** follow commit 14217529f (#3412) file for file.
  - `game/additional_cost.go`: the `ExileCards` field.
  - New `game/exile_card_additional_cost.go`: `planExileCards`, `validateCastExileCardsLocked`, and the record onto `PaidCost.Exiled`.
  - `game/mutations.go`: `CastSpellParams.ExileIDs`, validated at announce and paid beside the sacrifices.
  - `game/spell_copy.go`: copy keeps `Exiled`, if not already.
  - `actions/actions.go`: `exile_ids` on `cast_spell`. `lobby/http.go`, if it mirrors the params.
  - `protocol/view.go`: `AdditionalCostView.ExileCostN/Label/Options/Zone`.
  - `legal/cast.go` and new `legal/exile_card_cast_cost.go`: one payment.
  - `aiseat/heuristic/params.go` and `moves.go`: carry and price `exile_ids`.
  - `effects/additional_cost.go`: `ExileFromGraveyardCost`. `effects/registry.go`: the Register guards.
  - Client: `protocol.ts`, `api.ts`, `castPreview.ts`, `targeting.ts`, `timing.ts`, `Board.svelte`, reusing `exileCost.ts`.
  - `docs/protocol.md` and `docs/adding-cards.md`.
  - New `effects/abhorrent_oculus.go`: the cost, flying, `AtEachOpponentsUpkeep(…, Do(ManifestDread{}))`.
  - Registry: Oculus off `manifest-dread`. `exile-from-graveyard-additional-cost` keeps Draconic Intervention and Relentless Skaabs until slice 3.
- **Acceptance (tests):**
  - Casting Oculus with six named graveyard cards exiles them with the spell on the stack. A watcher of cards leaving a graveyard triggers above the spell.
  - Countering the spell leaves the cards in exile.
  - Five cards in the graveyard: the cast is refused (CR 118.3), the view's `exile_cost_options` has five entries, and the cast gate greys it.
  - Naming a card twice, naming a card not in the caster's graveyard, or naming a card also named in `delve_ids` is refused, and nothing moves.
  - Cast without paying its mana cost still demands the six.
  - `ctx.Exiled()` returns the six at resolution, and a copy of the spell keeps them.
  - The enumerator offers exactly one cast move with six named cards. The bot casts Oculus from a seven-card graveyard.
  - `Register` panics on the hand form, on N ≤ 0, on an optional cost, and in an `EitherCost` branch.
  - Client unit test: the cast flow opens the graveyard picker and sends `exile_ids`.
  - The v8 shape file is unchanged.
- **Depends on:** nothing.

### Slice 3 — `feat(cards): draconic intervention and relentless skaabs` · tier:1-mechanical

- **Scope:** new `effects/draconic_intervention.go` (X is the mana value of the card in `ctx.Exiled()`; the exile-instead-of-dying rider; self-exile) and `effects/relentless_skaabs.go` (the cost and undying). Registry: close `exile-from-graveyard-additional-cost` (#1948) for its fixed-count form and keep the X forms in `Missing`. The PR body refers to #1948 (`Refs #1948`) but does not close it: it is the tracker for slices 2 and 3.
- **Acceptance (tests):**
  - Exiling a 3-mana-value instant deals 3 to each non-Dragon creature, and a creature dealt damage this way that would die is exiled instead.
  - Relentless Skaabs refuses a graveyard with no creature card.
  - Its undying returns it with a +1/+1 counter.
- **Depends on:** slice 2.

### Slice 4 — `feat(server): manifest cards from a hand, and kozilek, the broken reality` · tier:2-standard

- **Scope:**
  - `game/battlefield_put.go`: `ManifestFromHandsThenForEffect`, and rewrite `ManifestForEffect`'s stale paragraph.
  - `game/chained_choice.go`: `ChooseManifest`.
  - New `effects/manifest_from_hand.go`: `EachManifestsFromHand`.
  - `aiseat/heuristic/choices.go`: `ChooseManifest` uses `valueKeptInHand`.
  - Client: the "Manifest" verb wherever `destination` picks the picker's verb.
  - New `effects/kozilek_the_broken_reality.go` and `effects/scroll_of_fate.go`.
  - Registry: Kozilek off `face-down-objects`, with "manifesting cards from a hand" out of `Missing`.
- **Acceptance (tests):**
  - Kozilek targets two opponents with five and one cards in hand. The first is asked for exactly two. The second is not asked and manifests their one card. The caster draws three.
  - Both manifests are face down under their owners' control, and only their owner can see them.
  - Countering Kozilek still resolves the trigger.
  - Zero targets: nothing is manifested and nothing is drawn.
  - A manifested instant from a hand stays face down when turned up (CR 701.40g).
  - A replacement that keeps one card off the battlefield (a test-only replacement on hand-to-battlefield entries) lowers the draw by one.
  - "Other colorless creatures you control" get +3/+2: a face-down 2/2 you control is 5/4, and Kozilek itself stays 9/9.
  - Scroll of Fate taps to manifest one chosen card.
  - The bot, when it is the target, names its lowest-value hand cards.
  - `TestClosureFieldsReachableFromGame` passes with no edit to `closure_fields.txt`, and the v8 shape file is unchanged.
- **Depends on:** nothing.

### Slice 5 — `feat(server): turn a manifested or cloaked card up by its morph or disguise cost too` · tier:2-standard

- **Scope:**
  - `game/face_down.go`: `TurnFaceUpRoute` and its three values, `TurnFaceUpOffers` (every arm stamps `Route`; the order; the route-naming labels when there are two offers). `TurnFaceUpOffer` delegates to it, and the decision-5 comment is updated.
  - `game/special_action.go`: `SpecialAction.Route`, `SpecialActionParams.Route` (refused on other kinds), and `SpecialActionsOfferedByCard` appends all offers.
  - `game/special_action_grant.go`: `specialActionOfferLocked` filters by route.
  - `actions/actions.go`: parse `route` on `special_action`, refusing an unknown value.
  - `protocol/view.go`: `SpecialActionView.Route`.
  - `legal/special_actions.go`: `route` in the params when the card offers `turn_face_up` twice.
  - `aiseat/heuristic/moves.go` and `params.go`: decode `route`, and add the megamorph bonus.
  - Client: `protocol.ts` (`route` on the special-action view), `contextMenu.logic.ts` (id and payload).
  - `docs/protocol.md`.
  - Registry: the double-offer clause out of `face-down-objects` `Missing`.
- **Acceptance (tests):**
  - A manifested Den Protector (megamorph {1}{G}, mana cost {1}{G}) offers two routes, `morph` and `mana_cost`, both {1}{G}, with labels "Turn face up {1}{G} (megamorph, +1/+1 counter)" and "Turn face up {1}{G} (mana cost)". `route: "morph"` turns it up with the counter, and `route: "mana_cost"` turns it up without one.
  - A manifested morph creature whose costs differ offers two routes, the cheaper first. Each turns it up for its own price.
  - A cloaked disguise card offers `disguise` and `mana_cost`.
  - A manifested noncreature card with no morph offers nothing. A manifested noncreature card with morph offers only `morph`.
  - A morphed (cast face down) card offers one route, `morph`, and its label is unchanged.
  - `special_action` with no `route` takes the first offer. An unknown route is refused by the parser. `route` on a `foretell` is refused.
  - The enumerator emits one move per offer, each with its `route`, and the move labels differ.
  - The bot, given both Den Protector routes and enough mana, takes the megamorph route.
  - The view lists both offers to the controller only, each with its `route`.
  - Client unit test: two menu rows with distinct ids and labels, each sending its `route`.
- **Depends on:** nothing. It is more useful after slice 4.

## Owner decisions (2026-10-14)

1. **#1948 is folded in.** Slices 2 and 3 ship its fixed-count form, and #1948 stays open as their tracker.
2. **Equal-cost turn-face-up routes: offer both, told apart by a `route` parameter.** This overrides the draft's recommendation to collapse them. Decision 4, the wire table and slice 5 are written to it.
3. **A hand of N cards or fewer is not asked:** the player manifests all of them (decision 3).
4. **Scheming Silvertongue moves to the second-main-phase condition in slice 1.**
5. **Scroll of Fate ships in slice 4.**
6. **Graveyard only:** `Register` refuses the hand form of `ExileCards` on a spell (decision 2).

## Amendment (2026-10-16, #1948): the X forms and the either/or branch

Decision 2 left "the X forms (Skeletal Scrying, Harvest Pyre, Chill Haunting, Haunting Misery), and Soaring Stoneglider's either/or branch" on #1948. Both ship now, as two small extensions of the same component, so the walk, the validator, the payer and the record stay the ones decision 2 chose.

**1. "Exile X cards from your graveyard" is the X the cast announces.** `game.ExileCost` gains `CountFromX`, built by `effects.ExileXFromGraveyardCost(label, match)` with `N` zero. CR 601.2b announces X with the spell, and CR 107.3i makes every X on the object one number, so the count is `CastSpellParams.XValue`: Skeletal Scrying's {X} in its mana cost and its exile are the same X, and Harvest Pyre, whose mana cost has no {X}, announces one anyway, the way Toxic Deluge's "pay X life" and Soul Immolation's "blight X" do. The caster names exactly X cards on `exile_ids`; X = 0 names none and is legal. The largest X is what the graveyard can pay (CR 118.3): `Game.ExileCardsXCeilingForEffect`, the number of cards the clause's own walk offers, read by the validator (an X above it is refused before anything moves), the view (`additional_cost.exile_cost_x`, `exile_cost_x_max`, and `demands_x`) and the enumerator, so the three agree (#544). The spell reads X through `ctx.X()`, which is also the number exiled. A free cast of Skeletal Scrying is X = 0 by CR 107.3b, as for any {X} spell; a free Harvest Pyre still announces X, as Toxic Deluge does.

- **Client.** The X prompt opens for the card (it demands X) and is capped at `exile_cost_x_max`; the graveyard picker opens after it, sized to X, with the "X" in the clause replaced by the number. X = 0, or a graveyard holding exactly X matching cards, needs no picker.
- **Enumerator.** The X search is capped by the ceiling, the move names the first X cards of the policy's cheapest-fuel-first order, and the move never carries an open X range, because its `exile_ids` fix the X.
- **Guards.** `effects.Register` refuses the X form on an activated or mana ability, on an optional cost, on an either/or branch, beside pay X life, blight X or a waterbend {X} (one announced X can't pay both), and with a printed count beside it.

**2. A fixed count may be a branch of an either/or cost.** Soaring Stoneglider's "exile two cards from your graveyard or pay {1}{W}" is `EitherCost(ExileFromGraveyardCost(2, …).Keyed("exile"), ManaAdditionalCost("{1}{W}").Keyed("mana"))`. `ChosenAdditionalCost` already hands the chosen branch to the plan as its mandatory entry (ADR 0100 §2), so the validator and payer needed nothing; `AdditionalCostBranchPayableLocked` now counts the graveyard for an exile branch, which is what the view's `branches[].payable`, the enumerator and `CastSpell` ask. The X form stays off a branch: no printed card puts one there.

**Rejected:** a separate "exile X" component. It would be a second walk and a second validator for one clause, the thing decision 2 avoided. **Rejected:** letting the enumerator offer an open X range for these casts, as it does for a plain {X} spell. The exiled cards are named per X, so a range would hand a seat an X its `exile_ids` don't match.
