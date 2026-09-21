# ADR 0080 — Attack taxes: "creatures can't attack you unless their controller pays {N}" (CR 508.1a)

**Status:** Accepted · 2026-09-21 · S37 — Combat correctness: damage steps, removal from combat, block legality
**Issues:** [#1063](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1063) (the seam), cards from [#294](https://github.com/krakenhavoc/cmd_and_ctrl/issues/294) (Propaganda, Ghostly Prison, Windborn Muse)
**Numbering:** swept with the AGENTS.md §4 loop on 2026-09-21 — `git fetch origin`,
then `git log --all --diff-filter=A --name-only -- 'docs/decisions/*.md'` over every
fetched ref, which reaches commits on remote heads a per-branch `ls-tree` also
reaches and additionally catches an ADR added and later renamed. The highest number
present anywhere was **0079** (`0079-transforming-a-permanent.md`); `0080` was free,
and the sweep was repeated immediately before the push. `0005`, `0024`, `0029` and
`0030` stay permanently unused.

**Related:** [ADR 0045](0045-combat-restrictions.md) (the restriction vocabulary,
which names this shape as deliberately outside it), [ADR 0073](0073-optional-additional-costs-and-the-cast-gate.md)
(the announce-time cost posture this copies), [ADR 0048](0048-cost-modification.md)
(the `CostQuery` shape the query mirrors), [ADR 0071](0071-designations-that-switch-abilities-on.md)
(`ActiveWhen`), [ADR 0033](0033-ai-bot-seat.md) (the enumerator's promise).

## Context

ADR 0045 built the restriction vocabulary as five absolute bits —
`CantAttack`, `CantBlock`, `CantBeBlocked`, `CantActivate`,
`CantActivateMana` — and said in its own words what it could not say:

> **"Can't attack unless its controller pays {2}"** — Propaganda,
> Ghostly Prison, Norn's Annex. This is a *cost* to attack, not a
> prohibition, and a bit cannot carry a cost. It wants an attack-cost
> pipeline shaped like the cast-cost one and keyed on the **defending
> player** rather than on the attacker.

`restrictions.go:57-64` repeats it, and `attack_target.go:27-33` again.
Three places in the tree describe the missing mechanism and none of
them builds it. #1063 is the tracker that row was missing.

The two cards are batch 01's, at EDHREC rank ~114–162 — two of the most
played enchantments in the format — and neither can ship weaker.
Shipping Propaganda with the clause omitted is strictly *stronger* for
every attacker at the table, which is the #259 direction the roadmap
refuses.

### What the declaration path looked like before this

`DeclareAttackers` (`mutations.go:6075`) takes the whole attacking set
in one call, validates each entry, taps it (CR 508.1f) and stages it;
`runStateChecksLocked` then locks the declaration in through
`commitAttackDeclarationLocked`, which is the ONE place an attack is
announced whichever verb staged it (#859). The per-creature
`DeclareAttacker` (`mutations.go:5947`) stages one attacker and is what
the legal-move enumerator emits, one move per (attacker, target) pair
(`legal/combat.go:64-81`).

Nothing on that path charges anything. The only thing a declaration
"costs" is the CR 508.1f tap.

## Decisions

### 1. `AttackTax` is a per-permanent static on the DEFENDER's side

A new file, `server/internal/game/attack_tax.go`:

```go
type AttackTax struct {
    Label      string                        // "Propaganda — {2} for each creature attacking you"
    AppliesTo  func(q AttackTaxQuery) bool   // nil = every creature attacking me
    ManaCost   func(q AttackTaxQuery) string // the mana half, PER attacking creature
    ActiveWhen Designation
}

type AttackTaxQuery struct {
    Game       *Game
    Source     Card      // the permanent whose clause this is
    Defender   uuid.UUID // the player being attacked — the "you"
    Attacker   Card      // the creature being declared
    Controller uuid.UUID // that creature's controller — "their controller"
}
```

Declared as a struct of hooks rather than an interface for the reason
`StaticAbility`, `CostModifier` and `ReplacementEffect` are: a card
file writes a literal.

**The "you" is structural, not a predicate.** A tax protects its
source's controller and nobody else. Every printed card in the family
says "creatures can't attack **you**", and making it a predicate would
let one card file typo its way into taxing the whole table. `AppliesTo`
narrows *which attacking creatures* are taxed, which is the axis real
cards vary on (a hypothetical "for each creature with flying"), and
nil — the whole family today — means all of them. A card that taxes
attacks on somebody else does not exist; when one does it gets a field,
not a silent reinterpretation of this one.

**`ManaCost` returns a string, not a `ParsedCost`.** Three reasons, all
load-bearing:

- The cost is per attacking creature and a declaration sums several,
  so the price is built by CONCATENATION — `{2}` three times is
  `{2}{2}{2}`, which `ParseCost` reads as six generic, no arithmetic
  and no normaliser to get wrong.
- A count-scaled tax is a string the card renders once
  (`"{" + itoa(n) + "}"`), which is Sphere of Safety and Collective
  Restraint without a second mechanism.
- Norn's Annex's `{W/P}` and any hybrid are already what the parser
  eats, so the Phyrexian half rides `strikePhyrexianLifeLocked`
  unchanged when that card is written.

Nil `ManaCost`, or one returning `""`, is a tax of nothing rather than
an error: a clause that names no price simply does not apply, which is
`CostModifier.Amount`'s rule.

**Collection mirrors every other catalog-contributed static.**
`CatalogAttackTaxes` is a hook wired in `carddef.go` beside
`CatalogCostModifiers`; `AttackTaxesForCard(c Card) []AttackTax` reads
it through `CatalogAbilityKey` so a permanent under a CR 613.1f
ability-removing effect stops taxing, and applies the `ActiveWhen`
designation gate in that accessor **and nowhere else** — which is what
makes the engine and the enumerator price a declaration identically,
the argument `CostModifier.ActiveWhen` already makes for casts.

### 2. One pricer, three callers

```go
func (g *Game) PriceAttackDeclaration(decls []AttackDeclaration) AttackTaxPrice          // takes RLock
func (g *Game) PriceAttackDeclarationForEffect(decls []AttackDeclaration) AttackTaxPrice // caller holds g.mu
func (g *Game) priceAttackDeclarationLocked(decls []AttackDeclaration) AttackTaxPrice    // the body
```

```go
type AttackTaxPrice struct {
    Cost  string          // concatenated mana, "" when the declaration is free
    Total ParsedCost      // Cost parsed, for CanPayFor / the tapper
    Lines []AttackTaxLine // one per (source, attacking creature), for the log and the client
}
```

The body walks the declarations; for each, resolves the defending
player with `defendingPlayerForAttackLocked` (so an attack on a
planeswalker or a battle is taxed by its CONTROLLER's / PROTECTOR's
Propaganda, CR 508.1a), then walks the battlefield for that player's
permanents and asks each of their taxes. `PriceCast` is the model and
the name is deliberately its sibling.

**It is `PriceCast`'s sibling, not `PriceCast`.** `PriceCast` is for
casts: it materialises a face, swaps in an alternative cost, adds the
commander tax and runs the CR 601.2f modifier pass, none of which a
declaration has. What the two genuinely share is the MANA MACHINERY
below them — `ParseCost`, `ManaPool.CanPayFor`, `autoTapLocked`,
`materializePlanLocked` — and that was already shared, by
`payAbilityManaCostLocked`, before this ADR (`special_action.go:253`
pays foretell through it). So nothing was extracted: the declaration
became the fourth caller of the payer that already existed.

### 3. The cost is charged at CR 508.1a, on the declaration, before the commit

Inside the declaration verbs, after the eligible set is settled and
**before** any creature is staged:

```
eligibility (CR 508.1c/d)  →  price the eligible set  →  PAY  →  stage  →  lock in
```

Three properties, each chosen rather than fallen into:

- **The price is over the ELIGIBLE set, not the submitted one.** A bulk
  submission silently skips entries that are tapped, summoning-sick or
  under `CantAttack` (the contract `DeclareAttackers` has had since
  #318). Pricing the submitted set would charge for attacks that never
  happen.
- **All or nothing (CR 508.1a).** If the eligible set's tax cannot be
  paid in full, NOTHING is declared and the verb returns
  `*AttackTaxUnpaidError`. A partially-paid declaration is not a legal
  declaration; the player's move is to submit a smaller set. This is
  the one place the bulk verb is not "skip what doesn't fit" — and it
  has to be, because which attacks to drop is the player's choice and
  the engine must not make it for them.
- **Payment happens before staging, so a refusal leaves the board
  untouched.** No creature is tapped, no `AttackingTarget` is written,
  `commitAttackDeclarationLocked` never runs, and no `EventAttack`
  fires. The verbs already return before staging on every other
  refusal; this joins them.

**No `pay_unless` prompt.** `PendingChoicePayUnless` exists and is the
wrong tool: it is a prompt raised *after* the thing happened, answered
later, and `choice_gate.go:144` deliberately does not let it block the
table. An attack tax is announce-time — CR 508.1a puts it inside the
turn-based action, alongside the CR 508.1f tap — so it is paid by the
action that declares the attack, exactly as a cast's additional costs
are paid by the action that casts the spell (ADR 0073 §4). The
declaration verbs also cannot yield: they are synchronous under one
write-lock acquisition, and teaching them to pause would mean a new
declaration-scoped resume frame for a payment that has no reason to
wait.

**What "the attacking player names what they pay for" means here.**
For a cast, ADR 0073 §2 needs an explicit `[]int` because paying a
kicker is a free choice that changes the spell. An attack tax is not a
choice: the price is a pure function of the declaration. So the
declaration IS the naming — the player names the attacks, the engine
prices exactly those, and the client shows the total before the action
is sent. What the wire adds is the *payment posture*, the same pair
every other paying action carries.

### 4. The payment posture is the one every other paying action uses

Both verbs gain a params twin rather than a changed signature —
`DeclareAttacker` has 190 call sites:

```go
type DeclareAttackersParams struct {
    AutoTap       bool
    LockedSources []uuid.UUID
    PhyrexianLife int
}

func (g *Game) DeclareAttackerWith(attackerID, targetID uuid.UUID, params DeclareAttackersParams) error
func (g *Game) DeclareAttackersWith(decls []AttackDeclaration, params DeclareAttackersParams) ([]uuid.UUID, error)
```

`DeclareAttacker` and `DeclareAttackers` are now one-line wrappers
passing the zero value, so every existing caller is unchanged and a
table with no tax on it is byte-identical to what it was.

**There is no permissive mode.** `payAbilityManaCostLocked`'s
"`!Strict && !AutoTap` → warn and mark it paid on paper" branch is for
the sandbox's hand-tracked mana; an attack tax waived on paper is
Propaganda as a blank, which is the outcome this ADR exists to prevent.
The declaration always passes `Strict: true`, which also makes the Go
zero value of `DeclareAttackersParams` the STRICT posture — pay from
the pool, refuse if short — rather than the lax one.

**The spend context is the zero `ManaSpendContext{}`**, the decision
`special_action.go:249-252` and `payCostLocked` already document: a
declaration is neither a cast nor an activation, so mana that may only
be spent to cast creature spells cannot fund it.

Mana triggers fire for free, because `payAbilityManaCostLocked` taps
through `materializePlanLocked`, which is one of the three
`fireManaTriggersLocked` sites (ADR 0074). Nothing about the trigger
path is duplicated here, and that is the whole reason the declaration
goes through the shared payer rather than tapping for itself.

### 5. The enumerator offers only affordable attacks, and says the price

`legal/combat.go`'s attack arm prices each candidate (attacker, target)
pair through `PriceAttackDeclarationForEffect` — the same function the
engine charges — and:

- **drops the move when the seat cannot afford it**, pool first then
  `AutoTapForCostForEffectExcluding`, which is exactly
  `enumerator.canPay`'s test for a cast (`legal/cast.go:1031`). This is
  the #544 rule: a bot is never offered a move the engine refuses.
- **stamps `MoveCost.Mana`**, a new string field beside `Life`,
  `Loyalty` and `Counters`, so a policy that may not import
  `internal/game` (ADR 0033 §3) can price the move from the wire
  payload alone. Without it a bot reading only `Params` prices an
  attack under Ghostly Prison exactly like a free one.
- **sets `auto_tap` on the move's params** when the pool alone is
  short but the tapper can cover it, which is what `legal/cast.go:936`
  already does for a cast. The `Move` doc's promise — `Params` is
  EXACTLY the payload that performs the move — is what makes this
  necessary rather than optional.

**The enumeration stays per-creature.** ADR 0033 §1's reason (attack
subsets across four opponents are not enumerable) is untouched, and it
composes with the tax correctly *because* the tax is charged per
declaration verb call: three separate `declare_attacker` actions under
Propaganda pay {2} three times, which is the same {6} one bulk call
pays. After each one the seat's mana has shrunk and the next
enumeration prices the next attack against what is left. A bot cannot
strand itself mid-swing, because a declaration it cannot pay for is
never offered.

### 6. The heuristic subtracts the tax

`heuristic/combat.go`'s `attackValue` gains one term: the move's
`MoveCost.Mana` converted to a generic count and multiplied by a new
`Config.ManaValue` weight, subtracted from the attack's value. A
bot under Propaganda with two lands and a 1/1 now passes instead of
spending its whole turn's mana on two damage; with a lethal push it
still swings, because `LethalBonus` dwarfs the term.

Deliberately a flat generic count rather than a colour-aware price: the
policy reads the wire and the wire carries a cost string, and a
colour-weighted model of "what else could this mana have bought" is the
fuel pricer's job (`heuristic/fuel.go`), not a combat term's.

### 7. What the wire grows

- `declare_attacker` and `declare_attackers` params gain `auto_tap`,
  `locked_sources` and `phyrexian_life` — the same trio
  `cast_spell`, `activate_ability` and `special_action` carry.
- `MoveCost.mana`, a string, omitempty.
- `AttackTargetView.tax`, a string, omitempty: the per-creature price
  of attacking that target, so the client can label the control
  without re-deriving a rule (#429's line: nothing in the client
  re-derives who can attack, and nothing here re-derives what it
  costs).
- One error code, `attack_tax_unpaid`, with the price and what was
  missing on the frame.

The client change is deliberately minimal for this PR: the attack
controls show the tax, the confirm sends `auto_tap: true`, and a
refusal renders the code's message. A richer picker — choosing WHICH
attacks to pay for when the seat can afford some of them, and a
lock-a-land preview like the cast modal's — is filed as a follow-up,
because the engine surface it would need (a per-attack breakdown the
client can subset) is already shipped in `AttackTaxPrice.Lines`.

### 8. Out of scope, stated

- **Count limits** — Silent Arbiter's "no more than one creature can
  attack each combat" and Crawlspace's "no more than two creatures can
  attack you each combat". ADR 0045's addendum Decision 18 left them
  out and this does not pick them up: a count is a property of the
  whole declaration, not a price on it, and it belongs beside
  `blockerBoundsLocked` as a set-shaped predicate. They are a
  different seam and keep their own row in `docs/engine-seams.md`. The
  enumerator problem is also genuinely harder for them — a per-move
  enumerator has to reason about a set — and nothing in this ADR makes
  it easier or harder.
- **A non-mana tax.** No printed attack tax charges anything but mana
  (Norn's Annex's `{W/P}` is mana with a life alternative, which the
  Phyrexian machinery already handles). `AttackTax` carries the mana
  component and `AttackTaxPrice.Lines` is already per-source, so a
  `Sacrifice` or `PayLife` sibling field is an addition to the struct
  and one more branch in the payer — the shape `AdditionalCost` has
  with its three components. It is not a redesign, and it is not built
  for a card nobody is waiting on.
- **A tax with a DURATION.** War Tax ("{X}, {T}: Creatures can't attack
  you this turn unless their controller pays {X} for each attacking
  creature") is an activated ability creating a turn-scoped tax, and
  the defender chooses whether to activate it at all. The collection in
  §1 reads catalog statics on the battlefield; a duration-carrying
  registry is the SECOND source it takes without changing
  `AttackTaxesForCard`'s signature — exactly the extension point ADR
  0073 §8 reserved for `CastRestriction`. Not built, because #755's
  scoped registry is for `StaticAbility` and teaching it a second
  payload type is its own decision.
- **Goad, and "attacks if able".** Requirements (CR 508.1d) do not
  exist in the engine at all (`GoadedBy` is a marker nothing enforces),
  and a requirement interacting with a tax — CR 508.1a's "the player
  is not required to attack if they cannot pay" — has no second half to
  interact with yet.
- **Re-pricing an attack after the declaration.** A Propaganda that
  enters after attackers are declared charges nothing, because CR
  508.1a is a turn-based action that has already happened. That is the
  same rule `restrictions.go`'s "when restrictions are checked"
  section states for the bits, and it needs no code.

## Consequences

- Propaganda, Ghostly Prison and Windborn Muse ship `full`.
  `docs/engine-seams.md`'s "Attack taxes" row moves to Closed.
- ADR 0045's "what this deliberately does not express" list loses its
  first bullet and gains a pointer here; `restrictions.go:57-64` and
  `attack_target.go:27-33` are rewritten to point at `attack_tax.go`
  rather than at a gap.
- A declaration can now FAIL for a reason the enumerator understood in
  advance, which is a new class for the room layer: no undo entry, no
  broadcast, the same as `ErrNoLegalAttackers`.
- Every test that declares an attacker on a board with no tax is
  unchanged, because the price of a free declaration is `""` and the
  payer is never called.

## Alternatives considered

**A sixth `Restriction` bit plus a side table of prices.** Rejected for
the reason ADR 0045 gives: the five bits are absolute and per-permanent
on the ATTACKER, and this is conditional and per-DEFENDER. A bit that
means "look somewhere else" is not a bit.

**Charging per creature inside the staging loop.** Rejected: it makes a
bulk declaration pay for the first two creatures and then fail on the
third, which is not a legal declaration and cannot be undone from
inside the loop.

**A `pay_unless` prompt after the declaration.** Rejected in §3.

**Extending `PriceCast` to take a "what is being priced" discriminator.**
Rejected: `PriceCast`'s body is the CR 601.2 announce sequence
end to end — face, alternative cost, commander tax, optional-cost mana,
the modifier pass. A declaration shares none of it. The machinery worth
sharing is one level down and was already shared.

**Teaching the enumerator to emit subsets so it could price a whole
swing.** Rejected: ADR 0033 §1 ruled subsets out for reasons that have
nothing to do with cost, and the per-verb-call charge composes
correctly without them (§5).
