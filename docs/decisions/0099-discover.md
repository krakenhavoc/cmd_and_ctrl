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
