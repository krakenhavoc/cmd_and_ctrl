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

- ~~**Player protection and player hexproof**~~ — Teferi's Protection,
  Leyline of Sanctity, The One Ring. `Player` carries no ability
  slice and `CanBeTargetedBy` is never reached for a player ref.
  Unchanged from ADR 0038. **Retired 2026-09-22 (#1197)** — see the
  amendment at the end of this ADR: `Player.Statics` is the slice,
  and the three choke points that each returned early for a player
  ref now answer.
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

## Amendment (2026-09-22, #1197): the RECEIVER may be a player (CR 702.16i, CR 702.11d)

§10 listed "player protection and player hexproof" as out of scope,
with one reason: `Player` carries no ability slice, so "you have
protection from everything" has nowhere to live and no consumer.
This amendment gives it both. Nothing above changes — the grammar,
the token, the quality kinds and the matcher are the ones already
shipped. What is new is a second kind of thing that can HAVE a
quality.

Note the two axes, because they are easy to conflate and this ADR
now has all four corners:

|                 | protection ON an object | protection ON a player |
| --------------- | ----------------------- | ---------------------- |
| **FROM** a characteristic | #662 (§1–§6) | this amendment |
| **FROM** a player | #980 (§7 amendment) | reachable, uncatalogued |

### A1. Where a player's abilities live: derived plus granted, never written

`Player` gets one new field, `Statics []PlayerStatic`, and the
engine gets one new reader. The field is modelled on
`Player.CastPermissions` (ADR 0066), which is the closest existing
thing: per-player, plain data with no closures, carrying a
`Duration`, cloned by value and mirrored into the snapshot rather
than rebuilt.

A player's abilities come from two places, and the split is the same
one `land_drops.go` and `CatalogNoMaxHandSize` (#338) already make:

- **DERIVED** — a permanent on the battlefield whose printed static
  says "you have hexproof" (Leyline of Sanctity, Aegis of the Gods).
  Declared as `effects.Spec.PlayerKeywords` and read through the
  `game.CatalogPlayerKeywords` hook on every query. Nothing is
  written to the player, so two Leylines compose and one of them
  leaving cannot revoke the other's grant — the argument
  `CatalogNoMaxHandSize` spells out at length, verbatim.

- **GRANTED** — a resolved spell or triggered ability that says "you
  gain protection from everything until your next turn" (Teferi's
  Protection, The One Ring). This one HAS to be stored: the source
  is a spell that is in a graveyard a moment later, which is the
  same reason `ScopedStatic` exists (scoped_statics.go). It carries
  a `Duration` and is swept through `durationExpiredLocked` — the
  one function ADR 0063 says decides when any continuous effect in
  this game is over.

**Why not a `ScopedStatic`.** The registry next door looks like the
obvious home and is the wrong one. A `ScopedStatic` is adapted into
a `ContinuousEffect` and applied by the CR 613 layer pass, whose
`Apply` signature is `(*Characteristic, *Card)` — a characteristic
of an OBJECT. A player has no `Characteristic` and no layer, which
is exactly the argument `Spec.NoMaxHandSize` and `Spec.CostModifiers`
already make for staying out of the layer engine. So a
`PlayerStatic` is a THIRD thing beside the two registries: a token,
an attribution, and a duration.

```go
type PlayerStatic struct {
    Keyword  string    // an engine token: "hexproof", "protection from everything"
    Source   uuid.UUID // attribution, for the log and the badge
    Label    string
    Duration Duration
}
```

`Keyword` is a token in the same closed vocabulary `Card` abilities
use, parsed by the same `ParseProtectionQuality`. That is the whole
point of putting it here rather than inventing a player-side enum:
"protection from everything" means one thing in this engine and one
file parses it.

### A2. One reader, three consumers — the same shape §1 chose for objects

`playerAbilityTokensLocked(p)` is the player-side twin of
`forEachAbilityToken`: the derived battlefield grants, then the
stored ones whose duration has not run out. The expiry test is in
the READER as well as in the sweep, for the reason
`CastPermissionActiveForEffect` gives — the sweep is hygiene, run at
known moments; the reader is the truth, and must be right between
them.

Three consumers, one line each, all of them the choke point that
already exists:

| rule | where | CR |
| --- | --- | --- |
| targeting | `specMatchesLocked` seat walk + `specMatchLocked`'s `case TargetPlayer` (targets.go) | 702.11d, 702.16i |
| damage | `protectionPreventsDamageLocked`'s non-permanent branch (builtin_replacements.go) | 702.16e |
| attachment | `attachmentLegalLocked`'s `TargetPlayer` host (attach.go) | 702.16c |

None of the three is a new pass. Each is the branch the existing
function already had and returned early from:
`protectionPreventsDamageLocked` said "not a permanent: a player…
player protection has no home yet"; `attachmentLegalLocked` tested
only `AttachedTo.Kind == TargetCard`; `CanBeTargetedBy`'s doc
comment said "players are not covered… `TargetPlayer` refs pass this
gate by not reaching it". All three comments were an IOU and this
amendment pays them.

**The enumerator and the view follow by construction.** The bot's
move list (`internal/legal`) and the client's `legal_targets` both
read `legalTargetsLocked`, which is `specMatchesLocked` with
`targeting=true`. Neither learns the rule; they cannot disagree with
it. That is the property §2 bought with the `TargetSource` refactor
and it pays out again here with no new call site.

**Blocking and equipping are not relevant.** CR 702.16f is about a
creature being blocked; CR 301.5c's Equipment attaches only to a
creature. A player can be enchanted (an "enchant player" Aura —
Curse of Opulence) and that is the only attachment half there is.

### A3. Hexproof on a player is the same asymmetry it is on an object

CR 702.11d is CR 702.11b with "player" for "permanent": can't be the
target of spells or abilities your OPPONENTS control. So the test is
`src.Controller != p.ID`, and a hexproof player may still target
themselves — which is load-bearing, not a nicety. Leyline of
Sanctity must not stop you from casting your own Sylvan Library, and
a player who could not target themselves would be unable to pay a
cost, take a draw trigger's downside, or aim their own removal-plus-
gain.

Shroud on a player is not in the grammar. No card prints it and
inventing the token would be a rule with no card behind it, which is
what `CanonicalKeywords` refuses a bare "protection" for.

### A4. What this does NOT build, stated the way §10 states things

- **Phasing (CR 702.26), #1199** — not modelled anywhere in this
  engine (`activation_tally.go` says so and has since S38). Teferi's
  Protection's "all permanents you control phase out" ships as a
  caveat, not as a half-built phase. It is the larger half of that
  card and it is a seam of its own.
- **"Your life total can't change", #1200** — a replacement effect with a
  duration longer than end of turn. `TurnScopedReplacements` has no
  duration field at all, which ADR 0063 Decision 8 states as
  deliberate ("no card needs a longer-lived replacement yet"). This
  is the first card that does, and it is not this seam.
- **Protection on a player FROM a player** (the fourth corner of
  the table above). Reachable — the quality kind, the matcher and
  the reader are all shared — but no catalogued card prints it, so
  nothing exercises it and it is claimed nowhere.
- **A player losing an ability.** CR 613 layer 6 can strip a
  permanent's abilities; there is no equivalent for a player and no
  card asks for one.
- **The client badge, #1201.** The wire carries a player's tokens
  (`PlayerView.keywords`); rendering a chip beside the life total is
  a client-only change.

### A5. Consequences

**The seam row closes.** "Player protection and player hexproof
(CR 702.16, CR 702.11)" in `docs/engine-seams.md`, 4 cards.

**The One Ring loses its caveat's first half.** The card is already
in the catalog and its caveat names this exact gap — "casting The
One Ring doesn't shield you". It now does.

**Three cards join the catalog**: Leyline of Sanctity, Aegis of the
Gods (both `CompletenessCaveats` only for their opening-hand /
nothing-else clauses) and Teferi's Protection (`CompletenessCaveats`
for phasing and the life lock).

**Uncatalogued printed player hexproof still does nothing.** The
derived half reads `Spec.PlayerKeywords`, a catalog declaration, not
Scryfall's keyword array — because "you have hexproof" is oracle
TEXT and not a keyword-ability line, so there is nothing for the
deck importer to stamp. That is the same posture §1 takes toward a
quality the grammar cannot parse: err weaker, and let the card carry
the ADR 0037 unimplemented badge until somebody writes it.

## Amendment (2026-09-24, #1417): a source that left BEFORE its damage event is read as it last existed

The third bullet of §3 says: "a source whose characteristics changed on
the way out … CR 608.2h says use the last existing characteristics".
The code kept that promise only for a source that left *after* its
damage event was created, such as the #694 pause, the CR 510.1c frame,
or an SBA sweep mid-step. A source that had *already* left, such as a
"when this dies" trigger or a creature killed in response to its own
ability, was snapshotted from its current zone. That meant its
graveyard card, which is a new object with printed characteristics
(CR 400.7). So a creature painted red that died dealt colourless damage
through protection from red.

**Decision.** The non-combat tail (`effectDamageTailLocked`) now takes
`sourceLKI` from `Game.lastKnownPermanents`, the per-object record
that ADR 0056's 2026-09-24 amendment already reads a departed source's
lifelink and deathtouch from. It uses the same object rules: by
`ObjectRef` when the caller named the object, and by "last object,
card not moved since" for a bare instance ID. A spell on the stack and
a live permanent keep the current-zone read, so a Lightning Bolt's
colour is still the one on the stack. Details are in
[ADR 0056 Decision 12](0056-infect-wither-toxic.md).

The catalog's "a red source you control" damage replacements (Torbran,
Ojer Axonil, Mechanized Warfare) now read the same `SourceLKI` through
`effects.damageSourceCharacteristics`, so protection and those cards
agree on a source's colour.

~~**Still open.**~~ **Closed by #1429**, see the next amendment. §2's DECLARED LIMITATION is narrower than it reads, and
it is tracked as
[#1429](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1429). For
an ability whose source has left, `stackItemSourceLocked` does not
return a source-less value. It returns the source's **graveyard card**.
So the CR 608.2b target re-check tests protection against the wrong
object. The damage half is now correct either way, but a target can
still be declared illegal (the ability fizzles) when the source that
last existed would have been allowed to target it.

## Amendment (2026-09-24, #1429): the CR 608.2b re-check reads a departed ability source as it last existed

Issue [#1429](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1429),
the "Still open" paragraph of the #1417 amendment above. §2's table
says the resolution re-check's source is "the stack item's card, looked
up live and falling back to `StackItem.SourceLKI`". No such field was
ever added. The code looked the source up with `LookupCardForEffect`,
which finds the card in whatever zone holds it now. For an ability
whose source had left the battlefield, that was the source's graveyard
card: a new object with printed characteristics (CR 400.7). The comment
on `stackItemSourceLocked` said it returned a source-less value, which
was also wrong.

So protection (CR 702.16b), and every other restriction that reads the
source's qualities, was judged against the wrong object at resolution:

- a source painted red that died could keep a pro-red target, and its
  ability resolved (#1417 then prevented the damage, so only the fizzle
  was missing);
- a printed-red source made colourless that died lost a pro-red target
  it was allowed to have, and its ability fizzled. That direction is
  weaker than printed and visible on the board.

**Decision.** `stackItemSourceLocked` (`server/internal/game/targets.go`)
returns `SourceSnapshot(item.Controller, …)` of the source's
`Game.lastKnownPermanents` record when the item is an ABILITY and
`departedAbilitySourceLocked` finds a record. CR 608.2h is the rule:
information about an object that has left the zone it was expected in
is its last-known information. Which object an item names is
`StackItem.SourceObject`, the exact source object that #1418 (ADR 0018
Decisions 14-16) stamps on every ability item, read through
`SourceObjectForEffect`:

- **That object is still on the battlefield**: it is read live. This
  check comes first, because an exit that a replacement stopped
  part-way can leave a record at a live epoch.
- **That object has left**: its record. This finds the right object
  even when the card is back on the battlefield as a new one. An
  activation's stamp is taken *before* its costs, so a source
  sacrificed to its own ability is still the object its record
  describes, even if the card has since come back.
- **That object was never a permanent** (a trigger from a card in a
  graveyard, an ability activated from a hand): there is no record, so
  the card is read where it is. That card *is* the source. This matters
  when the card was also a permanent earlier in the turn: it is judged
  as the graveyard card, not as the permanent it used to be.
- **An item with no stamp**, restored from a snapshot written before
  #1418, keeps the instance-ID rule from ADR 0056 Decision 10: the
  card's last battlefield record, only while its current epoch is
  exactly one past it. That rule never answers for a card on the
  battlefield, so a live source is read live.

The snapshot's controller comes from the record when the characteristic
has none, the same rule `SourceCharacteristics` uses. That is now one
helper, `lastKnownSourceCharacteristics`, shared with the damage tail
(#1417).

What does not change:

- **A spell** is its own source and is read on the stack. It never
  consults the record, because a spell is not a permanent and a card that
  was one earlier in the turn is a different object now. A spell item
  carries no `SourceObject`, so the guard in `stackItemSourceLocked` is
  what keeps it away from the unstamped fallback.
- **An ability with no matching record** keeps the lookup. That covers a
  channel or graveyard ability whose card was never on the battlefield,
  and a source that left by a route that writes no record (CR 800.4a).
- **Retargeting** (`retargetSourceLocked`) reads the same function, so a
  Deflecting Swat pointed at an ability whose source died is judged by
  the same object.

The §2 table row should read: "resolution re-check
(`TargetStillLegalForEffect`, CR 608.2b): a spell's own card on the
stack; an ability's source permanent live, or its last-known record
once it has left."

**Proof.** No catalog caveat named this gap. The proof uses #1417's
line through real cards: Murderous Redcap's enter trigger targets a
creature, Cerulean Wisps turns the Redcap blue in response, the Redcap
dies, and then Mother of Runes gives the creature protection.

| Mother of Runes names | Before | After |
|---|---|---|
| red | the trigger fizzles (graveyard card is red) | resolves, 2 damage |
| blue | the trigger resolves, damage prevented | fizzles (CR 608.2b) |

**Also covered, because #1418 landed first.** Two cases that the
instance-ID rule alone gets wrong are right, and each has a test:

- a graveyard trigger from a card that died earlier this turn is judged
  as the graveyard card, not as the permanent it was;
- an activation whose cost sacrificed its source, and whose card has
  since come back, is judged as the sacrificed object, not as the new
  one.

**Still open.**

- **Unstamped items** from a pre-#1418 snapshot keep the instance-ID
  rule, and with it both of the imprecisions above. That is a
  compatibility path, not a new gap.
- ~~**The re-target prompt for a COPY of an ability**
  (`abilitySourceCardLocked`, `server/internal/game/ability_copy.go`)
  still judges its picks against the source's current-zone card. The
  copy frame carries a `Card`, not a `Characteristic`. Its comment
  now says so. Filed as [#1449](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1449).~~
  **Closed by #1449**, the amendment below.
- **Hexproof from a quality** is not implemented as a keyword. When it
  is, it reads `TargetSource.Characteristics()` like protection does,
  so it will get the last-known source for free.

## Amendment (2026-09-24, #1449): an ability copy's new targets are judged against the departed source as it last existed

Issue [#1449](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1449),
the copy half the #1429 amendment above left open. When an ability is
copied and the copy's controller may choose new targets (CR 707.10c,
which is CR 115.7c by reference), three places judge the new targets:

1. the **offer**, which computes the prompt's legal set
   (`offerCopyTargetsLocked`, `server/internal/game/spell_copy.go`);
2. the **#809 refresh**, which re-reads an open prompt's legal set at
   every state check (`refreshTargetChoicesLocked`,
   `server/internal/game/trigger_target_timing.go`);
3. the **answer**, the CR 115.7 check on the chosen targets
   (`resolveCopyTargetsLocked`, `spell_copy.go`).

All three judged against the copy frame's `src`. For an ability copy
that is `abilitySourceCardLocked(item)`: the source looked up in its
current zone. Once the source had left the battlefield, that was its
graveyard card, a new object with printed characteristics (CR 400.7),
not the source as it last existed (CR 608.2h). So a copy of an ability
whose source was painted red and then died could be pointed at a
pro-red creature, and a copy whose printed-red source was made
colourless before it died could not. The original's CR 608.2b re-check
had been right since #1429, so the two disagreed about the same source.

**Decision.** One function, `copyTargetSourceLocked(cf *copyFrame)`
(`spell_copy.go`), is the `TargetSource` for all three places. For an
ABILITY copy whose source `departedAbilitySourceLocked` finds a record
for, it returns `SourceSnapshot(controller,
lastKnownSourceCharacteristics(rec))`. That is the read
`stackItemSourceLocked` makes for the original, from the same
`StackItem.SourceObject` (#1418): the frame keeps a value copy of the
original's item, and `createAbilityCopyLocked` gives the copy that same
`SourceObject`, so the prompt, the answer and the copy's own CR 608.2b
re-check at resolution all name the same object. Otherwise it returns a
snapshot of the frame's `src`, as before. A new field on `copyFrame`
(the issue's suggested `srcChars`) was not needed.

The read happens at each check rather than being frozen at the offer.
So a source that leaves while the prompt is open is judged as it last
existed when the prompt is refreshed and answered.

What does not change:

- **A spell copy** is judged against the copied spell. The guard is
  `cf.item.Kind != StackItemSpell`, the same one `stackItemSourceLocked`
  has, for the same reason: a spell's source is the spell, and a spell
  item carries no `SourceObject`, so without the guard it would reach
  `departedAbilitySourceLocked`'s unstamped instance-ID fallback and
  could be read as the permanent its card was earlier in the turn.
- **A live source**, and an ability whose source has no record, keep
  the snapshot of `src`. At the offer that replaces
  `SourceObject(controller, &src)`. The two are the same for legality:
  `TargetSource.Characteristics()` answers `SourceCharacteristics(Object)`
  for the one and the snapshot for the other, and nothing else reads
  `TargetSource.Object`.

**Fixed along the way.** When the #809 refresh finds an open copy prompt
with nothing left to offer, it withdraws the prompt and creates the copy
with the original's targets. It called `createSpellCopyLocked` for every
copy frame, so an ability copy's frame put a SPELL copy of the source
permanent's card on the stack. The refresh now calls `createCopyLocked`,
the one branch on `item.Kind` that the offer and the answer already use.
The LKI read above makes this path easier to reach, because the refresh
can now narrow a set that the old read left open.

**Proof.** No catalog caveat named this gap. The proof is #1429's line
through real cards, with a real copy effect: Murderous Redcap's enter
trigger targets a creature, Cerulean Wisps turns the Redcap blue in
response, the Redcap dies (black-red again in the graveyard), and
Strionic Resonator copies the trigger.

| Copy's new target | Before (graveyard card, black-red) | After (last known, blue) |
|---|---|---|
| a creature with protection from red | not offered, refused | offered, 2 damage |
| a creature with protection from blue | offered | not offered, refused |

**Still open.** Nothing on this path. Hexproof from a quality is still
not a keyword (the #1429 amendment's last bullet).
