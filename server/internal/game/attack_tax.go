package game

import (
	"strings"

	"github.com/google/uuid"
)

// attack_tax.go is CR 508.1a's cost to attack: "Creatures can't attack
// you unless their controller pays {2} for each creature they control
// that's attacking you" — Propaganda, Ghostly Prison, Windborn Muse,
// Sphere of Safety, Collective Restraint, Norn's Annex, War Tax.
// ADR 0080, issue #1063.
//
// # Why this is not a Restriction bit
//
// ADR 0045's vocabulary is five ABSOLUTE bits and it says so in its
// own words. Two things about an attack tax are the wrong shape for
// one:
//
//   - It is a COST, not a prohibition. A bit cannot carry a price, and
//     a bit that means "look somewhere else for the price" is not a
//     bit.
//   - It is keyed on the DEFENDING PLAYER, not on the attacking
//     creature. Restrictions ride on the permanent the effect is
//     attached to and are read off that permanent's effective
//     characteristics; "attacking me costs {2}" is a fact about a seat
//     that the attacker's characteristics cannot hold.
//
// So the tax is a catalog-contributed static in the shape
// CostModifier already uses (ADR 0048): a struct of hooks a card file
// writes as a literal, collected through a catalog key, gated by a
// designation in exactly one accessor.
//
// # Where it is charged
//
// At CR 508.1a, on the DECLARATION, by the declaration verb — after
// eligibility is settled and before anything is staged. Never as a
// pay_unless prompt afterwards: CR 508.1a is part of the turn-based
// action, alongside the CR 508.1f tap, so it is paid by the action
// that declares the attack exactly as a cast's additional costs are
// paid by the action that casts the spell (ADR 0073 §4).
//
// The payment goes through payAbilityManaCostLocked — the payer
// activated abilities and special actions already share — so the
// declaration taps through materializePlanLocked and ManaTrigger
// fires at the production site with no second copy of anything
// (ADR 0074).
//
// # What is deliberately not here
//
//   - COUNT limits (Silent Arbiter, Crawlspace). A count is a property
//     of the whole declaration, not a price on it; it belongs beside
//     blockerBoundsLocked as a set-shaped predicate and keeps its own
//     seam row. ADR 0080 §8.
//   - A NON-MANA tax. Nothing printed charges one (Norn's Annex's
//     {W/P} is mana with a life alternative, which the Phyrexian
//     machinery already handles). A Sacrifice or PayLife sibling of
//     ManaCost is an addition to this struct and one more branch in
//     the payer, the shape AdditionalCost has with its three
//     components — not a redesign.
//   - A tax with a DURATION (War Tax's "this turn"). The collection
//     below reads catalog statics off the battlefield; a
//     duration-carrying registry is the SECOND source it takes without
//     changing AttackTaxesForCard's signature, which is the extension
//     point ADR 0073 §8 reserved for CastRestriction.

// AttackTaxQuery is everything an attack tax may look at. Passed by
// value, for CostQuery's reason: the tax is consulted inside the
// declaration path under the write lock, and handing a card file
// pointers into the battlefield slice would invite it to mutate the
// board while pricing an attack.
//
// Game is a pointer because the interesting predicates are board-wide
// reads — Sphere of Safety counts enchantments, Collective Restraint
// counts basic land types. Treat it as READ-ONLY: a *ForEffect
// accessor is fine, a mutator is a bug and a public locking mutator is
// a deadlock.
type AttackTaxQuery struct {
	// Game is the game the declaration is happening in. Read-only.
	Game *Game

	// Source is the permanent contributing the tax, live off the
	// battlefield — so "as long as this is untapped" is an ordinary
	// predicate rather than machinery, the way CostQuery.Source's
	// Tapped bit is.
	Source Card

	// Defender is the player being attacked: the "you" of "creatures
	// can't attack you". Always Source.Controller — see the note on
	// AttackTax — and carried here so a predicate does not have to
	// re-derive it.
	//
	// For an attack on a planeswalker or a battle it is the seat that
	// permanent's controller or protector is, which
	// defendingPlayerForAttackLocked resolves. Norn's Annex prints
	// "you or planeswalkers you control" for exactly this reason.
	Defender uuid.UUID

	// Attacker is the creature being declared, as it stands at
	// declaration (post-layers).
	Attacker Card

	// Controller is the attacking creature's controller — "their
	// controller", the player who pays.
	Controller uuid.UUID

	// TargetKind is what the declaration actually names: the defending
	// player themselves, a planeswalker they control, or a battle they
	// protect (CR 508.1a). Scope handles the distinction every printed
	// card makes; this is here for a predicate that needs more.
	TargetKind AttackTargetKind

	// Target is the id of that thing — a seat id for a player, an
	// instance id for a permanent.
	Target uuid.UUID
}

// AttackTaxScope is WHAT a tax defends, which is the one line every
// card in the family prints differently and the one an implementation
// can get silently wrong in the direction of playing stronger than
// printed.
//
// Propaganda and Ghostly Prison say "creatures can't attack YOU": an
// attack on their controller's planeswalker is free. Norn's Annex and
// Sphere of Safety say "you OR PLANESWALKERS YOU CONTROL". The
// difference is worth a field rather than a predicate because the zero
// value has to be the narrower one — a scope a card file forgets must
// under-tax, never over-tax.
type AttackTaxScope uint8

const (
	// AttackTaxOnPlayer is "creatures can't attack you" — Propaganda,
	// Ghostly Prison, Windborn Muse. Only an attack naming the seat
	// itself is taxed. THE ZERO VALUE, deliberately.
	AttackTaxOnPlayer AttackTaxScope = iota

	// AttackTaxOnPlayerOrPlaneswalkers is "creatures can't attack you
	// or planeswalkers you control" — Sphere of Safety, Norn's Annex.
	//
	// Planeswalkers, and not battles: no printed card in the family
	// mentions battles, and a battle's defending player is its
	// PROTECTOR (CR 310.9), who is usually not the player whose
	// enchantment this is. Widening the scope to cover them would tax
	// an attack the card says nothing about.
	AttackTaxOnPlayerOrPlaneswalkers
)

// covers reports whether this scope taxes an attack naming `kind`.
func (s AttackTaxScope) covers(kind AttackTargetKind) bool {
	switch kind {
	case AttackTargetPlayer:
		return true
	case AttackTargetPlaneswalker:
		return s == AttackTaxOnPlayerOrPlaneswalkers
	default:
		return false
	}
}

// AttackTax is one "creatures can't attack you unless their controller
// pays {N} for each of them" static contributed by a permanent on the
// battlefield.
//
// Declared as a struct of hooks rather than an interface for the
// reason StaticAbility, CostModifier and ReplacementEffect are: a card
// file writes a literal, not a type.
//
// THE "YOU" IS STRUCTURAL, NOT A PREDICATE. A tax protects its
// source's controller and nobody else. Every printed card in the
// family says "creatures can't attack YOU", and making that a
// predicate would let one card file typo its way into taxing the whole
// table. AppliesTo narrows WHICH attacking creatures are taxed, which
// is the axis real cards vary on. A card that taxes attacks on someone
// else does not exist; when one does it gets a field, not a silent
// reinterpretation of this one.
type AttackTax struct {
	// Label is the clause as printed ("Creatures can't attack you
	// unless their controller pays {2} for each creature they control
	// that's attacking you"), for the log line, the client's tooltip
	// and for debugging a declaration that came out at an unexpected
	// price.
	Label string

	// Scope is what the tax defends: the controller alone
	// (Propaganda's "can't attack you", the zero value) or the
	// controller and their planeswalkers (Sphere of Safety, Norn's
	// Annex). Checked before AppliesTo, in the pricer.
	Scope AttackTaxScope

	// AppliesTo decides whether this tax touches this attacking
	// creature. Nil means "every creature attacking me", which is
	// Propaganda, Ghostly Prison and Windborn Muse exactly.
	//
	// Evaluated once per (tax, attacking creature) under the write
	// lock. Read-only.
	AppliesTo func(q AttackTaxQuery) bool

	// ManaCost is the mana half of the clause FOR ONE attacking
	// creature, as a cost string: "{2}", "{W/P}", or a count rendered
	// once ("{6}" for a Sphere of Safety beside five other
	// enchantments).
	//
	// A string rather than a ParsedCost because a declaration sums
	// several and the sum is CONCATENATION — "{2}" three times is
	// "{2}{2}{2}", which ParseCost reads as six generic with no
	// arithmetic to get wrong and no normaliser to write. Hybrid and
	// Phyrexian symbols are what the parser already eats, so a
	// {W/P} tax rides strikePhyrexianLifeLocked unchanged.
	//
	// A function rather than a constant because the family scales with
	// the board. Nil, or one returning "", is a tax of nothing rather
	// than an error — the rule CostModifier.Amount follows: a clause
	// that names no price simply does not apply.
	ManaCost func(q AttackTaxQuery) string

	// ActiveWhen is the CR 716 / 719 / 721 / 709.5 designation gate:
	// this tax applies only while its source has the designation
	// named. The zero value is "no gate".
	//
	// Evaluated in AttackTaxesForCard and NOWHERE ELSE, so a gated-off
	// tax never reaches the pricer — which means the engine and the
	// legal-move enumerator price a declaration identically, because
	// both read the same accessor (#544).
	ActiveWhen Designation
}

// CatalogAttackTaxes returns the attack taxes a catalog entry
// declares. Wired to the card catalog in carddef.go; nil in a test
// that builds cards by hand, which is the same contract every other
// catalog hook has.
var CatalogAttackTaxes func(key string) []AttackTax

// AttackTaxesForCard is the attack taxes a permanent contributes right
// now: its catalog entry's taxes, minus those a CR 613.1f
// ability-removing effect took away (CatalogAbilityKey) and minus
// those whose designation gate is unsatisfied.
//
// The one accessor. Both the engine's pricer and the legal-move
// enumerator reach a card's taxes through it, which is what keeps the
// price the bot is offered and the price the engine charges from
// drifting apart.
func AttackTaxesForCard(c Card) []AttackTax {
	if CatalogAttackTaxes == nil {
		return nil
	}
	key := CatalogAbilityKey(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogAttackTaxes(key), func(t AttackTax) Designation {
		return t.ActiveWhen
	})
}

// AttackTaxLine is one tax charged by one permanent for one attacking
// creature: the per-source breakdown behind AttackTaxPrice.Cost.
//
// Kept because the total alone cannot answer "why", and three
// consumers need the answer: the log line that says which enchantment
// charged, the client's per-target label, and the richer picker filed
// as #1063's follow-up, which subsets a declaration by what the seat
// can afford.
type AttackTaxLine struct {
	// Source is the permanent with the clause.
	Source uuid.UUID

	// Label is that clause as printed.
	Label string

	// Defender is the seat being attacked, and Attacker the creature
	// attacking it.
	Defender uuid.UUID
	Attacker uuid.UUID

	// Mana is what this one line charges, e.g. "{2}".
	Mana string
}

// AttackTaxPrice is what one attack declaration costs at CR 508.1a.
//
// The zero value is a free declaration — every table with no tax on it
// — and IsFree is the one test callers make, so a board the catalog
// says nothing about never reaches the parser or the payer.
type AttackTaxPrice struct {
	// Cost is the concatenated mana cost of the whole declaration,
	// "" when it is free. "{2}{2}" rather than "{4}" on purpose: see
	// AttackTax.ManaCost.
	Cost string

	// Total is Cost parsed, for ManaPool.CanPayFor and the auto-tap
	// planner. Parsed once here so no caller re-parses and no two
	// callers can disagree about what the string meant.
	Total ParsedCost

	// Lines is the per-source, per-creature breakdown, in declaration
	// order then battlefield order.
	Lines []AttackTaxLine
}

// IsFree reports that this declaration costs nothing, which is every
// declaration at a table with no attack tax on the battlefield.
func (p AttackTaxPrice) IsFree() bool { return p.Cost == "" }

// PriceAttackDeclaration prices a whole attack declaration at
// CR 508.1a. Takes the read lock.
//
// PriceCast's sibling, and deliberately not PriceCast itself:
// PriceCast's body is the CR 601.2 announce sequence end to end —
// face materialisation, the alternative-cost swap, the commander tax,
// optional-cost mana, the CR 601.2f modifier pass — and a declaration
// shares none of it. What the two share is the mana machinery one
// level down (ParseCost, ManaPool.CanPayFor, autoTapLocked,
// materializePlanLocked), which payAbilityManaCostLocked already
// shared between casts, activations and special actions before this
// existed.
func (g *Game) PriceAttackDeclaration(decls []AttackDeclaration) AttackTaxPrice {
	var out AttackTaxPrice
	g.ReadSnapshot(func() { out = g.priceAttackDeclarationLocked(decls) })
	return out
}

// PriceAttackDeclarationForEffect is PriceAttackDeclaration for a
// caller that already holds g.mu — the declaration verbs and the
// legal-move enumerator.
//
// Caller must hold g.mu.
func (g *Game) PriceAttackDeclarationForEffect(decls []AttackDeclaration) AttackTaxPrice {
	return g.priceAttackDeclarationLocked(decls)
}

// priceAttackDeclarationLocked is the body both of the above run, so
// the price the client previews, the price the enumerator filters on
// and the price the engine charges are one function's answer.
//
// For each declared attack it resolves the DEFENDING PLAYER — which
// for a planeswalker or a battle is that permanent's controller or
// protector, CR 508.1a — and asks every tax that player's permanents
// contribute. A declaration naming several defenders is priced against
// each of their boards separately, so Alice's Propaganda never prices
// an attack on Bob.
//
// Caller must hold g.mu.
func (g *Game) priceAttackDeclarationLocked(decls []AttackDeclaration) AttackTaxPrice {
	var out AttackTaxPrice
	if len(decls) == 0 || CatalogAttackTaxes == nil {
		// No catalog wired is every hand-built test board, and the
		// early-out keeps the battlefield walk below off a path that
		// runs once per (creature, target) pair in the enumerator.
		return out
	}
	var b strings.Builder
	for _, d := range decls {
		attacker := findBattlefieldCard(g, d.Attacker)
		if attacker == nil {
			continue
		}
		defender := g.defendingPlayerForAttackLocked(d.Target)
		if defender == uuid.Nil {
			continue
		}
		kind := g.classifyAttackTargetLocked(d.Target)
		for i := range g.Battlefield.Cards {
			src := &g.Battlefield.Cards[i]
			// The "you" is structural: only the defending player's
			// own permanents tax an attack on them.
			if src.Controller != defender {
				continue
			}
			taxes := AttackTaxesForCard(*src)
			if len(taxes) == 0 {
				continue
			}
			q := AttackTaxQuery{
				Game:       g,
				Source:     *src,
				Defender:   defender,
				Attacker:   *attacker,
				Controller: attacker.Controller,
				TargetKind: kind,
				Target:     d.Target,
			}
			for _, t := range taxes {
				if !t.Scope.covers(kind) {
					continue
				}
				if t.AppliesTo != nil && !t.AppliesTo(q) {
					continue
				}
				if t.ManaCost == nil {
					continue
				}
				mana := t.ManaCost(q)
				if mana == "" {
					continue
				}
				b.WriteString(mana)
				out.Lines = append(out.Lines, AttackTaxLine{
					Source:   src.InstanceID,
					Label:    t.Label,
					Defender: defender,
					Attacker: d.Attacker,
					Mana:     mana,
				})
			}
		}
	}
	out.Cost = b.String()
	if out.Cost == "" {
		return out
	}
	// A malformed cost string is a card-file bug, not a board state:
	// leave Total zero and let the payer's own ParseCost return
	// ErrInvalidParam, so there is exactly one place the parse is
	// judged.
	out.Total, _ = ParseCost(out.Cost)
	return out
}

// attackTaxAffordableLocked reports whether `payer` could pay `price`
// right now, without paying it: the mana pool first, then the
// auto-tapper when `params` allows it.
//
// Exactly `legal.enumerator.canPayExcluding`'s test, which is the
// point — the enumerator decides what to OFFER with it and this
// decides what to accept, so the two cannot disagree (#544).
//
// It exists for one caller: a bulk declaration that mixes seats (only
// an admin session can submit one) has to know every payer can pay
// BEFORE it charges any of them, or a refusal leaves the first seat's
// mana spent on a declaration that never happened. The check is exact
// rather than optimistic because the seats' pools and untapped
// permanents are disjoint.
//
// Caller must hold g.mu.
func (g *Game) attackTaxAffordableLocked(payer uuid.UUID, price AttackTaxPrice, params DeclareAttackersParams) error {
	if price.IsFree() {
		return nil
	}
	p := g.playerByIDLocked(payer)
	if p == nil {
		return ErrPlayerNotFound
	}
	cost, err := ParseCost(price.Cost)
	if err != nil {
		return &AttackTaxUnpaidError{Cost: price.Cost, Err: ErrUnparseableCost}
	}
	// The Phyrexian symbols the player said they would pay with life
	// leave the mana cost first, exactly as the payer strikes them.
	cost, _ = PhyrexianLifePlan(cost, p.ManaPool, ManaSpendContext{}, params.PhyrexianLife)
	if p.ManaPool.CanPayFor(cost, 0, ManaSpendContext{}) {
		return nil
	}
	if params.AutoTap {
		excluded := make(map[uuid.UUID]bool, len(params.LockedSources))
		for _, id := range params.LockedSources {
			excluded[id] = true
		}
		if _, ok := g.autoTapLocked(payer, cost, 0, excluded); ok {
			return nil
		}
	}
	return &AttackTaxUnpaidError{
		Cost: price.Cost,
		Err:  &InsufficientManaError{Missing: p.ManaPool.MissingFor(cost, 0, ManaSpendContext{})},
	}
}

// payAttackTaxLocked charges `price` to `payer` as part of declaring
// attackers (CR 508.1a).
//
// ALL OR NOTHING. A declaration whose tax cannot be paid in full is
// not a legal declaration, so this returns an error and spends
// nothing; the caller must not have staged anything yet. Which attacks
// to drop when the seat can only afford some of them is the player's
// choice, made by submitting a smaller declaration, and the engine
// must not make it for them.
//
// There is no permissive mode. payAbilityManaCostLocked's
// "!Strict && !AutoTap → warn and mark it paid on paper" branch exists
// for the sandbox's hand-tracked mana; an attack tax waived on paper
// is Propaganda as a blank, which is the outcome this file exists to
// prevent. Strict is therefore always true, which is also what makes
// the zero value of DeclareAttackersParams the strict posture — pay
// from the pool, refuse if short — rather than the lax one.
//
// The spend context is the zero ManaSpendContext, the decision
// special_action.go and payCostLocked already document: a declaration
// is neither a cast nor an activation, so mana that may only be spent
// to cast creature spells cannot fund it.
//
// `source` is the permanent named on the mana-spent event. The first
// line's source is used, which is the enchantment a one-tax table has
// and an arbitrary-but-stable pick when two of them charge; the LINES
// carry the real breakdown.
//
// Caller must hold g.mu in write mode.
func (g *Game) payAttackTaxLocked(payer uuid.UUID, price AttackTaxPrice, params DeclareAttackersParams) error {
	if price.IsFree() {
		return nil
	}
	p := g.playerByIDLocked(payer)
	if p == nil {
		return ErrPlayerNotFound
	}
	source, label := uuid.Nil, "attack tax"
	if len(price.Lines) > 0 {
		source, label = price.Lines[0].Source, price.Lines[0].Label
	}
	excluded := make(map[uuid.UUID]bool, len(params.LockedSources))
	for _, id := range params.LockedSources {
		excluded[id] = true
	}
	_, err := g.payAbilityManaCostLocked(p, source, label, price.Cost, ActivateAbilityParams{
		Strict:        true,
		AutoTap:       params.AutoTap,
		PhyrexianLife: params.PhyrexianLife,
	}, ManaSpendContext{}, excluded)
	if err != nil {
		return &AttackTaxUnpaidError{Cost: price.Cost, Err: err}
	}
	return nil
}

// AttackTaxUnpaidError is what a declaration refused for want of the
// CR 508.1a cost returns: the price that was owed, and the payer's own
// reason for failing it (an *InsufficientManaError naming the missing
// symbols, usually).
//
// A struct rather than a bare sentinel for CantCastError's reason: the
// client's toast has to say what the attack cost and what was short,
// and a bool would have meant a second call to find out.
type AttackTaxUnpaidError struct {
	// Cost is the declaration's whole price, e.g. "{2}{2}".
	Cost string

	// Err is why it could not be paid, unwrapped so
	// errors.As(err, &InsufficientManaError{}) reaches the missing
	// symbols.
	Err error
}

func (e *AttackTaxUnpaidError) Error() string {
	if e.Cost == "" {
		return "game: the attack tax for this declaration was not paid"
	}
	return "game: attacking costs " + e.Cost + " and it was not paid"
}

// Unwrap lets errors.Is(err, ErrAttackTaxUnpaid) match, and — through
// the joined list — errors.As reach the payer's own error.
func (e *AttackTaxUnpaidError) Unwrap() []error {
	if e.Err == nil {
		return []error{ErrAttackTaxUnpaid}
	}
	return []error{ErrAttackTaxUnpaid, e.Err}
}

// DeclareAttackersParams is the payment posture of an attack
// declaration — the same trio cast_spell, activate_ability and
// special_action carry.
//
// The ZERO VALUE is strict: pay the tax out of the mana pool, refuse
// the declaration if the pool is short. That is deliberate, and it is
// why there is no Strict field to forget to set. See
// payAttackTaxLocked.
//
// Every field is inert at a table with no attack tax on it, which is
// why DeclareAttacker and DeclareAttackers keep their signatures and
// pass the zero value.
type DeclareAttackersParams struct {
	// AutoTap lets the declaration tap lands for the tax when the
	// mana pool alone cannot cover it, through the same planner and
	// the same executor a cast uses — so ManaTrigger fires for the
	// taps (ADR 0074).
	AutoTap bool

	// LockedSources are permanents the player has told the auto-tapper
	// to leave alone, the field cast_spell already carries.
	LockedSources []uuid.UUID

	// PhyrexianLife is how many Phyrexian symbols in the tax the
	// player is paying with life instead of mana (CR 107.4f) — Norn's
	// Annex's {W/P}, when that card is written.
	PhyrexianLife int
}
