package game

import (
	"strings"

	"github.com/google/uuid"
)

// block_tax.go is CR 509.1d's cost to block: "As long as this creature
// is attacking, creatures can't block unless their controller pays {1}
// for each of those creatures" — Archangel of Tithes. The mirror of the
// CR 508.1h attack tax (attack_tax.go, ADR 0080), one step over.
// ADR 0221, issue #5375.
//
// # How it differs from the attack tax
//
//   - WHO IT PROTECTS. An attack tax is keyed on the defending player
//     and the "you" is structural. A block tax is printed as "creatures
//     can't block", with no "you": every creature blocking anything is
//     taxed, whoever controls the source. Archangel of Tithes is only
//     ever attacking on its controller's turn, so in practice it taxes
//     every defender's blocks.
//   - WHEN IT IS PAID. An attack is staged and paid in the same verb.
//     Blocks are STAGED (DeclareBlocker / DeclareBlockers) and stay
//     revisable until the defender's declaration completes
//     (block_completion.go), so the price is only fixed at that
//     completion, and that is where it is paid (CR 509.1d-f: the total
//     is determined, mana abilities may be activated, then it is paid
//     in full). Staging a block charges nothing.
//
// # Where it is charged
//
// At each of block_completion.go's completion points that a defender
// can still choose at:
//
//   - finish_blocks (FinishBlocksWith) pays with the action's payment
//     posture — the auto_tap / locked_sources / phyrexian_life trio the
//     attack declaration carries. Unpaid, the finish is REFUSED with a
//     *BlockTaxUnpaidError and the declaration stays pending, staged
//     blocks and all: the defender revises (takes a block back) or
//     makes mana and tries again. Which blocks to drop is their choice,
//     exactly as ADR 0080 leaves which attacks to drop to the attacker.
//   - A pending defender's pass pays with the zero posture (the mana
//     pool, strictly) and is refused unpaid, for the same reason.
//   - The step ending with the defender still pending (completion point
//     4) cannot be refused: the step cannot end on an unfinished
//     turn-based action. The pool is tried strictly; if it is short,
//     the defender's staged blocks are taken back and they have
//     declared none — the one declaration that never owes a block tax,
//     and the only one the engine can make for them without choosing
//     WHICH blocks they keep.
//
// No pay_unless prompt, for ADR 0080 §3's reason: the cost is part of
// the turn-based action, paid by the action that completes it.
//
// # Requirements (CR 509.1c)
//
// "If a creature can't block unless a player pays a cost, that player
// is not required to pay that cost, even if blocking with that creature
// would increase the number of requirements being obeyed." So a
// creature a block tax would charge for is left out of the requirement
// search when it is not already blocking (block_requirements.go): a
// "must block if able" creature under Archangel of Tithes is never
// owed.

// BlockTaxQuery is everything a block tax may look at. Passed by value,
// for AttackTaxQuery's reason. Game is read-only.
type BlockTaxQuery struct {
	// Game is the game the declaration is happening in. Read-only.
	Game *Game

	// Source is the permanent contributing the tax, live off the
	// battlefield — so "as long as this creature is attacking" is an
	// ordinary predicate on Source.AttackingTarget.
	Source Card

	// Blocker is the creature being declared as a blocker, as it stands
	// (post-layers).
	Blocker Card

	// Controller is the blocking creature's controller — "their
	// controller", the player who pays.
	Controller uuid.UUID
}

// BlockTax is one "creatures can't block unless their controller pays
// {N} for each of those creatures" static contributed by a permanent on
// the battlefield. A struct of hooks, for AttackTax's reason.
type BlockTax struct {
	// Label is the clause as printed, for the log line, the client's
	// block-staging hint and the refusal.
	Label string

	// AppliesTo decides whether this tax touches this blocking
	// creature right now. Nil means every blocking creature, always.
	// Archangel of Tithes' "as long as this creature is attacking" is
	// written here (WhileSourceAttacking), because it is a fact about
	// the source at the moment the price is determined (CR 509.1d).
	//
	// Evaluated once per (tax, blocking creature) under the lock.
	// Read-only.
	AppliesTo func(q BlockTaxQuery) bool

	// ManaCost is the mana FOR ONE blocking creature, as a cost string
	// ("{1}"). Concatenated per blocker, AttackTax.ManaCost's reason.
	// Nil, or one returning "", is a tax of nothing.
	ManaCost func(q BlockTaxQuery) string

	// NonMana is the non-mana half of the clause FOR ONE blocking
	// creature (#6190, CR 509.1d: "Costs may include paying mana,
	// tapping permanents, sacrificing permanents, discarding cards, and
	// so on") — Hollow Warrior's "can't ... block unless you tap an
	// untapped creature you control not declared as an attacking or
	// blocking creature this combat", in the vocabulary an ability's
	// cost already uses (combat_cost.go). The zero value asks for
	// nothing, which is every block tax before it.
	//
	// Plain data, not a func: AttackTax.NonMana's reason.
	NonMana CombatCost

	// ActiveWhen is the designation gate, evaluated in
	// BlockTaxesForCard and nowhere else (AttackTax.ActiveWhen's
	// reason). The zero value is "no gate".
	ActiveWhen Designation
}

// BlockedBySelf is the AppliesTo for a tax a creature charges for its
// OWN block — Hollow Warrior's "can't ... block unless you tap an
// untapped creature you control" (#6190). A block tax is printed with
// no "you" and so binds every blocking creature by default; this is the
// predicate that narrows it to the permanent that prints it.
func BlockedBySelf(q BlockTaxQuery) bool {
	return q.Blocker.InstanceID == q.Source.InstanceID
}

// WhileSourceAttacking is the AppliesTo for "as long as this creature
// is attacking" — Archangel of Tithes' block half.
func WhileSourceAttacking(q BlockTaxQuery) bool {
	return q.Source.AttackingTarget != uuid.Nil
}

// CatalogBlockTaxes returns the block taxes a catalog entry declares.
// Wired in carddef.go; nil in a test that builds cards by hand.
var CatalogBlockTaxes func(key string) []BlockTax

// BlockTaxesForCard is the block taxes a permanent contributes right
// now: its catalog entry's, minus those an ability-removing effect took
// (CatalogAbilityKey) and those whose designation gate is unsatisfied.
// The one accessor the engine, the enumerator and the view all read.
func BlockTaxesForCard(c Card) []BlockTax { return blockTaxesOf(&c) }

// blockTaxesOf is BlockTaxesForCard without the copy.
func blockTaxesOf(c *Card) []BlockTax {
	if CatalogBlockTaxes == nil {
		return nil
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogBlockTaxes(key), func(t BlockTax) Designation {
		return t.ActiveWhen
	})
}

// BlockTaxLine is one tax charged by one permanent for one blocking
// creature: the breakdown behind BlockTaxPrice.Cost.
type BlockTaxLine struct {
	// Source is the permanent with the clause, Label that clause.
	Source uuid.UUID
	Label  string

	// Blocker is the creature blocking.
	Blocker uuid.UUID

	// Mana is what this one line charges, e.g. "{1}".
	Mana string
}

// BlockTaxPrice is what one defender's block declaration costs at
// CR 509.1d. The zero value is a free declaration.
type BlockTaxPrice struct {
	// Cost is the concatenated mana, "" when free ("{1}{1}" for two
	// blockers under Archangel of Tithes).
	Cost string

	// Total is Cost parsed, for CanPayFor and the auto-tap planner.
	Total ParsedCost

	// Lines is the per-source, per-blocker breakdown, in battlefield
	// order of the blockers then of the sources.
	Lines []BlockTaxLine

	// NonMana is what the declaration owes in permanents (#6190): one
	// line per (taxing permanent, blocking creature) whose clause asks
	// for a sacrifice, a tap or a return. Paid from the permanents the
	// defender names when they finish blocking, never chosen by the
	// engine.
	NonMana []CombatCostLine
}

// IsFree reports that the declaration costs nothing.
func (p BlockTaxPrice) IsFree() bool { return p.Cost == "" && len(p.NonMana) == 0 }

// Describe is the price as a player reads it: "{1}{1}", "tapping an
// untapped creature you control", or both. "" when free.
func (p BlockTaxPrice) Describe() string {
	nm := combatCostDescribe(p.NonMana)
	switch {
	case nm == "":
		return p.Cost
	case p.Cost == "":
		return nm
	default:
		return p.Cost + " and " + nm
	}
}

// PriceBlockDeclaration prices `seat`'s staged blocks — what finishing
// the declaration would charge right now. Takes the read lock.
func (g *Game) PriceBlockDeclaration(seat uuid.UUID) BlockTaxPrice {
	var out BlockTaxPrice
	g.ReadSnapshot(func() { out = g.priceStagedBlocksLocked(seat) })
	return out
}

// PriceBlockDeclarationForEffect is PriceBlockDeclaration for a caller
// that holds g.mu — the enumerator and the view.
func (g *Game) PriceBlockDeclarationForEffect(seat uuid.UUID) BlockTaxPrice {
	return g.priceStagedBlocksLocked(seat)
}

// PriceBlockersForEffect prices a hypothetical set of blocking
// creatures — the enumerator's "what would this block add". Ids not on
// the battlefield are skipped. Caller must hold g.mu.
func (g *Game) PriceBlockersForEffect(blockers []uuid.UUID) BlockTaxPrice {
	cards := make([]*Card, 0, len(blockers))
	for _, id := range blockers {
		if c := findBattlefieldCard(g, id); c != nil {
			cards = append(cards, c)
		}
	}
	return g.priceBlockersLocked(cards)
}

// priceStagedBlocksLocked prices every creature `seat` controls that is
// blocking. Only meaningful while the seat's declaration is pending —
// after it, the blocks are paid for (CR 509.1d: the total is locked in)
// and nobody asks.
//
// Caller must hold g.mu.
func (g *Game) priceStagedBlocksLocked(seat uuid.UUID) BlockTaxPrice {
	if CatalogBlockTaxes == nil || g.Battlefield == nil || seat == uuid.Nil {
		return BlockTaxPrice{}
	}
	var cards []*Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == seat && c.BlockingTarget != uuid.Nil {
			cards = append(cards, c)
		}
	}
	return g.priceBlockersLocked(cards)
}

// priceBlockersLocked is the one pricer: for each blocking creature,
// every block tax on the battlefield that applies to it. A creature
// blocking several attackers (#1706) is ONE creature and pays once —
// the clause is "for each of those creatures", not per block.
//
// Caller must hold g.mu.
func (g *Game) priceBlockersLocked(blockers []*Card) BlockTaxPrice {
	var out BlockTaxPrice
	if CatalogBlockTaxes == nil || len(blockers) == 0 || g.Battlefield == nil {
		return out
	}
	var b strings.Builder
	for _, blk := range blockers {
		for i := range g.Battlefield.Cards {
			src := &g.Battlefield.Cards[i]
			taxes := blockTaxesOf(src)
			if len(taxes) == 0 {
				continue
			}
			q := BlockTaxQuery{Game: g, Source: *src, Blocker: *blk, Controller: blk.Controller}
			for _, t := range taxes {
				if t.AppliesTo != nil && !t.AppliesTo(q) {
					continue
				}
				if !t.NonMana.Empty() {
					out.NonMana = append(out.NonMana, CombatCostLine{
						Source:   src.InstanceID,
						Label:    t.NonMana.Label,
						Payer:    blk.Controller,
						Creature: blk.InstanceID,
						Cost:     t.NonMana,
					})
				}
				if t.ManaCost == nil {
					continue
				}
				mana := t.ManaCost(q)
				if mana == "" {
					continue
				}
				b.WriteString(mana)
				out.Lines = append(out.Lines, BlockTaxLine{
					Source:  src.InstanceID,
					Label:   t.Label,
					Blocker: blk.InstanceID,
					Mana:    mana,
				})
			}
		}
	}
	out.Cost = b.String()
	if out.Cost != "" {
		// A malformed string is a card-file bug: the payer's own
		// ParseCost judges it, attack_tax.go's rule.
		out.Total, _ = ParseCost(out.Cost)
	}
	return out
}

// blockTaxedLocked reports whether blocking with `c` would cost
// anything — the CR 509.1c test the requirement search asks.
//
// Caller must hold g.mu.
func (g *Game) blockTaxedLocked(c *Card) bool {
	if CatalogBlockTaxes == nil || c == nil {
		return false
	}
	return !g.priceBlockersLocked([]*Card{c}).IsFree()
}

// DeclareBlockersParams is the payment posture of completing a block
// declaration — DeclareAttackersParams' trio, for the same reason. The
// ZERO VALUE is strict: pay from the pool, refuse if short. Inert at a
// table with no block tax.
type DeclareBlockersParams struct {
	// AutoTap lets the completion tap lands for the tax when the pool
	// alone cannot cover it.
	AutoTap bool

	// LockedSources are permanents the auto-tapper must leave alone.
	LockedSources []uuid.UUID

	// PhyrexianLife is how many Phyrexian symbols in the tax are paid
	// with life (CR 107.4f).
	PhyrexianLife int

	// Pay is the permanents the defender names for a NON-MANA cost to
	// block (#6190, CR 509.1d): the creature Hollow Warrior taps. Empty
	// at every table with no such cost on it, and ids sent for a
	// declaration that owes none are refused rather than ignored.
	Pay CombatCostPayment
}

// payBlockTaxLocked charges `seat`'s staged blocks their CR 509.1d
// price. ALL OR NOTHING (CR 509.1f: "partial payments are not
// allowed"): on failure nothing is spent and a *BlockTaxUnpaidError is
// returned. Always Strict, payAttackTaxLocked's reason. The spend
// context is the zero one: a declaration is neither a cast nor an
// activation.
//
// The mana-spent event the shared payer emits names the first line's
// source and clause, which is the log's record of the payment.
//
// Caller must hold g.mu in write mode.
func (g *Game) payBlockTaxLocked(seat uuid.UUID, params DeclareBlockersParams) error {
	price := g.priceStagedBlocksLocked(seat)
	if price.IsFree() {
		return nil
	}
	p := g.playerByIDLocked(seat)
	if p == nil {
		return ErrPlayerNotFound
	}
	// #6190: the permanents first, before any mana is spent, so a
	// refused completion spends nothing.
	if err := g.combatCostRefusalLocked(seat, price.NonMana, params.Pay); err != nil {
		return err
	}
	if price.Cost == "" {
		return g.payCombatCostsLocked(seat, price.NonMana, params.Pay)
	}
	source, label := price.Lines[0].Source, price.Lines[0].Label
	excluded := make(map[uuid.UUID]bool, len(params.LockedSources))
	for _, id := range params.LockedSources {
		excluded[id] = true
	}
	taxCost, perr := ParseCost(price.Cost)
	if perr != nil {
		return &BlockTaxUnpaidError{Cost: price.Cost, Err: ErrInvalidParam}
	}
	if _, err := g.payAbilityManaCostLocked(p, source, label, taxCost, ActivateAbilityParams{
		Strict:        true,
		AutoTap:       params.AutoTap,
		PhyrexianLife: params.PhyrexianLife,
	}, ManaSpendContext{}, excluded); err != nil {
		return &BlockTaxUnpaidError{Cost: price.Cost, Err: err}
	}
	// Last, for payAttackTaxLocked's reason: a sacrifice is a
	// battlefield exit and the mana is already paid.
	return g.payCombatCostsLocked(seat, price.NonMana, params.Pay)
}

// settleBlockTaxForcedLocked is completion point 4's version: the step
// is ending, so the declaration completes whatever happens. The pool is
// tried strictly; if it cannot pay, `seat`'s staged blocks are taken
// back — declared none, the declaration that owes nothing.
//
// Caller must hold g.mu in write mode.
func (g *Game) settleBlockTaxForcedLocked(seat uuid.UUID) {
	if g.blockDeclarationDoneLocked(seat) || !g.isDefendingPlayerLocked(seat) {
		return
	}
	if err := g.payBlockTaxLocked(seat, DeclareBlockersParams{}); err == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == seat && c.BlockingTarget != uuid.Nil {
			c.clearBlocking()
		}
	}
}

// BlockTaxUnpaidError is what a completion refused for want of the
// CR 509.1d cost returns: the price owed and the payer's reason.
type BlockTaxUnpaidError struct {
	// Cost is the declaration's whole price, e.g. "{1}{1}".
	Cost string

	// Err is why it could not be paid.
	Err error
}

func (e *BlockTaxUnpaidError) Error() string {
	if e.Cost == "" {
		return "game: the block tax for this declaration was not paid"
	}
	return "game: blocking costs " + e.Cost + " and it was not paid"
}

// Unwrap lets errors.Is(err, ErrBlockTaxUnpaid) match and errors.As
// reach the payer's own error.
func (e *BlockTaxUnpaidError) Unwrap() []error {
	if e.Err == nil {
		return []error{ErrBlockTaxUnpaid}
	}
	return []error{ErrBlockTaxUnpaid, e.Err}
}
