# ADR 0072 — Protection (CR 702.16)

**Status:** Accepted · 2026-09-18 · S40 — S30's tail: protection, regeneration, emblems, winning by effect
**Issue:** [#662](https://github.com/krakenhavoc/cmd_and_ctrl/issues/662) · Refs [#883](https://github.com/krakenhavoc/cmd_and_ctrl/issues/883)
**Supersedes:** [ADR 0038 §7](0038-protection-style-keywords.md)'s protection
paragraph and its S30 follow-up's "Protection is unchanged and still absent".
Ward is untouched; everything ADR 0038 says about ward still stands.
**Numbering:** re-checked immediately before the push, after a full
`git fetch --all --prune`. Every `docs/decisions/` path ever touched on any
remote ref was listed from `git log --all --name-only --diff-filter=AM`
(296 remote heads). The highest number in existence anywhere was 0071
(designations, #969), which reached `develop` while this branch was open —
so `develop` carries up to 0071 and this ADR was rebased onto it without
renumbering. 0072 was free on every ref.

## Context

Protection is the last of the CR 702 protection-style keywords with
no implementation, and ADR 0038 §7 said why: it is the only one whose
test is against the **source object** rather than against a
controller, and the engine's targeting choke point never receives
one.

It is four checks, and the acronym is the issue's: **DEBT**.

| | rule | the hook, on `develop` `b78afff7` |
|---|---|---|
| **D**amage from a source with the quality is prevented | CR 702.16e | `RepEventDamage` carries `DamageSource`, an ID. The source may be a spell (never on the battlefield) or a permanent that has already left |
| **E**nchanted / equipped / fortified by an object with the quality | CR 702.16c-d | `attachmentLegalLocked` has the Aura or Equipment itself (`game/attach.go`), but a catalogued Aura returns early before any new check |
| **B**locked by a creature with the quality | CR 702.16f | `BlockPairRefusalLocked` has both creatures, and #705 left slot 3 of its check order reserved for exactly this |
| **T**argeted by a spell with the quality, or by an ability from a source with it | CR 702.16b | `CanBeTargetedBy(c, zone, caster uuid.UUID)` and `specMatchLocked(caster, …)` get only the **controller's ID** |

Three of the four already hold the object they need. The fourth is a
refactor, and it is the reason this is one ADR rather than four card
files.

Two structural facts shape every decision below.

**A quality is a parameter, and `Characteristic.Abilities` is a
`[]string` of bare tokens.** ADR 0038 treated that as a blocker. It
is not: a token can carry its parameter, as long as exactly one
reader parses it and the grammar it parses is closed.

**A source can be gone by the time the rule is asked.** Combat damage
is snapshotted into a `damageTail` at event-creation time for exactly
this reason (#694: a CR 616 ordering prompt is answered after the
rest of the damage step has landed and been swept). A spell source is
never on the battlefield at all. Reading "the card's characteristics
in the graveyard" is the wrong answer in both cases — a pumped,
colour-shifted attacker is not the card that fell into the yard.

## Decisions

### 1. Representation: parameterised tokens, one closed-grammar reader

Protection lives in `Characteristic.Abilities` as
`protection from <quality>` — one token per quality, so Baneslayer
Angel carries two:

```
protection from Demons
protection from Dragons
```

`game/protection.go` is the **one** reader. Nothing else in the
engine, the catalog, the bot or the client splits a protection string.

```go
func ProtectionQualities(c *Card) []ProtectionQuality   // the reader
func ProtectionTokens(printed string) ([]string, bool)  // the parser
func ProtectedFrom(protected *Card, src *Characteristic) bool
```

The grammar is CLOSED, in the same sense `canonicalKeywords` is
closed: a shape joins it in the change that teaches the engine to
honour it.

| quality | token | matched against |
|---|---|---|
| colour | `protection from red` | the source's effective `Colors` |
| card type | `protection from artifacts` | the source's effective `Types` |
| subtype | `protection from Demons` | the source's effective `Subtypes`, plus `AllCreatureTypes` (#939) when the subtype is a creature type |
| everything | `protection from everything` | matches every source (CR 702.16j) |

Plurals are printed, not invented: the token quotes the card
("protection from **Demons** and from **Dragons**", "protection from
**artifacts**"), and the reader singularises. Storing a canonical
singular instead would have made the client's badge tooltip read
"Protection from Demon", and the tooltip shows the quality.

**Anything the grammar cannot parse mints no token.** "Protection
from monocolored", "protection from each of the chosen colors",
"protection from opponents": each yields nothing, the permanent gets
no protection at all, and the card keeps its ADR 0037 unimplemented
flag because `keywordOnlyLine` calls the same parser. That is the
Hexproof-from posture of ADR 0038 §6, and it errs weaker.

`canonicalKeywords` gains `protection`, in this change, as the table
documents it must be. It is the table's one **parameterised** entry:
a bare `protection` is never a wire token and `CanonicalKeywords`
refuses it explicitly, because a quality-less badge is a promise the
rules layer cannot keep.

`CanonicalKeyword` grew a plural sibling for this and only this:
`CanonicalKeywords(part) ([]string, bool)`, because one printed
clause — "protection from Demons and from Dragons" — is two tokens.
The three readers that scan printed text (`deck.printedKeywords`,
`deck.keywordLines`, `coverage.keywordOnlyLine`) use the plural form;
every other caller keeps the singular wrapper.

### 2. The targeting source: one struct, one choke point, every caller

`game.TargetSource` replaces the bare `caster uuid.UUID` on the
targeting half of `targets.go`:

```go
type TargetSource struct {
	Controller uuid.UUID     // who is choosing (CR 601.2c), always set
	Object     *Card         // the live source object, when there is one
	Snapshot   *Characteristic // a value copy, when the object may be gone
}
```

`Controller` is what the predicates (`CardOK`, `PlayerOK`) receive, so
**no catalog file changes** — the ~300 `Targets:` clauses are
untouched. `Object` / `Snapshot` are what CR 702.16b reads. Exactly
one of the three constructors is used at each call site:

- `SourceObject(controller, card)` — a live spell or ability source
- `SourceSnapshot(controller, chars)` — a value copy the caller kept
- `SourceChooser(controller)` — **no source**, for the cost-payment
  enumerations that reuse `TargetSpec` as a predicate and do not
  target (ADR 0038 §4). `specMatchLocked(…, targeting=false)` never
  reaches the protection check, so a source-less value there is not a
  hole; it is the declaration that paying a cost is not targeting.

Every targeting call site now names its source. The list is the
whole of it, and it is in the ADR because the next person to add a
targeting path needs to know they owe one:

| call site | source |
|---|---|
| cast announce (`mutations.go` `CastSpell`) | the spell card |
| activated-ability announce (`activated.go`) | the ability's source permanent |
| trigger target pick (`trigger_target_timing.go`, `pending_choice.go`) | `pickTargetFrame.source` (already a value copy) |
| pending-choice resume (`pending_choice.go` `ResolvePickTargets`) | the same frame |
| spell-copy re-target resume (`spell_copy.go`) | `copySpellFrame.src`, a **value copy** — the original can be countered between the prompt and the answer |
| resolution re-check (`TargetStillLegalForEffect`, CR 608.2b) | the stack item's card, looked up live and falling back to `StackItem.SourceLKI` |
| the view's legal-target stamps (`protocol/view.go`) | the card being cast / the ability's permanent |
| the bot's enumerator (`legal.legalTargetSets`, `legal/cast.go`) | the same |

`SpecCandidatesForEffect` keeps its `chooser uuid.UUID`. It is the
non-targeting sibling and gains nothing from a source.

### 3. Sources that have left: last-known information on the damage event

`ReplacementEvent` gains one field:

```go
// SourceLKI is the damage source's characteristics as they were when
// the event was CREATED (CR 608.2h / CR 702.16e).
SourceLKI *Characteristic
```

It is set in the two tail constructors that already snapshot
deathtouch, lifelink and commander status — `combatDamageTailLocked`
and `effectDamageTailLocked` — and in `damageTailFromFrame` from the
CR 510.1c prompt's frame, so all six damage entry points are covered
by the one `damageThroughReplacementsLocked` body that lands it on
the event.

`Characteristic.Controller` carries the source's controller inside
the snapshot, so nothing needs a second field for CR 702.16k.

Three things this buys that a `DamageSource` lookup cannot:

- **A spell source.** Lightning Bolt is never on the battlefield.
  `findBattlefieldCard` returns nil for it, which is why lifelink and
  deathtouch were only ever read for permanents. The LKI lookup walks
  every zone, so the Bolt's redness is on the event.
- **A source that died in the same damage step.** #694's window:
  the event is created, a CR 616 prompt pauses it, the rest of the
  step lands and the SBA sweep takes the attacker, and the prompt is
  answered. The snapshot is from before the pause.
- **A source whose characteristics changed on the way out.** A
  Giant Growth'd, Painter's Servant'd attacker is red on the
  battlefield and not red in the graveyard. CR 608.2h says use the
  last existing characteristics; the graveyard card is a new object.

### 4. Damage prevention: one engine built-in, applied first

`protectionPreventsDamageReplacement` (`builtin_replacements.go`) is
the CR 702.16e half — the third engine-owned replacement, written the
way #960's regeneration shield is, because protection belongs to the
rules and not to any object.

Two properties, and both are decisions rather than side effects.

**It applies FIRST, with no CR 616 ordering prompt.**
`ReplacementEffect` gains `Preemptive bool`, and
`applyReplacementsLocked` applies a preemptive effect before it
considers prompting. Protection is the only thing that sets it.

**It spends no shield.** This is the point of applying first, and it
is what #420 asks for: a Circle-of-Protection-style
`PreventNextDamage` with 4 charges facing a protected source must
still have 4 charges afterwards, because the damage it would have
absorbed never reached it.

**This is a DECLARED SIMPLIFICATION of CR 616.1.** By the letter,
protection is a prevention effect like any other, the affected
object's controller orders it against the others, and there are
boards where a player would rather a different one applied —
Phytohydra ("if damage would be dealt to this creature, put that many
+1/+1 counters on it instead") wants the replacement, not the
prevention. The engine takes the prompt away. The reasons:

1. Protection is total on the event (`ev.Cancel()`), so in every
   ordering where it applies at all, it is the last thing that
   happens to that event. The only observable difference is whether
   something else got to *charge itself* first.
2. Charging a shield against damage that was going to be prevented
   anyway is strictly worse for the protected player, and this repo's
   posture is to err weaker only against the player the rule is
   protecting — not for them.
3. The prompt is unanswerable copy. "Order these: protection from
   red; prevent the next 4 damage" has one sane answer and asking it
   every combat step is noise.

Phytohydra, and any future "damage would be dealt … instead" card
whose controller would prefer it over protection, is the known cost.
It is noted here rather than in a card file because no such card is
in the catalog.

**CR 615.12 ("this damage can't be prevented") is out of scope and is
not modelled anywhere,** including for protection. Banefire already
says so in its own caveat (`cards/effects/banefire.go`); nothing
suppresses a prevention effect in this engine, so unpreventable
damage is prevented by protection here. That is stronger than
printed for the protected creature and it is the pre-existing gap,
not a new one.

### 5. Blocking: #705's reserved slot

CR 702.16f goes in `BlockPairRefusalLocked` at position 3, the slot
ADR 0045's addendum reserved for it, with `BlockReasonProtection`
("protection") — a token that addendum also reserved. The check reads
the blocker's effective characteristics and controller, both of which
the function already has.

The declaration's #857 lock-in re-checks it exactly as it re-checks
every other pair rule, because it calls the same function. There is
no second copy and there must never be one (ADR 0045 §3).

The **attacking** half only. CR 702.16f is about a creature with
protection that is attacking or blocking; "can't be blocked by" and
"can't block" are the same pair predicate here, so one check answers
both directions: a pro-red blocker may not be blocked by a red
attacker is not a thing that happens, but a pro-red *blocker* facing
a red attacker is legal (protection does not stop you blocking), and
the check is written so it does not refuse that.

### 6. Attachments: before the catalogued-Aura early return

The check goes at the top of `attachmentLegalLocked`, ahead of the
`TargetSpecFor` branch that returns early for a catalogued Aura. CR
702.16c-d make an Aura, Equipment or Fortification with the quality
an illegal attachment, and the existing SBA (#939 / #963) then does
the right thing without a new rule: `attachmentSBALocked` sends an
illegal **Aura** to the graveyard (CR 704.5m) and merely **unattaches**
an Equipment or Fortification (CR 704.5n).

The protection test is against the ATTACHMENT, not against its
controller — a red Aura an opponent controls and a red Aura you
control both fall off a pro-red creature.

### 7. Prompts: neither one is new machinery, and one of them waits

**Choose a color (Mother of Runes)** uses the existing
`choose_color` `PendingChoiceKind` and
`QueueColorChoiceThenForEffect` (#742, closed seam). The ability
resolves, the prompt is put to the controller, and the answer feeds
`GrantKeywordUntilEOT{Keywords: []string{"protection from " + colour}}`
— a layer-6 grant, read back by the same one reader. No new kind, no
`internal/legal` case, no restorability census entry.

**Choose a player (True-Name Nemesis, CR 702.16k)** does NOT ship
here, and the reason is a dependency rather than a design gap. The
quality is "the chosen player", which is a stable SEAT REFERENCE
(never a raw UUID in a display string — the badge tooltip renders
the token). The decision, for whoever lands it:

- the answer is stored on the permanent as `Card.ChosenPlayer`
  (`uuid.UUID`), modelled exactly on `Card.ChosenColor` — cleared by
  `entry_tail.go` and `zone.go` on every zone change, carried in the
  snapshot, classified `carried` for clone/undo;
- the token is `protection from the chosen player`, and
  `ProtectionQualities` resolves it against that field, so no raw
  UUID ever reaches a string;
- the prompt is [#929](https://github.com/krakenhavoc/cmd_and_ctrl/issues/929)'s
  `QueueChoosePlayerForEffect`, which is built on the existing
  `option_pick` kind and adds no `PendingChoiceKind` — True-Name
  Nemesis needs its as-enters (CR 614.12) sibling, on
  `ChooseColorAsEnters`'s pattern;
- the matcher compares the source's **controller**, not its
  characteristics, which is why `TargetSource` and `SourceLKI` both
  carry a controller.

Shipping the field with no writer would have put a dead branch in
the one reader this ADR exists to keep small.

**Correction (2026-09-19, #929).** This section said "It lands with
#929." It does not, and the split is worth stating precisely. #929
shipped only the RESOLUTION-time prompt — `QueueChoosePlayerForEffect`
on the existing `option_pick` kind, whose answer lives on
`StackItem.Payload` for the duration of one resolution. True-Name
Nemesis needs the other half: a choice made AS the permanent enters
(CR 614.12) and stored on the permanent for the rest of its life,
which is a different mechanism with its own field, its own snapshot
classification and its own clearing rule. That half, the player
quality in the grammar and the card are **#980**; everything this
section specifies above still stands as its design.

**Amendment (2026-09-19, #980): it shipped, and the design above held.**
Every bullet is as written — `Card.ChosenPlayer` on `ChosenColor`'s
pattern, classified `carried`, the token `protection from the chosen
player`, the matcher on the source's controller. Four things the design
did not say, each of which the build had to decide:

- **Where the seat is resolved.** The token names no player, so a
  quality parsed on its own protects from nobody. `ProtectionQuality`
  gains a `Player uuid.UUID` that the READER fills in
  (`bindProtectionQuality`, called by both `ProtectionQualities` and
  `MatchedProtection`) off the card it was handed. That is what lets all
  four DEBT checks keep their existing `ProtectedFrom(card, chars)`
  call — none of them changed — and it is why no raw UUID reaches a
  display string: `Printed` stays the card's own words and the id sits
  beside it.
- **The prompt is a third door, not a flag.**
  `QueueChoosePlayerAsEntersForEffect` sits beside
  `QueueChoosePlayerForEffect` rather than inside it. They share the
  option list (`seatChoiceOptionsLocked`) and therefore the kind, the
  gate, the enumerator, the wire and the CR 800.4a pruning; what differs
  is only where the answer goes, and folding that into a flag would have
  made one function answer two questions about lifetime.
- **Which prompt shape.** Queued from the `AsEnters` hook, not by
  pausing the CR 614 pipeline — S26's declared simplification
  (`creature_type_choice.go`) carried forward a third time, for its
  original reason: `entryResumable` is set on the land-play branch
  alone, and True-Name Nemesis is a creature.
- **No layer bump.** `ResolveCreatureTypeChoice` and
  `ResolveColorChoice` both bump `layerVersion` because their answers
  are `AppliesTo` inputs to their permanent's statics. A chosen player
  is not one: protection rides the printed token, which has been in
  `Abilities` since the permanent entered, and the reader binds the seat
  at check time rather than baking it into a characteristic.

**§10's "player protection" row is unchanged and is a different rule.**
A player HAVING protection (Teferi's Protection, Leyline of Sanctity)
still has no home; this is a permanent protected FROM a player.

**The grammar's closed set is now five shapes**, and the table in §1
should be read with a fifth row: `protection from the chosen player`,
matched against the source's `Characteristic.Controller`. The
`ProtectionFromChosenPlayer` constant is the only spelling; a card file
that types the phrase by hand and misses mints no token at all.

**Two cards, not one.** True-Name Nemesis is the proof for the quality;
Sawhorn Nemesis reads the same stored field from a damage replacement
and never touches `protection.go`, which is the evidence that
`Card.ChosenPlayer` is general as-enters machinery rather than a
protection back door.

### 8. Bots read the reader

`aiseat/heuristic`'s `hasKeyword` prefix-matches `"protection"` and
throws the quality away. The block planner now calls
`game.ProtectedFrom` through the enumerator's legality check
(`internal/legal` reaches `BlockPairRefusalLocked` already), and
`score.go` keeps its prefix match for VALUING a creature — which is
the one place where "has some protection" is the honest question.

The enumerator needs no protection code of its own. It reaches
targeting through `LegalTargetsForEffect` and blocking through
`BlockPairRefusalLocked`, so it inherited both filters the way ADR
0038 §5 describes; what it needed was the source, and it now passes
one.

### 9. Undo, clone and the snapshot

- **Tokens are characteristics.** `Characteristic.Abilities` is
  already snapshotted and cloned; a printed protection rides
  `Card.Keywords` and a granted one rides the layer engine, both of
  which round-trip today. Nothing new.
- **`SourceLKI` rides the damage event**, which is engine plumbing
  inside one `applyReplacementsLocked` call and never outlives it.
  A CR 616 pause stores the event in the pending-choice frame, which
  is the same road `damageTail` already travels.
- **`Preemptive`** is a field on a `ReplacementEffect` value, and
  built-ins are shared by reference across a clone
  (`clone.go`) exactly as they were.
- **`Card.ChosenPlayer`** is `carried` (§7, landed with #980), and
  cleared at BOTH CR 400.7 sites — `zone.go`'s battlefield exit and
  `resetAsNewObjectLocked`'s new-object reset — because a blinked
  permanent reaches the second without passing the first. It is not a
  copiable value (CR 707.2) and needs no rule to say so:
  `CopiableValuesOf` projects printed characteristics and never looks
  at it, exactly as ADR 0071 arranged for `ClassLevel`.

### 10. Out of scope, stated

- **Player protection and player hexproof** — Teferi's Protection,
  Leyline of Sanctity, The One Ring. `Player` carries no ability
  slice and `CanBeTargetedBy` is never reached for a player ref.
  Unchanged from ADR 0038.
- **CR 615.12 "can't be prevented"** — §4.
- **CR 616.1 ordering against protection** — §4, a declared
  simplification.
- **Cross-layer qualities** — "protection from the colour of your
  choice" is a colour token chosen once and granted (§7); "protection
  from each colour among permanents you control" and friends would
  need the token to be recomputed per-check, which the grammar
  deliberately cannot express.
- **CR 702.16m (two instances)** — two identical tokens are one
  quality for every check here; the client dedupes its badge row
  rather than rendering two.

## Consequences

**Four catalog cards lose their caveat.** Baneslayer Angel (a subtype
quality — a changeling source counts as a Demon and as a Dragon
through `AllCreatureTypes`), Sword of Fire and Ice, Sword of Feast
and Famine and Animar, Soul of Elements all go from
`CompletenessCaveats` to `CompletenessFull` on this axis. Mother of
Runes joins the catalog.

**The protection seam row closes** in `docs/engine-seams.md`; the
choose-a-player row (True-Name Nemesis, Gluntch, Skullwinder, Victory
Chimes) stays open and is now protection's only remaining dependency.

**Uncatalogued printed protection now works**, for any quality the
grammar parses — which is every colour, every card type and every
creature type. That is a behaviour change for imported decks and it
is the point: a Progenitus or a Sword of Light and Shadow in
somebody's deck stops being a lie. Cards whose quality the grammar
refuses keep the ADR 0037 unimplemented badge, so nobody is told a
card works when it does not.

**The `TargetSource` signature change is the maintenance cost.**
Every future targeting path has to name a source, and a path that
genuinely has none has to say so with `SourceChooser` — which reads
as a claim ("this is not targeting") rather than as an omission.
That asymmetry is deliberate and is the same one ADR 0038 §4 chose
for the cost-payment split.
