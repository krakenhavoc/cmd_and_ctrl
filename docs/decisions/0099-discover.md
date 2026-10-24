# ADR 0099 — Discover

**Status:** Accepted · 2026-09-30 · Post-S30 — Rolling deck-driven catalog growth. The owner accepted every recommendation; see [Owner decisions](#owner-decisions-2026-09-30).
**Issue:** [#1112](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1112), the Pirates deck tracker. The `discover` row in `server/internal/roadmap/registry.go` names it, with Hit the Mother Lode and Brass's Tunnel-Grinder waiting. #1751 shipped Tecutlan, the Searing Rift (Brass's back face) with its discover trigger left out and a caveat.
**Numbering:** I ran the AGENTS.md §4 sweep on 2026-09-30: `git fetch --prune`, then every `docs/decisions/` file name on every remote branch (37 heads). The highest number anywhere is **0098**, and 0099 is free. 0098, 0100, 0101 and 0102 are reserved for ADRs being written at the same time.
**Builds on:**
- The S28 cascade implementation (`game/cascade.go`, `cards/effects/cascade.go`). It has no ADR of its own.
- [ADR 0066](0066-granted-cast-and-play-permissions.md), granted cast permissions, and its miracle amendment (#1665), where passing priority closes the window.
- [ADR 0054](0054-dice-rolls-and-coin-flips.md), the keyed RNG behind "a random order".
- [ADR 0034](0034-multi-face-cards.md), faces.
- [ADR 0041](0041-game-persistence.md), restore points.
- [ADR 0091](0091-hideaway.md), which states the "grant, not an inline cast" posture most recently.

---

## Context

Discover is an Ixalan keyword action (LCI, 2023). It has spread to Commander precons and Universes Beyond sets since. The engine has no primitive for it. The registry row says so: "no primitive anywhere in `internal/game` or `internal/cards/effects` (re-checked 2026-09-30)".

It is cascade's closest relative. Both exile from the top of the library until a qualifying nonland card, offer a free cast, and bottom the rest at random. The engine already has every piece cascade needs:

- the exile-until walk, face up and marked known;
- the `may_cast` prompt;
- a `{0}` cast permission on the exiled card;
- the keyed random bottom (`PutOnBottomInRandomOrderForEffect`);
- an end-step delayed trigger that tidies up an uncast hit.

This ADR decides how much of that discover reuses, and what it adds.

### The rules

Checked against the pinned CR text (effective August 7, 2026).

- **CR 701.57a.** "Discover N" means: "Exile cards from the top of your library until you exile a nonland card with mana value N or less. You may cast that card without paying its mana cost if the resulting spell's mana value is less than or equal to N. If you don't cast it, put that card into your hand. Put the remaining exiled cards on the bottom of your library in a random order."
- **CR 701.57b.** A player has "discovered" once the 701.57a process is complete, "even if some or all of those actions were impossible". An empty library still discovers.
- **CR 701.57c.** If the last card exiled has mana value N or less, it is "the discovered card", whether it was cast or put into a hand.
- **CR 702.85a** (cascade, for comparison). A **triggered ability** of a spell on the stack: "…a nonland card whose mana value is **less than** this spell's mana value. You may cast that card without paying its mana cost if the resulting spell's mana value is less than this spell's mana value. Then put all cards exiled this way that weren't cast on the bottom…"
- **CR 608.2g.** A spell that an effect lets a player cast during resolution is cast then, following 601.2a–i, and "no player receives priority after it's cast". The cast ignores the card's own timing.
- **CR 107.3b.** A spell with {X}, cast without paying its mana cost, has X = 0.
- **CR 118.9.** "Cast it without paying its mana cost" is an alternative cost.
- **CR 202.3d, 712.8a, 712.8f, 715.4.** Outside the stack, a split card's mana value is its halves combined. A double-faced card has only its front face's characteristics, and an adventurer card only its normal ones. On the stack, a modal DFC is the face that was cast, and a split card is the half that was cast. That is why 701.57a checks "the resulting spell's mana value" separately from the hit.
- **CR 117.3b.** The active player receives priority after a spell or ability resolves.
- **CR 903.9b.** A commander that would be put into its owner's hand may go to the command zone instead.

### How discover differs from cascade

| | Cascade (CR 702.85) | Discover (CR 701.57) |
|---|---|---|
| What it is | A triggered ability of a spell, "when you cast this spell" | A keyword **action**: an instruction in any effect (a spell's resolution, a trigger, an activated ability, a loyalty ability) |
| The number | This spell's mana value, captured at cast | N, given by the instruction. Often X, read from the board or the triggering event |
| What is a hit | Nonland, mana value **less than** the number | Nonland, mana value **N or less** |
| The resulting spell | Mana value must be less than the number | Mana value must be **N or less** |
| If you don't cast it | Goes to the bottom with the rest | **Goes into your hand** |
| Its own event | None. Nothing says "whenever you cascade" | **Yes.** Curator of Sun's Creation and Val trigger on "whenever you discover" |
| Result a card reads | None | "The discovered card" (CR 701.57c). Hit the Mother Lode reads its mana value |
| Who does it | The caster | Usually "you". Zoyowa's Justice makes the **target's owner** discover |

The hand fallback changes the prompt. A declined cascade loses the card, so "yes" and "no" carry a real trade-off. A declined discover keeps it. Under the engine's grant posture that makes accepting weakly dominant **unless the grant's window is short**. Decision 3 is about exactly that.

### The cards

The Scryfall dump has **35** oracle cards that discover or refer to discovering. Placeholder printings are filtered out.

**31 are Commander-legal and are the cards covered:**

- **"Discover N" from a resolving spell:** Daring Discovery, Hit the Mother Lode, Walk with the Ancestors, Hurl into History, Contest of Claws, Zoyowa's Justice.
- **From a triggered ability:** Aloy, Savior of Meridian; Caparocti Sunborn; Chimil, the Inner Sun; Digsite Conservator; Dinosaur Egg; Etali's Favor; Franklin Richards, Ascendant; Geological Appraiser; Monstrous Vortex; Pantlaza, Sun-Favored; Primordial Gnawer; Trumpeting Carnosaur; Zoetic Glyph; Tecutlan, the Searing Rift (the back face of Brass's Tunnel-Grinder).
- **From an activated or loyalty ability:** Buried Treasure; Ellie and Alan, Paleontologists; Hidden Cataract; Hidden Courtyard; Hidden Necropolis; Hidden Nursery; Hidden Volcano; Long-Range Sensor; Quintorius Kand (−3); Swashbuckler's Whip (a granted ability, [ADR 0093](0093-abilities-granted-to-other-permanents.md)).
- **"Whenever you discover":** Curator of Sun's Creation.

**4 are out of scope.** Three are Alchemy-only: Ashaya's Enduring Bond, Scalesoul Gnome (conjure) and Val, Marooned Surveyor (seek). The fourth is a Planechase plane, Oteclán.

Several covered cards need something besides discover. Each implementing PR re-checks those needs in code, because the seams doc runs stale.

- **Chimil**'s "Spells you control can't be countered" is a static grant over the stack. `cant_be_countered.go` says it is not modelled. Chimil ships with a caveat on that line or waits. That is owner question 5.
- **Tecutlan** needs the #1547 spend rider (`WhenManaSpent`), which exists.
- **Contest of Claws** needs excess damage (as in Hell to Pay).
- **Dinosaur Egg** needs last-known toughness (`ctx.TriggeringPermanent()`).
- **Swashbuckler's Whip** needs a granted bundle.
- **Pantlaza** and **Curator** need the once-per-turn gates.

As far as I can tell, all of those shapes already exist.

---

## Options

### A. Reuse cascade's machinery, with a short window (recommended)

Discover is a new engine primitive built from cascade's parts:

- the same exile-until walk, factored out and shared;
- the same `may_cast` prompt, with discover's own labels;
- the same `{0}` cast permission on the hit, plus two things cascade lacks. The cast obeys CR 608.2g's timing (`TimingFlash`, as suspend, madness, miracle and hideaway already do), and the resulting spell's mana value is capped.

The window closes on the discoverer's **next priority pass**, which is miracle's rule (#1665). The card then goes to hand.

A "discover" event fires when the card's fate is settled. The effect gets a continuation that is handed the discovered card.

### B. A dedicated `discover` prompt kind

This is Option A with a new `PendingChoiceKind` in place of `may_cast`. The client copy would be cleaner. The cost is one more kind to classify in the choice gate, the enumerator and the departure table, plus a new resume-frame route in the closure ratchet. The answer is still a yes or no, so a new kind buys only copy, and a label on the existing kind gives the same copy.

### C. An inline cast during resolution (CR 608.2g as printed)

The prompt's answer would *be* the cast. The client would run the whole announcement (face, targets, modes, X locked at 0) while the resolution waits, and the spell would go on the stack above the rest of the resolving effect with no priority in between. This is the only faithful option. It would also fix cascade, suspend, madness, miracle and hideaway at once.

But it is the frame ADR 0066, the cascade header and ADR 0091 all declined to build: "the announce path has no frame for a half-validated cast". It is a cross-cutting engine change of its own and deserves its own ADR. It should not ride on one keyword action.

### D. Put-into-hand only

Discover without the free cast. It is weaker than printed, so every card would carry a caveat. It is not worth shipping when A is a few hundred lines.

**Recommendation: A.** The one real departure from the printed card is the one ADR 0091 already declared for every "you may cast it" a resolution offers: the cast comes after the resolution, and players may respond in between. What A adds keeps the discover cards from being stronger than printed. The window closes on the next pass, so a player cannot hold a free spell across the turn and still keep the card.

---

## Decision (proposed)

### 1. One engine primitive, callable from any effect

```go
// game: caller holds g.mu (it runs inside a resolution)
func (g *Game) DiscoverThenForEffect(player, source uuid.UUID, n int,
    then func(g *Game, r DiscoverResult) error) error

type DiscoverResult struct {
    N          int        // the number the instruction named
    Discovered uuid.UUID  // CR 701.57c; uuid.Nil when the walk found nothing
    ManaValue  int        // the discovered card's mana value, as it was in exile
}
```

Card side, in `cards/effects/discover.go`:

```go
Discover{N: 4}.Apply(ctx)                                   // Daring Discovery; Player defaults to the controller
Discover{Player: owner, N: mv}.Apply(ctx)                   // Zoyowa's Justice: "that player discovers X"
Discover{N: 10, Then: func(ctx *Context, r game.DiscoverResult) error { … }}.Apply(ctx) // Hit the Mother Lode
WheneverYouDiscover(label, effect)                          // Curator of Sun's Creation
```

It is an instruction, not a trigger, so no `FromStack` is needed. The caller works out N.

- **Board-read values** are read at resolution (CR 608.2h). Aloy's greatest power and Franklin's fixed 6 are examples.
- **Trigger-event values** ride the item, captured by the fill-in `Build` exactly as cascade captures its mana value. Examples are Monstrous Vortex's and Tecutlan's "that spell's mana value", Dinosaur Egg's toughness through `ctx.TriggeringPermanent()`, and Hurl into History's countered spell, read before it is countered.

`Then` runs once, after the prompt is answered. It also runs, with `Discovered == uuid.Nil`, when nothing was found or the discoverer has left the game. That is CR 701.57b's "even if … impossible" and the #865 rule that a `Then` always runs.

### 2. The walk is cascade's, shared

`CascadeForEffect`'s loop becomes `exileUntilLocked(player, hit func(Card) bool) (pile []uuid.UUID, hit uuid.UUID, err error)`. It exiles face up, marks each card known, and emits the zone move. Cascade passes `mv < lessThan` and discover passes `mv <= n`. Both keep cascade's rule that a card whose cost `ParseCost` cannot read is **not a hit**. A split card's joined cost is the example. It keeps the walk weaker than printed rather than stronger.

If the walk finds nothing, the pile goes to the bottom at random, the continuation runs, and the event fires (Decision 5).

### 3. The prompt and the grant

On a hit, the discoverer is asked through the existing `may_cast` prompt:

- the reason reads "Discover 4 — cast Lightning Helix without paying its mana cost?";
- `AcceptLabel` is "Cast it free" and `DeclineLabel` is "Put it into your hand";
- a new `MayCastKeyword` field is set to `"discover"`, so the client can name the rule. See §Client.

Either answer bottoms the rest of the pile at random first, as cascade does.

**Decline** routes the card from exile to its owner's hand through the one exit primitive. So CR 903.9b offers a discovered commander the command zone.

**Accept** stamps a `ScopeCards` permission on the exiled card object:

| Field | Value | Why |
|---|---|---|
| `Cost` | `"{0}"` | "Without paying its mana cost". It locks X at 0 (CR 107.3b) through the existing `CastCost.LocksXAtZero` |
| `Timing` | `TimingFlash` | CR 608.2g. A discovered sorcery can be cast in combat or on an opponent's turn, as printed |
| `CastOnly` | true | "Cast", and a land is never a hit |
| `Duration` | until end of turn | A backstop only (see below) |
| `MaxSpellManaValue` (**new**) | `n` | CR 701.57a's "if the resulting spell's mana value is N or less" |
| `Discover` (**new**) | `{N, Source}` | Marks the grant as a discover grant, for Decision 4 and Decision 5 |

`MaxSpellManaValue` is a `*int`, because discover 0 is a real instruction (Dinosaur Egg with 0 toughness). Nil means no cap. `CastSpell` checks it once the face and X are fixed, and `CastOffersForLocked` checks it too. So the view, the bot enumerator and the engine all refuse the same casts: a modal DFC's expensive back face (Valki into Tibalt, cast from the back) and an adventure half that costs more than N. Cascade takes the same field with `lessThan-1`, which is owner question 2.

### 4. The window closes on the discoverer's next pass

Miracle has already built this rule (#1665): `PassPriority` calls `closeMiracleWindowLocked(holder)` **before** the pass runs. A discover grant closes in the same place. When the holder passes priority, each discover grant they hold whose card is still in exile is dropped, and the card is routed to its owner's hand. It is the decline branch, run late.

After the resolution, the active player gets priority first (CR 117.3b). So on another player's turn (Hurl into History, Zoyowa's Justice) the table may act before the discoverer does. That is ADR 0091's declared response window, and it is unchanged.

The discoverer must pass before the step can end, so the end-of-turn duration never lapses in play. If the sweep ever does find a live discover grant, it moves the card to hand the same way, so no card is ever stranded in exile.

A consequence: cascade's end-step delayed trigger (`cascade/bottom-if-not-cast`) has **no** discover counterpart. No new body key is needed.

Why not cascade's end-of-turn window: under it, "Cast it free" strictly beats "Put it into your hand". The player keeps a free instant-speed spell for the rest of the turn, and the card still reaches their hand at the end step. That is stronger than printed in the one way a discover card must not be.

### 5. "Whenever you discover" is its own event, fired when the card is settled

There is a new `EventDiscover` with these fields:

- `Actor` is the discoverer;
- `Source` is the discovering card;
- `Amount` is N;
- `CardID` is the discovered card, or `uuid.Nil`.

It is **not** a flag on `EventKeywordAction`. It is a distinct keyword, with payoffs that must not fire on a scry, which is the reasoning `EventSurveil` records.

It fires when the process is complete (CR 701.57b). Under the grant model that is when the card's fate is settled:

- **Decline:** after the card reaches hand.
- **Accept and cast:** in `CastSpell`, right after the `EventCast` of the spell that used a discover grant.
- **Accept and pass:** when the Decision 4 lapse moves the card to hand.
- **No hit, or discoverer gone:** at once.

Firing it on the cast keeps the paper ordering. The discovered spell is already on the stack when Curator's "discover again" trigger is put above it, so the trigger resolves first. Firing it at the answer would reverse that. The grant carries `Discover{N, Source}`, which is how the late firing knows N. Owner question 3.

Curator reads `ev.Amount` for "the same value" and gates itself with `TriggeredThisTurn`.

`projectEvent` gets an arm: "Alice discovers 4 — Lightning Helix", or "… — nothing". The hit was exiled face up, so the name is public. This satisfies `TestEveryEventKindIsNarratedOrDeliberatelySilent`.

### 6. The bot

The enumerator changes nothing. `may_cast` already offers accept and an always-legal decline (`legal/choices.go`). A live grant already produces cast-from-exile moves through `CastOffersForLocked`, now with the cap. Pass is always there.

The heuristic scores `may_cast` today only as "unrecognised choice", a tie. Under Decision 4, accepting is never worse than declining. So the policy says **accept**, and the ordinary priority policy decides whether to cast the card. If it passes instead, the card goes to hand, which is the correct outcome. This also gives cascade a sane default.

`aiseat/` still imports nothing from `internal/game`, because it reads only the view.

### 7. The client

- **`ChoicePromptModal`'s `may_cast` branch.** The copy today is cascade-only: the "cascade · CR 702.85" tag, "To the bottom", and the hint about the bottom of the library. That copy is already wrong for suspend and madness. The branch reads `accept_label` and `decline_label`, and the view projects them for `may_cast` as it does for `confirm`. It also reads a new `may_cast_keyword` for the tag and the hint. For discover the tag is "discover · CR 701.57", and the hint says that passing without casting puts the card into your hand.
- **After "Cast it free"**, the client hands the exiled card to `handlePlayCard` with `from_zone: "exile"`, the one cast entry point (#874). The face picker, targets and modes open straight away. If the player cancels, the card waits in exile with the ordinary impulse button, and a banner says "Cast X free, or pass to put it into your hand". Nothing new is dispatched.

### 8. Snapshots

- **While the prompt is open**, the table is not a restore point. `mayCastResume` is already counted in `ContinuationCensus.ChoiceResumeFrames`, exactly as with an open cascade prompt. No new route is added to `closure_fields.txt`, because the frame is the existing one, and no ceiling moves. The `Then` continuation lives in that frame.
- **After the answer**, everything is data. The grant is a `CastPermission` whose two new fields are additive under the current schema version, and gets recorded with `-update-shape`. The event kind is a new string. The rest of the pile is already on the bottom. A table whose discoverer has not yet passed is a restore point.
- **Randomness** is the `random_order` stream (ADR 0054), so an undo or a restore replays the same bottom order.

---

## Consequences

- One new file each side: `game/discover.go` and `cards/effects/discover.go`. There is also a small refactor of `game/cascade.go` (the shared walk), two additive `CastPermission` fields, a hook beside `closeMiracleWindowLocked`, one event kind with a log arm, and the view and client label changes.
- The `discover` registry row flips to implemented in the engine PR, with a closed-seam fragment. Tecutlan's caveat goes stale that day and is cleared. A `cards/coverage` probe for "discover" makes a stale caveat fail the build.
- The 31 cards come in follow-up card batches. Chimil keeps a caveat on its first line unless the owner holds it.
- **Declared departure:** there is a priority window between the resolution and the cast. Other players can act in it on another player's turn. This is ADR 0091's statement, unchanged.
- **Declared departure:** a split card is never discovered, because its joined cost is unreadable. That is weaker than printed, and cascade shares it.

## Out of scope

- The inline CR 608.2g cast (Option C), for discover, cascade and the rest. It needs its own ADR.
- Any "if you would discover" replacement. No printed card has one, so discover does not open a `RepEventKeywordAction` window.
- The four non-Commander cards.

---

## Open questions for the owner

1. **The window.** Should a discover grant close on the discoverer's next priority pass, with the card going to hand (miracle's rule, recommended)? Or at end of turn with an end-step fallback to hand (cascade's, madness's and hideaway's rule, which lets "Cast it free" strictly beat "Put it into your hand")?
2. **Cascade.** Should the same PR move cascade onto the shared pieces too: the `MaxSpellManaValue` cap, `TimingFlash` (a cascaded sorcery castable on an opponent's turn) and the pass-closed window? That changes how a shipped keyword plays. The alternative is to leave cascade exactly as it is and share only the walk.
3. **When "whenever you discover" fires.** When the card is settled (on the cast, so Curator's trigger resolves before the discovered spell, as in paper; recommended)? Or at the prompt's answer, which is simpler but reverses that order?
4. **Prompt shape.** Should discover reuse `may_cast` with labels and a keyword field (recommended), or get a dedicated `discover` prompt kind (Option B)?
5. **Partial cards.** Should Chimil, the Inner Sun ship with a caveat on "Spells you control can't be countered", or wait for that static grant?
6. **Hit the Mother Lode with no hit.** I read CR 701.57c as saying that an empty walk has no discovered card, so no Treasures are made. Do you agree?

---

## Owner decisions (2026-09-30)

The owner accepted every recommendation.

1. **The window.** A discover grant closes on the discoverer's next priority pass, and the card goes to hand. This is miracle's rule (Decision 4).
2. **Cascade: yes.** The same engine work moves cascade onto the shared pieces, not just the walk:
   - the `MaxSpellManaValue` cap, set to the cascading spell's mana value minus one;
   - `TimingFlash`, so a cascaded sorcery can be cast on an opponent's turn (CR 608.2g);
   - the pass-closed window. An uncast cascade hit goes to the bottom of its owner's library when the caster next passes priority, and `cascade/bottom-if-not-cast` stops being scheduled.

   The engine PR lists every cascade behaviour this changes.
3. **"Whenever you discover"** fires when the card is settled. On a cast, that is right after the `EventCast`, so Curator of Sun's Creation's trigger resolves before the discovered spell (Decision 5).
4. **Prompt shape.** Discover reuses `may_cast`, with the accept and decline labels and a keyword field (Decision 3 and §Client).
5. **Chimil, the Inner Sun** ships with a player-facing caveat on "Spells you control can't be countered".
6. **Hit the Mother Lode** makes no Treasures on an empty walk. With no hit there is no discovered card (CR 701.57c).

---

## Amendment 2026-10-13 — "Exile until ... you may cast it" with the caller's test (#3213)

**Why it lands here.** Tibalt's Trickery (Kuja deck, slice kuja-b) is cascade's walk with a different hit test: "exile cards from the top of their library until they exile a nonland card with a different name than that spell. They may cast that card without paying its mana cost. Then they put the exiled cards on the bottom of their library in a random order." It changes no earlier decision.

### Decision

1. **One entry point.** `Game.ExileUntilCastFreeForEffect(player, source, hit func(Card) bool)` in `game/exile_until_cast.go`. It is `CascadeForEffect` on the pieces this ADR already shares: `exileUntilLocked` for the walk, the `may_cast` prompt, a `{0}` flash-timed grant that lapses to the bottom of the library on the chooser's next pass (`LapseToLibraryBottom`), and `PutOnBottomInRandomOrderForEffect`.
2. **No mana value cap.** `MaxSpellManaValue` is nil. The hit test is the caller's, so the card can carry any condition (here: nonland and a different name). A walk that finds nothing puts the pile back and asks nothing.
3. **The chooser is the walked player.** The caller names whose library is walked and who chooses. For Tibalt's Trickery that is the countered spell's controller, not the caster.
4. **`hit` is not stored.** It runs during the walk only. It may close over scalars read before the call (a name), never a `*Card`, so the continuation that follows stays restoreable.
5. **Prompt keyword.** The prompt asks under `MayCastKeywordExileUntil` (`"exile_until"`), which the client words like cascade and opens a free cast for (`mayCast.ts`, `freeCastRequest.ts`).

### Wire, client and bot

No new field and no new prompt kind: `may_cast` with a new `may_cast_keyword` value. The client adds one copy case and puts the keyword in `mayCastKeywordsThatOpenACast`. The enumerator offers accept and decline for every `may_cast`, whatever its keyword.

### Cards

- **Tibalt's Trickery: Full.** The counter, the random 1, 2 or 3 (the keyed stream, so undo redraws the same number), the mill through its CR 614 window, then this walk.

### Tests

`game/exile_until_cast_test.go` (the walk stops on the caller's hit; accept grants an uncapped free flash cast and bottoms the pile; decline and no-hit bottom everything) and `cards/effects/tibalts_trickery_test.go` (the victim, not the caster, mills, walks and chooses; a land and a same-name card are passed over; decline; no target spell refused).

### Back-out

Make the walk ignore the caller's test (`hit` replaced by a constant false): `TestExileUntilCastStopsOnTheCallersHit` and the Tibalt's Trickery tests fail.

### Snapshot impact

None. The grant, the prompt and the keyword string are existing persisted shapes.

## Amendment 2026-10-13: the same window over a card in hand (#3270)

An owner report from live play: "Buster Sword: it gives me the option but does not cast it from my hand." Buster Sword, Kari Zev's Expertise, Rishkar's Expertise, Surtland Elementalist and Great Intelligence's Plan all say "you may cast a spell from your hand without paying its mana cost". Each queued a `choose_cards` pick and then stamped a one-shot `{0}` permission over the card (§4's grant model, without §4's window), so picking did nothing visible: the player had to find the card and cast it themselves, at the card's own timing, any time that turn. CR 608.2g puts the cast inside the resolution, ignoring timing, and if the player does not cast it then, the chance is gone.

**Decision.**

1. **The window is §4's, for a hand.** The pick's grant is `Game.FreeCastFromHandPermission`: `Cost: "{0}"`, claimed under `AltCostKeyGrantedCost` (#3234), `TimingFlash`, `CastOnly` and `LapseOnPass: LapseStaysInHand`. The holder's next priority pass ends it and the card stays in hand, the way `LapseStaysInExile` leaves a rebound card in exile. Nothing moves when it fires. An uncastable pick (no legal target) therefore costs nothing, and a free spell cannot be held across the turn.
2. **One helper for the five cards.** `effects.offerFreeHandCast` (`effects/free_hand_cast.go`) filters the hand by a `CardPredicate`, queues the pick and grants the permission on the answer. The prompt carries `choose_destination: "cast_free"` (`game.ChooseCastFree`), the same vocabulary #2680 gave the put-from-hand prompts.
3. **The client opens the cast.** Answering a `cast_free` pick with a card sets `handFreeCastRequest`; once the snapshot shows the `granted-cost` offer on the hand card, Board starts the cast chain with that offer already chosen (a card with a question of its own before the cost, such as a face or kicker, goes through the ordinary chain with the offer preselected in `AlternativeCostModal`). The same `freeCastRequest` shape as §7's, for a hand card instead of an exiled one.
4. **The bot.** `heuristic` prices a `cast_free` pick as a cast (`puts.go`: what the card resolves into, over what keeping it is worth) rather than as a discard, and the enumerator already offers the claimed cast, so a bot picks a card and casts it before it passes.
5. **The Great Synthesis** (a grant over every card in hand, no pick) moves from `LapseStaysInExile`, a no-op for a hand card, to `LapseStaysInHand`. Behaviour is unchanged.

The SANDBOX SIMPLIFICATION of §Decision 4 stands: the cast is an ordinary `cast_spell` taken right after the pick, not a frame inside the resolution. The rest of the resolution has already run when the pick is asked.

**Back-out.** Removing the `Timing`/`LapseOnPass` lines from the grant, the `ChooseCastFree` destination, or the client's `handFreeCastRequest.set` each fails a named test; the pull request lists them.

## Amendment 2026-10-13: a free cast of an exiled card, or into your hand (#3245)

Beseech the Mirror, bargained: "you may cast the exiled card without paying its mana cost if that spell's mana value is 4 or less. Put the exiled card into your hand if it wasn't cast this way." That is discover's pair of branches over a card a search has already exiled, with the card's own cap and no discover event.

**Decision.**

1. **One helper.** `Game.CastExiledFreeOrToHandForEffect(player, source, card, maxMV)` (`game/cast_free_or_hand.go`). A land, or a card over the cap, goes straight to its owner's hand with no question. Otherwise it asks a `may_cast` prompt; accept grants §4's window (`{0}`, `TimingFlash`, `CastOnly`, `MaxSpellManaValue`, `LapseOnPass: LapseToHand`) with no `Discover` record, and decline puts the card into its owner's hand. The existing non-discover `LapseToHand` arm returns it on the holder's next pass.
2. **Prompt keyword.** `MayCastKeywordFreeOrHand` (`"free_or_hand"`). The client words it like discover and lists it in `mayCastKeywordsThatOpenACast`, so accepting opens the cast chain.

**Back-out.** Dropping `LapseOnPass` from the grant fails `TestBeseechTheMirrorAcceptedGrantLapsesToHand`; dropping the cap test sends a five-mana card to the prompt and fails `TestBeseechTheMirrorBargainedOverFourGoesToHand`.

**Snapshot impact.** None. The grant, the prompt and the keyword string are existing persisted shapes.

## Amendment 2026-10-14: a PAID cast during the resolution, with an "if you don't" (#3510)

Chandra, Torch of Defiance's first +1: "Exile the top card of your library. You may cast that card. If you don't, Chandra deals 2 damage to each opponent." CR 608.2g puts the cast inside the resolution, so its timing is ignored, but unlike every cast in this ADR so far it is PAID at the card's own mana cost (and mana abilities may be activated during it, CR 601.2g). "If you don't" is judged by whether that cast was made. The registry row read a `CastPermission` as the wrong shape because one is offered at normal timing and outlasts the resolution; §4's pass-closed window is neither, so it is the right shape, and what was missing was a consequence for not casting.

**Decision.**

1. **`CastPermission.IfNotCast`**, a `CastFollowUpKey` beside `FollowUp`: the body that runs when a pass-closed (`LapseOnPass`) permission ends without its cast, from `lapseCastPermissionLocked`, so the holder's pass and the end-of-turn backstop sweep both run it, once, after the card has gone where its lapse sends it. A cast spends the grant (`consumePassClosedGrantLocked`) and runs nothing. A key, not a func, for `FollowUp`'s reason: the permission is persisted.
2. **One helper.** `Game.ExileTopThenPaidCastOrElseForEffect(PaidCastOrElse)` (`game/cast_paid_or_else.go`) exiles the top card (a commander's CR 903.9 question pauses it) and calls `OfferPaidCastFromExileForEffect`, which asks a `may_cast` under the new keyword `MayCastKeywordPaidOrElse` (`"paid_or_else"`). Accept grants `PaidCastOrElsePermission`: no `Cost` (the printed cost), `TimingFlash`, `CastOnly`, `LapseOnPass: LapseStaysInExile`, `IfNotCast`. Decline runs the body at once. A land, an empty library, or a commander sent to the command zone instead asks nothing and runs the body: none of them can be "that card" cast.
3. **Client.** `mayCast.ts` words `paid_or_else` as a paid cast, takes the server's decline label ("Don't cast it (2 damage to each opponent)"), and `mayCastKeywordsThatOpenACast` lists it, so accepting starts the ordinary cast chain from exile, payment included.
4. **Bot.** The enumerator already offers both answers to a `may_cast`, the decline always legal. The heuristic's yes stands: accepting is never worse than declining, because a pass without the cast runs the same consequence. The enumerator then offers the paid cast from exile like any other.

The SANDBOX SIMPLIFICATION of §Decision 4 stands, and it is a little wider for a paid cast: the holder may take other actions (play a land, cast another spell) before the exiled card or instead of it. The player never gets both the cast and no damage without casting, so nothing is stronger than printed in what it produces; only the sequencing is looser.

**Cecily, Haunted Mage (#2110).** The #3270 pick already is her shape. She ships on `offerFreeHandCastFrom`, the Context-free form of `offerFreeHandCast`, called from a `DrawThen` body so the life loss and the eleven-card count wait for the draw.

**Back-out.** Removing the `IfNotCast` block from `lapseCastPermissionLocked` fails `TestPaidCastOrElsePassingWithoutCastingRunsIfNotCastOnce`, `TestPaidCastOrElseSweepBackstopRunsIfNotCast` and `TestChandraTorchAcceptThenPassDealsTheDamage`; dropping `TimingFlash` or `LapseOnPass` from the grant fails `TestPaidCastOrElseAcceptGrantsAPaidPassClosedCast`; setting `Cost: "{0}"` fails `TestChandraTorchAcceptPaysTheCostAndDealsNoDamage`.

**Snapshot impact.** One new omitempty field on a persisted shape (`ifNotCast`). An old snapshot reads it empty, which is every permission written before this.

## Amendment 2026-10-15: a free play or cast from the top of the library, else into your hand (#3801)

Ziatora's Envoy: "Whenever this creature deals combat damage to a player, look at the top card of your library. You may play a land from the top of your library or cast a spell with mana value less than or equal to the damage dealt from the top of your library without paying its mana cost. If you don't, put that card into your hand." Rashmi, Eternities Crafter prints the revealed, spell-only form. The pieces existed apart: §4's pass-closed free cast (from exile, or from hand since #3270), `PlayLandDuringResolutionForEffect` (hideaway's land, any zone). Nothing offered either one over a card still in the **library**, in one window, with "if you don't, into your hand" as the lapse.

**Decision.**

1. **One helper.** `Game.OfferFreePlayFromTopOrHandForEffect(FreePlayFromTop)` (`game/free_play_from_top.go`) looks at (or, with `Reveal`, reveals) the top card of `Player`'s library and runs the offer. `MaxManaValue` caps the spell inclusively; `Land` says the clause also allows a land play. An empty library does nothing.
2. **A nonland card within the cap** gets a `may_cast` under the existing `free_or_hand` keyword. Accepting grants §4's window over the card **where it sits in the library**: `Zone: ZoneLibrary`, `ScopeCards` (the object, so the position is not re-checked once the grant is made), `Cost: "{0}"`, `TimingFlash`, `CastOnly`, `MaxSpellManaValue`, and `LapseOnPass: LapseToHand`. `lapseCastPermissionLocked` already reads the zone the grant opens, so an uncast card goes from the library to its owner's hand on the holder's next pass. Declining puts it into the hand at once.
3. **A land** the player could play right now (`CanPlayLandDuringResolutionForEffect`: CR 305.2b land drop left, CR 305.3 their turn, CR 101.2 no "can't play lands") gets a `may_cast` under the new keyword `play_or_hand` (`MayCastKeywordPlayOrHand`). Accepting plays it through `PlayLandDuringResolutionForEffect`, a real land play that spends the drop (CR 305.2a); if the play ends up ignored the card goes to the hand. Declining puts it into the hand.
4. **Everything else** (a land with no drop left, a spell over the cap, a land when the card does not allow the play, an unreadable cost) goes straight to the hand with no question: the player could not have played it, so "if you don't" is already true.
5. **The look is private (CR 701.20).** `LookAtTopOfLibraryForEffect` marks only the looker a knower. `MayCastPrompt.Private` marks the offer, because its question names the card: the existing `PendingChoice.private` / `private_text` pair (ADR 0133) already strips words from every non-chooser, and the view now also clears `may_cast_card` for a private `may_cast` and words the neutral line for it, so the card's instance id is not a handle on a hidden-zone card. A reveal (Rashmi) is public and the prompt is not private.
6. **Client.** `mayCast.ts` words `play_or_hand` ("Play it" / "Put it into your hand"), which does **not** open a cast; `free_or_hand` still does. Board's free-cast request (`freeCastLocation`) now looks in the viewer's own library as well as exile and casts with `from_zone: "library"`, and the may_cast prompt finds its card there too. A looked-at library card is on the wire for its looker only.
7. **Bot.** The enumerator labels a `play_or_hand` accept "play it". The heuristic's yes stands: either branch is a free card.

The SANDBOX SIMPLIFICATION of §Decision 4 stands: the cast is an ordinary `cast_spell` after the resolution rather than a frame inside it. The trigger's controller is the active player (combat damage), so CR 117.3b gives them priority first.

**Cards.** Ziatora's Envoy (Full, with blitz from #3772), Rashmi, Eternities Crafter (Full). Both are one file and one call.

**Back-out.** Dropping `LapseOnPass` from the grant fails `TestZiatorasEnvoyPassingWithoutCastingPutsItInHand`; clearing `Private` fails `TestZiatorasEnvoyTheLookIsPrivate` and `TestALookedAtTopCardOfferIsPrivateToItsChooser`; offering a land without the land-drop gate fails `TestZiatorasEnvoyNoLandDropLeftPutsTheLandInHand`; removing the mana-value cap fails `TestZiatorasEnvoySpellAboveTheDamageGoesToHand`.

**Snapshot impact.** None. The grant is the existing `CastPermission` with fields already persisted, and a `private` prompt carries a continuation frame, so it is never written to a restore point.

## Amendment 2026-10-15: any number of free casts from hand, and a free cast from among the cards you looked at (#2110)

The two cards left on the #3270 row. Aetherflux Conduit: "Draw seven cards. You may cast any number of spells from your hand without paying their mana costs." Aetherworks Marvel: "Look at the top six cards of your library. You may cast a spell from among them without paying its mana cost. Put the rest on the bottom of your library in a random order." Both are §Decision 4's window with a different set of cards under it.

**Decision.**

1. **"Any number" is one grant per card, no prompt.** `effects.grantFreeHandCastsFrom` (`effects/free_hand_cast.go`) stamps `Game.FreeCastFromHandPermission` over every nonland card in the hand, from the draw's continuation (`DrawNThenForEffect` with a `DrawThen`, as Cecily's), so the seven new cards are included and a draw that stops on a prompt grants over the hand the player really ends up with. One grant per card, not one grant naming all of them: a `ScopeCards` permission is consumed by the first cast it pays for (`consumePassClosedGrantLocked`), which would end "any number" after one. The holder's next pass lapses every grant together (`LapseStaysInHand`) and the cards stay in hand. `offerFreeHandCastFrom` and this share `freeCastHandCandidates`, so "a nonland card" has one definition. The Great Synthesis keeps its own key and flicker and is not touched.
2. **A look is a private pick over the library.** `effects.lookAtTopAndOfferFreeCast` (`effects/free_library_cast.go`) looks at the top N (`LookAtTopOfLibraryForEffect`: only the looker becomes a knower), queues a `choose_cards` over the nonland cards with `Zone: ZoneLibrary` and the new destination `game.ChooseCastFreeFromLibrary` (see 3), bottoms the rest in a random order (`PutOnBottomInRandomOrderForEffect`, which also clears the knowers: the looker knows the cards are in the bottom few, not where), and grants `Game.FreeCastFromLibraryPermission` over the pick where it sits, now the top card of the library. The permission is `FreeCastFromHandPermission`'s for a library card (`{0}`, `TimingFlash`, `CastOnly`) with `LapseToLibraryBottom`: a pick the holder passes on goes to the bottom with the rest, which is what "put the rest" says when the answer is none. It is Sunbird's Invocation's finish for a reveal; the differences are the private look and no mana value cap. No spell among the six queues no prompt, and the six still go under.
3. **Wire.** `choose_destination: "cast_free_library"` (`game.ChooseCastFreeFromLibrary`), an open-vocabulary string like its siblings. The client opens the pick's cast chain from the viewer's library pile through the existing `freeCastRequest` (`freeCastLocation` already looks in the viewer's library, #3801), not through `handFreeCastRequest`, and words the prompt ("the rest go on the bottom"). The bot needs nothing new: `valuePutOntoBattlefield` has no opinion about a candidate that is not in hand, so the pick falls to `valueTakenFromLibrary` (name the best card) and the enumerator already offers the claimed library cast. Nothing about the prompt, the cards or the permission reaches another seat: the prompt is chooser-only, the cards are known to the looker alone, and the library card carries `castable_here` only on its owner's wire.
4. **The energy half.** Conduit's trigger reads the mana spent off the payment record when the spell is cast (Abby's read) and carries it on the item. A spell paid for nothing, or on paper (a waived charge), gives no energy, so the card is `CompletenessCaveats` for the same reason Abby is. Marvel's trigger is `APermanentYouControlWasPutIntoAGraveyard`, the existing shared test with no type list: a departure read, so a stolen permanent counts for the thief, a token counts, a bounce does not.

The SANDBOX SIMPLIFICATION of §Decision 4 stands: each cast is an ordinary `cast_spell` taken after the resolution rather than a frame inside it.

**Cards.** Aetherflux Conduit (Caveats), Aetherworks Marvel (Full).

**Back-out.** Granting only the first hand card (`break` in `grantFreeHandCastsFrom`) fails `TestAetherfluxConduitDrawsSevenAndCastsAnyNumberOfSpellsFree`; dropping `LapseOnPass` from the library permission fails `TestAetherworksMarvelUncastPickAndDeclinedLookGoToTheBottom`; dropping the `ChooseCastFreeFromLibrary` destination fails `TestAetherworksMarvelPaysSixLooksAtSixAndCastsOneFreeOfThem` and the view test `TestLibraryFreeCastPickPromptDeclaresTheCastToItsChooserOnly`; `PayEnergy(0)` fails `TestAetherfluxConduitRefusesWithFortyNineEnergy`; a zero amount on Conduit's item fails `TestAetherfluxConduitCastingASpellGivesEnergyEqualToTheManaSpent`; removing the empty-type-list shortcut fails `TestAetherworksMarvelGivesEnergyForEachPermanentYouControlPutIntoAGraveyard`. `TestALookedAtLibraryPickIsCastableForItsOwnerOnly` pins that no other seat sees the pick or the grant.

**Snapshot impact.** None new. The grants are the existing `CastPermission`, and `cast_free_library` is one more value of the open `ChooseDestination` string a restore carries as itself.

## Amendment 2026-10-23: the cast is made during the resolution (#4243)

Cascade's and discover's "you may cast it" is no longer a `may_cast` yes/no followed by a pass-closed cast after the resolution.
Both ask the `resolution_cast` prompt of [ADR 0066](0066-granted-cast-and-play-permissions.md)'s 2026-10-23 amendment, inside the
resolution (CR 608.2g): the hit is cast there or declined there, and the passed-over cards are bottomed when it is answered.
The permission this ADR describes (§3, §4) is unchanged and still the data under that prompt; the pass-closed window is now only
the fallback for an offer made with no open resolution. `DiscoverThenForEffect` and `effects.Discover.Then` are replaced by
`DiscoverForEffect` and a step record (`effects.DiscoverStepKey`).
