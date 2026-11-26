package game

import (
	"fmt"
	"slices"
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

	// Attacker is the attacking creature this block is against, for a
	// tax priced per blocked attacker (BlockTax.PerAttacker, #6191):
	// Hipparion's "can't block creatures with power 3 or greater unless
	// you pay {1}", Awesome Presence's "enchanted creature can't be
	// blocked unless defending player pays {3} for each creature they
	// control that's blocking it". The zero Card for every other tax.
	Attacker Card

	// Blocking is every attacking creature the blocker is (or, for a
	// proposed block, would be) blocking, as they stand. Read by a tax
	// charged once per blocking creature that still cares what it
	// blocks: Heat Wave's "can't block creatures you control unless
	// their controller pays 1 life for each blocking creature" (#6191).
	Blocking []Card
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

	// PerAttacker prices this tax once per (blocking creature, attacker
	// it blocks) pair rather than once per blocking creature (#6191):
	// AppliesTo, ManaCost and LifeCost see the pair's attacker on
	// BlockTaxQuery.Attacker. A creature blocking two attackers that
	// both qualify pays twice — Awesome Presence's "for each creature
	// they control that's blocking it" counts blockers of ONE attacker.
	PerAttacker bool

	// LifeCost is the life half of the clause FOR ONE blocking creature
	// (or pair, under PerAttacker) — Heat Wave's "pays 1 life for each
	// blocking creature they control" (#6191, CR 509.1d). Nil, or a
	// non-positive answer, charges none. Checked against the life total
	// before anything is paid (CR 119.4) and paid after the mana.
	LifeCost func(q BlockTaxQuery) int
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

	// Attacker is the attacker of a per-pair line (BlockTax.PerAttacker);
	// uuid.Nil for a line charged per blocking creature.
	Attacker uuid.UUID

	// Life is the life this one line charges (#6191).
	Life int
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

	// Life is the life the declaration owes (#6191): the sum of every
	// line's Life.
	Life int
}

// IsFree reports that the declaration costs nothing.
func (p BlockTaxPrice) IsFree() bool { return p.Cost == "" && p.Life <= 0 && len(p.NonMana) == 0 }

// Describe is the price as a player reads it: "{1}{1}", "tapping an
// untapped creature you control", or both. "" when free.
func (p BlockTaxPrice) Describe() string {
	var parts []string
	if p.Cost != "" {
		parts = append(parts, p.Cost)
	}
	if p.Life > 0 {
		parts = append(parts, fmt.Sprintf("%d life", p.Life))
	}
	if nm := combatCostDescribe(p.NonMana); nm != "" {
		parts = append(parts, nm)
	}
	return strings.Join(parts, " and ")
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
	plans := make([]blockPlan, 0, len(blockers))
	for _, id := range blockers {
		if c := findBattlefieldCard(g, id); c != nil {
			plans = append(plans, g.stagedBlockPlanLocked(c))
		}
	}
	return g.priceBlockPlansLocked(plans)
}

// PriceBlockAdditionForEffect is what declaring `pairs` on top of
// `seat`'s staged blocks would ADD to the declaration's CR 509.1d price
// (#6191): the lines the combined declaration has that the staged one
// does not. A per-blocker tax a staged creature already pays adds
// nothing; a per-attacker tax (Hipparion, Awesome Presence) adds a line
// for each new pair it applies to, and a tax that reads what the blocker
// blocks (Heat Wave) is judged on the combined blocks. The enumerator's
// "what does this block cost". Caller must hold g.mu.
func (g *Game) PriceBlockAdditionForEffect(seat uuid.UUID, pairs []BlockDeclaration) BlockTaxPrice {
	var plans []blockPlan
	index := map[uuid.UUID]int{}
	if g.Battlefield != nil && seat != uuid.Nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller == seat && c.BlockingTarget != uuid.Nil {
				index[c.InstanceID] = len(plans)
				plans = append(plans, g.stagedBlockPlanLocked(c))
			}
		}
	}
	staged := g.priceBlockPlansLocked(plans)
	combined := make([]blockPlan, len(plans))
	for i, p := range plans {
		combined[i] = blockPlan{card: p.card, attackers: append([]*Card(nil), p.attackers...)}
	}
	for _, pr := range pairs {
		atk := findBattlefieldCard(g, pr.Attacker)
		if atk == nil {
			continue
		}
		i, ok := index[pr.Blocker]
		if !ok {
			blk := findBattlefieldCard(g, pr.Blocker)
			if blk == nil {
				continue
			}
			i = len(combined)
			index[pr.Blocker] = i
			combined = append(combined, blockPlan{card: blk})
		}
		if !slices.ContainsFunc(combined[i].attackers, func(c *Card) bool { return c.InstanceID == atk.InstanceID }) {
			combined[i].attackers = append(combined[i].attackers, atk)
		}
	}
	return blockPriceMinus(g.priceBlockPlansLocked(combined), staged)
}

// blockPriceMinus is the lines of `full` that `base` does not have,
// matched on (source, label, blocker, attacker); the non-mana lines on
// (source, label, creature).
func blockPriceMinus(full, base BlockTaxPrice) BlockTaxPrice {
	type key struct {
		src, blk, atk uuid.UUID
		label         string
	}
	have := map[key]int{}
	for _, l := range base.Lines {
		have[key{l.Source, l.Blocker, l.Attacker, l.Label}]++
	}
	var out BlockTaxPrice
	var b strings.Builder
	for _, l := range full.Lines {
		k := key{l.Source, l.Blocker, l.Attacker, l.Label}
		if have[k] > 0 {
			have[k]--
			continue
		}
		out.Lines = append(out.Lines, l)
		b.WriteString(l.Mana)
		out.Life += l.Life
	}
	nm := map[key]int{}
	for _, l := range base.NonMana {
		nm[key{l.Source, l.Creature, uuid.Nil, l.Label}]++
	}
	for _, l := range full.NonMana {
		k := key{l.Source, l.Creature, uuid.Nil, l.Label}
		if nm[k] > 0 {
			nm[k]--
			continue
		}
		out.NonMana = append(out.NonMana, l)
	}
	out.Cost = b.String()
	if out.Cost != "" {
		out.Total, _ = ParseCost(out.Cost)
	}
	return out
}

// blockPlan is one blocking creature and the attackers it blocks (or
// would block), the unit the pricer charges.
type blockPlan struct {
	card      *Card
	attackers []*Card
}

// stagedBlockPlanLocked is `c` with the attackers it is blocking now.
func (g *Game) stagedBlockPlanLocked(c *Card) blockPlan {
	p := blockPlan{card: c}
	if c.BlockingTarget == uuid.Nil {
		return p
	}
	for _, id := range append([]uuid.UUID{c.BlockingTarget}, c.AlsoBlocking...) {
		if a := findBattlefieldCard(g, id); a != nil {
			p.attackers = append(p.attackers, a)
		}
	}
	return p
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
	var plans []blockPlan
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == seat && c.BlockingTarget != uuid.Nil {
			plans = append(plans, g.stagedBlockPlanLocked(c))
		}
	}
	return g.priceBlockPlansLocked(plans)
}

// priceBlockPlansLocked is the one pricer: for each blocking creature,
// every block tax on the battlefield that applies to it. A creature
// blocking several attackers (#1706) is ONE creature and pays a
// per-blocker tax once — the clause is "for each of those creatures",
// not per block; a PerAttacker tax (#6191) is asked once per attacker
// it blocks instead.
//
// Caller must hold g.mu.
func (g *Game) priceBlockPlansLocked(plans []blockPlan) BlockTaxPrice {
	var out BlockTaxPrice
	if CatalogBlockTaxes == nil || len(plans) == 0 || g.Battlefield == nil {
		return out
	}
	var b strings.Builder
	for _, plan := range plans {
		blk := plan.card
		blocking := make([]Card, 0, len(plan.attackers))
		for _, a := range plan.attackers {
			blocking = append(blocking, *a)
		}
		for i := range g.Battlefield.Cards {
			src := &g.Battlefield.Cards[i]
			taxes := blockTaxesOf(src)
			if len(taxes) == 0 {
				continue
			}
			base := BlockTaxQuery{Game: g, Source: *src, Blocker: *blk, Controller: blk.Controller, Blocking: blocking}
			for _, t := range taxes {
				if !t.PerAttacker {
					g.priceBlockTaxLineLocked(&out, &b, t, base, src.InstanceID, uuid.Nil)
					continue
				}
				for _, a := range plan.attackers {
					q := base
					q.Attacker = *a
					g.priceBlockTaxLineLocked(&out, &b, t, q, src.InstanceID, a.InstanceID)
				}
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

// priceBlockTaxLineLocked adds what one tax charges for one blocking
// creature (or one pair) to `out`.
func (g *Game) priceBlockTaxLineLocked(out *BlockTaxPrice, b *strings.Builder, t BlockTax, q BlockTaxQuery, src, attacker uuid.UUID) {
	if t.AppliesTo != nil && !t.AppliesTo(q) {
		return
	}
	if !t.NonMana.Empty() {
		out.NonMana = append(out.NonMana, CombatCostLine{
			Source:   src,
			Label:    t.NonMana.Label,
			Payer:    q.Controller,
			Creature: q.Blocker.InstanceID,
			Cost:     t.NonMana,
		})
	}
	mana := ""
	if t.ManaCost != nil {
		mana = t.ManaCost(q)
	}
	life := 0
	if t.LifeCost != nil {
		life = max(t.LifeCost(q), 0)
	}
	if mana == "" && life == 0 {
		return
	}
	b.WriteString(mana)
	out.Life += life
	out.Lines = append(out.Lines, BlockTaxLine{
		Source:   src,
		Label:    t.Label,
		Blocker:  q.Blocker.InstanceID,
		Mana:     mana,
		Attacker: attacker,
		Life:     life,
	})
}

// blockTaxedLocked reports whether blocking with `c` would cost
// anything — the CR 509.1c test the requirement search asks.
//
// Caller must hold g.mu.
func (g *Game) blockTaxedLocked(c *Card) bool {
	if CatalogBlockTaxes == nil || c == nil {
		return false
	}
	return !g.priceBlockPlansLocked([]blockPlan{{card: c}}).IsFree()
}

// blockPairTaxedLocked reports whether `c` blocking `attacker` would
// cost anything — CR 509.1c's test for a pair, which a tax that reads
// the attacker (#6191) can only answer per pair.
//
// Caller must hold g.mu.
func (g *Game) blockPairTaxedLocked(c, attacker *Card) bool {
	if CatalogBlockTaxes == nil || c == nil || attacker == nil {
		return false
	}
	return !g.priceBlockPlansLocked([]blockPlan{{card: c, attackers: []*Card{attacker}}}).IsFree()
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
	// #6191, CR 119.4: the life is checked before anything is paid, with
	// the life the defender said they would pay for Phyrexian symbols.
	if price.Life > 0 {
		need := price.Life + max(params.PhyrexianLife, 0)*PhyrexianLifePerSymbol
		if !g.CanPayLifeLocked(p, need) {
			return &BlockTaxUnpaidError{Cost: price.Describe(), Err: fmt.Errorf("%w: paying %d life needs a life total of at least %d (CR 119.4)", ErrInvalidParam, need, need)}
		}
	}
	source, label := uuid.Nil, "block tax"
	if len(price.Lines) > 0 {
		source, label = price.Lines[0].Source, price.Lines[0].Label
	}
	if price.Cost == "" {
		if err := g.payBlockTaxLifeLocked(seat, source, price); err != nil {
			return err
		}
		return g.payCombatCostsLocked(seat, price.NonMana, params.Pay)
	}
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
	if err := g.payBlockTaxLifeLocked(seat, source, price); err != nil {
		return err
	}
	// Last, for payAttackTaxLocked's reason: a sacrifice is a
	// battlefield exit and the mana is already paid.
	return g.payCombatCostsLocked(seat, price.NonMana, params.Pay)
}

// payBlockTaxLifeLocked pays the life half of a block tax (#6191) once
// its mana is paid: paying life is losing it (CR 119.4), through
// payLifeAsCostLocked's CR 614 window. Validated before anything was
// paid, so a refusal here is that function's backstop.
//
// Caller must hold g.mu in write mode.
func (g *Game) payBlockTaxLifeLocked(seat, source uuid.UUID, price BlockTaxPrice) error {
	if price.Life <= 0 {
		return nil
	}
	if err := g.payLifeAsCostLocked(source, seat, price.Life); err != nil {
		return &BlockTaxUnpaidError{Cost: price.Describe(), Err: err}
	}
	return nil
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
