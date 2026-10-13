# ADR 0152 — More Than Meets the Eye, convert and living metal (CR 702.162, 701.28, 702.161)

**Status:** Accepted (owner decisions 2026-10-13) · 2026-10-13 · S53 — Faces, zones, and shared payment (ADR first), tracker [#1910](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1910)
**Issues:** [#1856](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1856) (seam: More Than Meets the Eye, convert and living metal — Goldbug, Humanity's Ally). Found by [ADR 0106](0106-five-small-seams-from-the-s50-rechecks.md) PR 3 (#1806), S50 tracker #1784. Registry row: `more-than-meets-the-eye` (missing; the implementation closes it).
**Owner decisions:** the owner answered all five questions on 2026-10-13, each with the recommended option (a). The answers are recorded inline under [Questions for the owner (answered 2026-10-13)](#questions-for-the-owner-answered-2026-10-13) and are binding on the implementation PRs in §8.
**Numbering:** assigned by the S53 plan. Before writing, I checked every ref with `git log --all -- 'docs/decisions/015*'` after `git fetch`: no branch has a 015x file. The highest number on `origin/develop` is 0149. This ADR is **0152**.
**Builds on:** [ADR 0034](0034-multi-face-cards.md) (the face model), [ADR 0079](0079-transforming-a-permanent.md) (transform in place; "return it transformed"), [ADR 0107](0107-state-triggers-rebound-disturb-and-damage-prevention.md) §4 (disturb: `AlternativeCost.CastsFace`), [ADR 0118](0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md) (alternative costs from any zone), [ADR 0126](0126-bots-that-play-their-decks.md) (the heuristic prices what a card does), [ADR 0132](0132-day-and-night.md) (daybound's "can't transform except").

This ADR was written design-first. No code changed with it.

---

## Context

### The rules

From the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`):

- **CR 702.162a, More Than Meets the Eye:** "represents a static ability that functions in any zone from which the spell may be cast. 'More Than Meets the Eye [cost]' means 'You may cast this card converted by paying [cost] rather than its mana cost.'" It follows the rules for alternative costs (CR 601.2b, 601.2f–h): it can't be combined with another alternative cost, and additional costs and the commander tax still apply.
- **CR 701.28a, convert:** "To convert a permanent, turn it so that its other face is up. This follows rules 701.27a–f, 712.9–10, and 712.18. Those rules apply to converting a permanent just as they apply to transforming a permanent."
- **CR 701.28b:** converting is not turning face up or face down. **701.28c:** an instruction to convert a permanent that isn't a double-faced card or token does nothing. **701.28d:** nor does one whose other face is an instant or sorcery face.
- **CR 701.28e** (the twin of **701.27f**): an activated or triggered ability of a permanent that tries to convert it does so only if the permanent "hasn't converted or transformed since the ability was put onto the stack". For a delayed trigger, the reference point is when the delayed trigger was created. "If the permanent has already transformed or converted, an instruction to do either is ignored."
- **CR 701.28f:** "If a spell or ability states that a permanent can't transform, that permanent also can't convert."
- **CR 701.27e:** a "transforms into" trigger fires when the object "either transforms or converts" and has the named characteristic afterwards. This is the only place CR 701.27 lets a convert count as a transform. A plain "whenever … transforms" does not see a convert (CR 701.28 joins the two actions only through the rules it lists).
- **CR 702.161a, living metal:** "During your turn, this permanent is an artifact creature in addition to its other types."
- **CR 712.8c / 712.11a / 712.11d:** a card cast "converted" is put on the stack back face up, with only its back face's characteristics, and its mana value is its front face's. The front face's ability that allows the converted cast is also considered when deciding whether it can be cast.
- **CR 712.13:** a resolving double-faced spell enters with the face that was up on the stack. **712.14a:** "put onto the battlefield converted" enters back face up.
- **CR 712.18:** a permanent that transforms or converts does not become a new object. **712.20:** an "As [this] transforms" clause also applies as it converts.
- **CR 301.7a–b:** a Vehicle has its printed power and toughness only while it is a creature.

### What the engine has

Most of the work exists already.

- **The face model and the in-place turn-over.** `game.TransformPermanentForEffect` (`server/internal/game/transform.go`) checks `CanTransform` (CR 712.9, 712.10, 712.15a) and daybound/nightbound (CR 702.145b), then calls `transformInPlaceLocked`. That flips `ActiveFace` with `Card.SetFace`, drops the effective cache, emits `EventTransform` (which bumps the layer version in `layer_listener.go`) and runs the new face's `AsTransformsInto`. Convert is that verb under another name (CR 701.28a, 712.18).
- **An alternative cost that casts the back face.** Disturb added `AlternativeCost.CastsFace` (`game/alternative_cost.go`). `faceForClaimLocked` and `frontFaceCastOffer` (`game/face.go`) put the card on the stack back face up while judging the claim off the front face (CR 712.11d). `faceOnResolve` keeps that face on the battlefield. `AlternativeCost.CastFaceOf` lets the legal-move enumerator (`legal/cast.go`) and the view's offer stamp (`protocol/view.go`, `castGateFace`) read the back face's targets and gates. `effects.checkCastsFace` (`cards/effects/registry.go`) rejects the malformed shapes at Register time. A disturb offer is bound to the graveyard by `FromZone`. With `FromZone` empty, `validateCastPathLocked` (`game/cast_zones.go`) lets an offer be claimed from the hand, the command zone, or any zone a permission opens. That is CR 702.162a's "any zone from which the spell may be cast".
- **Mana value.** `manaCostForValue` (`game/cost_modifier.go`) already reads a `transform` card's front-face cost when its back face is up (CR 712.8c, 712.8e).
- **"Return it converted."** `ReturnFromGraveyardTransformedForEffect` and `ExileAndReturnTransformedForEffect` (`game/transform.go`) put a card onto the battlefield back face up (CR 712.14a).
- **"During your turn" statics.** `firstStrikeDuringYourTurn` (`cards/effects/first_strike_your_turn.go`) is a layer static whose `AppliesTo` reads the active player. `onTurnBeganLocked` bumps the layer version on every turn change, so such a static is re-evaluated without any extra code. Crew's `BecomeCreatureUntilEOT` (`cards/effects/vehicles.go`) shows the layer-4 half: add `Artifact` and `Creature` and leave the printed power and toughness alone.

What is missing:

1. The keyword offer: no `MoreThanMeetsTheEye` constructor, and the keyword is not in `canonicalKeywords` (`game/keywords.go`).
2. The verb: nothing converts, and nothing in the log or the events tells a convert from a transform.
3. CR 701.27f / 701.28e as an engine rule. The current transform cards check it themselves, each in its own way (`garrukTransform` checks `ActiveFace != 0`, and Growing Rites of Itlimoc relies on the same face check). Transformers turn over back and forth, often several times in a turn, so a face check is no longer enough. Two "convert it" triggers from one double-strike hit are the common case.
4. Living metal: no static, no keyword.
5. The bot prices every cast as the card's front face. The move carries the face (`legal` sets it to the back face for a `CastsFace` offer), but the heuristic's `castParams` (`aiseat/heuristic/params.go`) does not read it, and `AlternativeCostView` does not say which face an offer casts. Disturb got away with this because a disturb cast is the card's only cast from the graveyard. A More Than Meets the Eye cast sits beside the printed cast in hand, and the two produce different permanents.

### The cards

The Scryfall dump the repo uses (`data/scryfall/default-cards.json`, refreshed 2026-09-24) has **15** commander-legal cards with More Than Meets the Eye, all `transform` layout, all from *Transformers* (`bot`). 13 of them have living metal on the back face. Blaster and Soundwave have a noncreature artifact back face with no living metal. Every one of the 15 converts itself, and nothing in the dump converts another permanent. The other cards that mention "convert" (Grimlock, Dinobot Leader, the Un-cards and one Alchemy token) are not commander-legal or say "convert" in another sense. None of the 15 is in the catalog today.

| Card | Living metal | Text beyond the three keywords, to gap-check when its card file is written |
|---|---|---|
| Goldbug, Humanity's Ally // Goldbug, Scrappy Scout | yes | prevent combat damage to attacking Humans; "your second spell each turn"; Human spells can't be countered (ADR 0106 PR 2); "Goldbug and at least one Human attack" |
| Arcee, Sharpshooter // Arcee, Acrobatic Coupe | yes | "remove one or more +1/+1 counters" as a variable cost; "a spell that targets one or more creatures or Vehicles you control" |
| Blaster, Combat DJ // Blaster, Morale Booster | no | grants modular 1; modular 3; move X counters |
| Blitzwing, Cruel Tormentor // Blitzwing, Adaptive Assailant | yes | life lost this turn; a keyword chosen at random |
| Cyclonus, the Saboteur // Cyclonus, Cybertronian Fighter | yes | connive; an additional beginning phase (Sphinx of the Second Sun's `AddPhases`) |
| Flamewar, Brash Veteran // Flamewar, Streetwise Operative | yes | intel counters on face-down exiled cards; "discard your hand" as a cost |
| Jetfire, Ingenious Scientist // Jetfire, Air Guardian | yes | remove counters from among artifacts; restricted mana; adapt |
| Megatron, Tyrant // Megatron, Destructive Force | yes | opponents can't cast spells during combat; {C} per life lost; excess damage |
| Optimus Prime, Hero // Optimus Prime, Autobot Leader | yes | bolster; "return it to the battlefield converted" (CR 712.14a); "whenever you attack" |
| Prowl, Stoic Strategist // Prowl, Pursuit Vehicle | yes | the owner may play the exiled card; "a card exiled with Prowl" is played; "the second time this ability has resolved this turn" |
| Ratchet, Field Medic // Ratchet, Rescue Racer | yes | life gained this turn; "triggers only once each turn" |
| Slicer, Hired Muscle // Slicer, High-Speed Antagonist | yes | give control until end of turn, goad, "can't be sacrificed this turn"; a delayed "convert it at end of combat" |
| Soundwave, Sonic Spy // Soundwave, Superior Captain | no | copy a card from a graveyard and cast the copy; odd/even mana value triggers; named tokens |
| Starscream, Power Hungry // Starscream, Seeker Leader | yes | the monarch |
| Ultra Magnus, Tactician // Ultra Magnus, Armored Carrier | yes | put onto the battlefield tapped and attacking; formidable; a delayed "convert … at end of combat" |

The issue said "~15 cards". There are exactly 15, and the issue's one waiting card, Goldbug, is among them.

## Decisions

### 1. More Than Meets the Eye is disturb's offer without the zone

`effects.MoreThanMeetsTheEye(cost string) game.AlternativeCost` returns

```go
game.AlternativeCost{
	Key:       "more_than_meets_the_eye",
	Label:     "More Than Meets the Eye " + cost,
	ManaCost:  cost,
	CastsFace: 1,
}
```

and goes on the front face's entry (the bare oracle ID). It is `effects.Disturb` with `FromZone` left empty, so the offer can be claimed wherever the card can be cast: from the hand, from the command zone (all 15 are legendary and most are commanders), and from exile under a permission that does not insist on its own price. CR 702.162a says the ability "functions in any zone from which the spell may be cast", and the empty `FromZone` already implements that.

Nothing in the cast path changes. `checkCastsFace` already enforces the offer's shape. `faceForClaimLocked` already accepts face 1 or "unset" for a `transform` card and refuses everything else. `castPathKey` already reads the zone off the front face (CR 712.11d). `faceOnResolve` already keeps the back face (CR 712.13). The commander tax is added the usual way, because the tax applies to any cast from the command zone whatever price is paid (CR 903.8). The spell's mana value on the stack is the front face's (`manaCostForValue`, CR 712.8c), so a "counter target spell with mana value 3 or less" judges a converted Goldbug as 3.

`more than meets the eye` joins `canonicalKeywords`, and `cards/coverage/caveats.go` gains a row of disturb's shape (`altCost("more_than_meets_the_eye")`), so the coverage detector reports a card with the keyword and no offer.

### 2. Convert is the transform verb with one flag

`Game.ConvertPermanentForEffect(cardID uuid.UUID) error` is `TransformPermanentForEffect` with the same guards, in the same file. The two share one body, `turnOverInPlaceLocked(card *Card, converted bool)`, which is `transformInPlaceLocked` renamed. The guards are the same because CR 701.28a imports them:

- `CanTransform` (CR 712.9, 712.10, 712.15a; ADR 0145's meld refusal) answers "nothing happens" (CR 701.28c, 701.28d).
- `hasDayNightKeyword` refuses daybound and nightbound permanents. CR 702.145b's "can't transform except due to its daybound ability" also means it can't convert (CR 701.28f). No Transformer has daybound, but a copy effect could combine them.

The turn-over emits `EventTransform` with a new `Event.Converted bool`. **Question 1** asks whether that should be a separate event kind instead. Every existing reader of `EventTransform` is something a convert should also reach:

- `layerVersionBump` (the printed characteristics changed);
- the three catalog triggers that watch it (Brutal Cathar, Howlpack Piper, Concealing Curtains), all "transforms into" triggers, which CR 701.27e says fire on a convert too;
- `runAsTransformsIntoLocked`, which CR 712.20 applies to a convert.

A future card with a plain "whenever … transforms" checks `!ev.Converted`. The log narrates `LogTransform` with a `converted` flag, and the client says "Goldbug, Humanity's Ally converted into Goldbug, Scrappy Scout" rather than "transformed".

Card files spell it `effects.ConvertThis{}`, the twin of `TransformThis`. `effects.Convert{Target}` exists for symmetry, but no printed card converts another permanent. "Return it to the battlefield converted" (Optimus Prime) reuses `ReturnFromGraveyardTransformedForEffect`. That is CR 712.14a, a new object, and no convert happens.

### 3. CR 701.27f and 701.28e become an engine rule

A permanent remembers how many times it has turned over: `Card.TurnOvers int`, incremented by `turnOverInPlaceLocked` on both transform and convert. It is not reset on a zone move, because the object check already covers that (`SourceObject`, #1418, CR 400.7). An ability item records the count when it is put on the stack: `StackItem.SourceTurnOvers`, set in the same places `SourceObject` is (`stampTriggerSource`, the two activation paths, the ability copy). A delayed trigger records it when it is created (`DelayedTrigger.SourceTurnOvers`, set by `ScheduleDelayedTriggerForEffect` from the resolving item's source). `effects.ConvertThis` and `effects.TransformThis` then do nothing if the source's count has moved since. That is CR 701.28e's "an instruction to do either is ignored".

This is what makes the common Transformers case right. Blitzwing back face up has double strike from an Equipment. It hits a player twice, so two "convert it" triggers go on the stack. The first one turns it face up, and the second one does nothing. A face check (Garruk's) gets this case right too. But a face check gets it wrong when the permanent has turned over twice since the trigger, which Arcee and Jetfire can do at instant speed with their own activated abilities. Only the count matches the rule. **Question 2** offers the cheaper alternatives.

Both new fields are additive carried fields of snapshot schema v7, plain ints with no closure route. Garruk Relentless and Growing Rites of Itlimoc can drop their own face checks later. This ADR does not touch them.

### 4. Living metal is a catalog static on the back face

`effects.LivingMetal() game.StaticAbility` is a layer-4 static on the face that prints it (`<oracle_id>#1`). `AppliesTo` is "this object, and its controller is the active player". `Apply` adds `Artifact` and `Creature` to the types, like `BecomeCreatureUntilEOT`, and sets no power or toughness: the printed 1/3 on Goldbug, Scrappy Scout is the face's own and applies as soon as it is a creature (CR 301.7b). The turn-change bump re-evaluates it, as it does for Razorkin Needlehead's first strike.

What follows comes from existing code, not new code:

- On an opponent's turn the Vehicle is not a creature. It can't block, it isn't hit by "destroy all creatures", and combat damage can't be dealt to it. It does not fall off as a 0-toughness creature, because its printed power and toughness don't apply (CR 301.7a).
- Summoning sickness is about continuous control since the turn began (`HasSummoningSickness`, CR 302.6), not about when it became a creature. So a Goldbug cast converted this turn can't attack without haste, and one converted in place from a creature that was already there can.
- A "creature dies" trigger reads `Event.LastKnownTypes` (#1675). A Vehicle that dies on its own turn died as a creature, and one that dies on an opponent's turn did not.

`living metal` joins `canonicalKeywords` for the badge and the coverage signal. **Question 3** asks whether the engine should derive the static from the keyword instead of the card file declaring it.

### 5. The wire and the client

- `AlternativeCostView` gains `casts_face int` (omitempty). It is set from `AlternativeCost.CastsFace`, so disturb starts sending it too. The alternative-cost picker (`AlternativeCostModal.svelte`) shows the face's name under the label ("More Than Meets the Eye {W}{U} — casts Goldbug, Scrappy Scout") and uses that face for the preview. The cast payload is unchanged: the client sends the key and no face, which `faceForClaimLocked` accepts, as `disturb.test.ts` already confirms.
- The log line from §2 says "converted".
- Nothing else. The hover panel already shows both faces of a `transform` card. The board tile shows whichever face is up. A living-metal Vehicle shows as a creature on its controller's turn because the tile reads the effective types.

### 6. The bot

The heuristic has to price a converted cast as the permanent it produces. Otherwise it always takes the cheaper More Than Meets the Eye price for what it thinks is the same card. That is right for Goldbug, where {W}{U} buys a 1/3 instead of {1}{W}{U} for a 3/3, but wrong for Megatron, where 4 mana buys a 4/5 Vehicle instead of 6 mana for a 7/5 creature.

- `castParams` reads the move's `face`, and `valueOfCast` prices `resolvedValueFor` on that face's `CardFaceView`. That view's type line, power and toughness are already on the wire. A face with living metal is priced as a creature of its printed size, because it is one on every turn the bot attacks with it. Its value as a blocker is not counted.
- "You may convert X" uses whatever the card's optional-effect prompt is, and the heuristic answers it with its existing yes/no default. Card PRs measure it per card, as ADR 0126 asks.
- Tests: the move enumerator already offers the converted cast with face 1 (pinned in `legal`). A whole-game `aiseat` test in the shape of `disturb_game_test.go` has a bot holding Goldbug with three lands and checks that it is offered both prices and that whichever it takes resolves on the face it paid for.

**Question 4** asks whether this pricing is in scope now.

### 7. Testing

- `game`: convert keeps the object (same `InstanceID`, counters, damage, attachments, tapped state, timestamp; no zone move or ETB); `EventTransform{Converted: true}` bumps the layer version; CR 701.28c/d do nothing; a daybound permanent can't convert (CR 701.28f); the turn-over count stops a second trigger and a delayed trigger (CR 701.28e), and a trigger put on the stack after the turn-over still works; the count survives clone and restore.
- `cards/effects`: a converted cast from the hand and from the command zone (commander tax charged, back face on the stack, front-face mana value, back face on the battlefield); the printed cast still enters front face up; the offer can't be combined with another alternative cost; living metal makes the face a creature on its controller's turn only, is re-evaluated at the turn change, can't block on an opponent's turn, and keeps its printed power and toughness with a +1/+1 counter on it.
- `protocol`: `casts_face` is on the offer view; the log line says converted. `client`: the picker shows the face name (a `vitest` beside `disturb.test.ts`).
- `aiseat`: §6's whole-game test. `cards/coverage`: the two new detector rows. `roadmap`: the generated table is updated with `-update` when the row closes.
- A catalog game soak (`AISEAT_CATALOG_GAMES`) with the shipped cards in the pool, as for any new mechanic.

### 8. Delivery

1. **Engine:** §1–§4 plus their tests, the coverage rows and the keywords. Goldbug, Humanity's Ally ships as the proof card if its other text is already covered (Q5), and the registry row closes.
2. **Wire, client and bot:** §5 and §6.
3. **Cards:** the other 14, gap-checked one at a time against the table above. A card whose other text needs a new seam ships with a caveat or waits on its own issue, as usual.

## Out of scope

1. **Double-faced tokens.** CR 701.28c names them. The engine has none (ADR 0079 out of scope 1), and no Transformers card makes one.
2. **Converting another permanent.** No printed card does it. `effects.Convert{Target}` exists, but no card uses it.
3. **The `bot` set's single-faced reprints and the Un-set "convert" cards.** They are not commander-legal, or "convert" means something else on them.
4. **Removing the per-card face checks** from Garruk Relentless and Growing Rites of Itlimoc. They become redundant after §3 and can go in a later cleanup.

## Alternatives considered

- **Convert as a plain transform, with no flag.** This is the smallest change, and it is right for every catalog card today. But the log would say "transformed" for a mechanic whose cards say "convert", and a future plain "whenever … transforms" card (CR 701.28 joins the two actions only through the rules it lists) would wrongly fire on a convert, with no way to tell the two apart. The flag costs one bool.
- **A separate `EventConvert` kind.** This is the cleanest separation. But every existing `EventTransform` reader would need a second case (the layer bump, the three "transforms into" triggers, `AsTransformsInto` per CR 712.20), and a reader that forgets the second case fails silently. This is Question 1's option (b).
- **More Than Meets the Eye as a `CastPermission`.** A permission is a right granted by an effect. More Than Meets the Eye is a price the card prints. Disturb already settled this.
- **Living metal as a scoped effect applied at the start of each turn.** That stores state for something that is a pure read of whose turn it is, and needs a cleanup on a control change mid-turn. The static needs neither.
- **Leaving CR 701.28e to each card file.** That is what transform does today. With 15 cards that turn over in both directions, that means 15 hand-written checks, and the double-flip case would still be wrong.

## Questions for the owner (answered 2026-10-13)

Each question lists the recommended option first. The owner chose (a), the recommended option, on all five questions on 2026-10-13. The options not chosen are kept for the record.

1. **How a convert appears in events (§2; CR 701.27e, 701.28a, 712.20).**
   - **(a) Recommended:** `EventTransform` with `Converted: true`. The layer bump, "transforms into" triggers and `AsTransformsInto` work with no change, and a plain "whenever … transforms" reader checks `!Converted`.
   - (b) A new `EventConvert` kind, added to every reader that should see it. Cleaner to read, but a reader that forgets the second case fails silently.
   - (c) No distinction: a convert is a transform. The log says "transformed". This is wrong for a future plain "transforms" trigger.
   - **Decision:** (a). A convert is `EventTransform` with `Converted: true`.

2. **CR 701.27f / 701.28e: "if it has already converted, the instruction is ignored" (§3).**
   - **(a) Recommended:** an engine rule. `Card.TurnOvers`, recorded on stack items and delayed triggers, is checked by `ConvertThis` and `TransformThis`. Two additive v7 fields. Matches the rule exactly, including two turn-overs in a row.
   - (b) A face check in `ConvertThis`: convert only if the source still shows the face whose ability this is. No new state. Right for the double-strike case, wrong when the permanent turned over twice since the ability went on the stack.
   - (c) Nothing. Two "convert it" triggers would flip it twice.
   - **Decision:** (a). CR 701.27f / 701.28e is an engine rule: `Card.TurnOvers`, recorded on stack items and delayed triggers, is checked by `ConvertThis` and `TransformThis`.

3. **Where living metal comes from (§4; CR 702.161a).**
   - **(a) Recommended:** `effects.LivingMetal()`, declared on the back face's entry by the card file, with the keyword in `canonicalKeywords` for the badge. This is how flashback and disturb work, and it is reviewed together with the rest of the card.
   - (b) Derived by the engine from the printed keyword (the way phasing is read from `Effective().Abilities`), so a back face with the keyword works with no card file. That covers the 13 Vehicles without card files, but none of them can be cast converted or convert without one anyway, and it adds a layer-4 effect that depends on a layer-6 read.
   - **Decision:** (a). Living metal is `effects.LivingMetal()`, declared by the card file on the back face's entry, and the keyword joins `canonicalKeywords`.

4. **The bot's price for a converted cast (§6).**
   - **(a) Recommended:** in scope: `casts_face` on the offer, and the heuristic prices the face the move casts. Without it the bot always pays the cheaper price, which is right for Goldbug and wrong for Megatron and Ultra Magnus.
   - (b) Out of scope for now: the bot prices both casts as the front face, with a caveat on each card, and a follow-up issue.
   - **Decision:** (a). The bot prices converted casts now: `casts_face` goes on the offer view, and the heuristic prices the face the move casts.

5. **How many cards ship with the seam (§8).**
   - **(a) Recommended:** the engine PR ships Goldbug, Humanity's Ally (the issue's waiting card) if its other text is already covered, and one card issue is filed for the other 14, to be gap-checked one at a time.
   - (b) All 15 in this sprint, each card's other text included. Several need seams of their own (Soundwave's cast of a copy from a graveyard, Flamewar's intel counters, Slicer's "can't be sacrificed"), so this is several sprints of work, not one.
   - (c) The engine only, with no proof card. That leaves the registry row's `Waiting` empty with nothing showing that the seam works.
   - **Decision:** (a). Goldbug, Humanity's Ally is the proof card in the engine PR, and one card issue is filed for the other 14.

## Consequences

**Cards.** Once the engine PR merges, 15 commander-legal cards can be cast converted, convert, and attack on their own turn as Vehicles. How many of them ship Full depends on each one's other text.

**Cost.** One constructor and one keyword for More Than Meets the Eye; one verb, one event flag and one log flag for convert; one static and one keyword for living metal; two additive snapshot ints for CR 701.28e; one wire field; a face read in the heuristic. No new prompt kind and no schema bump.

**The thing most likely to be got wrong later.** Using `ExileAndReturnTransformed` or `ReturnFromGraveyardTransformed` for a printed "convert it". Those make a new object (CR 400.7). A convert is the same object turning over (CR 712.18): its counters, damage and attachments stay, and nothing enters.
