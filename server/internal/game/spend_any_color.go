package game

import "github.com/google/uuid"

// spend_any_color.go — "you may spend mana as though it were mana of any
// color" as a PLAYER's static (#1600, ADR 0066's 2026-10-02 amendment):
// Chromatic Orrery for its controller, Mycosynth Lattice for every
// player, Oath of Nissa for its controller's planeswalker spells.
//
// THE RULE. CR 609.4b: an effect that lets a player spend mana "as
// though it were mana of any color" changes how that player may PAY a
// cost. It changes neither the cost nor the mana: a {2}{U} is still a
// {2}{U}, a Mountain still makes red, and "spend this mana only to cast
// creature spells" still binds (the Chromatic Orrery rulings). Any mana
// may pay a coloured symbol, colourless included. Coloured mana still
// cannot pay a {C}, because colourless is not a colour (CR 106.1b).
//
// WHY NOT THE PERMISSION FOLD. A cast permission's AnyColor
// (CastPermission, spendAsThoughAny) is about ONE spell, so it lives in
// the one cast pricer. This grant is about a PLAYER and every cost they
// pay: spells, activated and mana abilities, special actions, attack
// taxes, and the "unless pays" family (ward, Rhystic Study, cumulative
// upkeep). Those are priced by five different functions and paid by
// six, so the grant is read at the PAYMENT, not at a pricer:
// costAsPaidByLocked is called once at the top of each payment and each
// affordability probe, and the solvers do the rest. A pricer is never
// asked, so the price a player is SHOWN — `cast_prices`, an ability
// row's charged cost, the preview's `cost` — stays the printed one,
// which is what CR 609.4b says it is.
//
// WIDENED, NOT FOLDED. The permission fold moves a coloured symbol into
// the generic demand. That throws away which colour the card printed,
// and the solvers' colourless-first generic order would then pay a
// Firespout's {R/G} with the Orrery's colourless and lose its red mode.
// "You MAY spend" never makes a card worse, so this fold marks each
// coloured requirement ColorRequirement.AnyMana and keeps its Options.
// Both solvers try the printed colours first and pay a widened symbol
// as generic only when nothing of its colour is left
// (ManaPool.attemptSpend, solveColored). The cost a player could pay
// without the grant is paid exactly as it would have been.
//
// AFTER the taps. Convoke and waterbend (CR 702.51, CR 701.67a) are
// not mana, and a green creature still cannot convoke a {U} under an
// Orrery. Every payment site calls this after the tap subtraction and
// the delve reduction, so the widening reaches only what is left for
// mana to pay. Before the Phyrexian strike, as the permission fold is,
// so a widened {B/P} keeps its "or 2 life" half (#1589).

// AnyColorSpendWhose is which players a printed "spend mana as though
// it were mana of any color" static covers.
type AnyColorSpendWhose uint8

const (
	// AnyColorSpendYou is "You may spend mana as though it were mana
	// of any color": the permanent's controller (Chromatic Orrery,
	// Oath of Nissa).
	AnyColorSpendYou AnyColorSpendWhose = iota
	// AnyColorSpendEveryPlayer is "Players may spend mana as though it
	// were mana of any color" (Mycosynth Lattice).
	AnyColorSpendEveryPlayer
)

// AnyColorSpendStatic is one printed "spend mana as though it were mana
// of any color" static on a permanent. Catalog data, never stored:
// read live off the battlefield, keyed by CatalogAbilityKey, so two
// Orreries compose, one leaving cannot revoke the other's grant, and an
// Orrery that has lost its abilities grants nothing.
type AnyColorSpendStatic struct {
	// Label is the clause as printed.
	Label string
	// Whose is which players it covers.
	Whose AnyColorSpendWhose
	// Covers narrows the grant to some payments, asked of the spend
	// context the payment is made under — "to cast planeswalker
	// spells" (Oath of Nissa). Nil covers every cost the player pays.
	// A context with no object (an "unless pays" cost, an attack tax)
	// is a payment a narrowed grant does not reach.
	Covers func(ManaSpendContext) bool
}

// CatalogAnyColorSpend returns the spend statics a permanent with the
// given catalog key has. carddef.go sets it from CardDef.AnyColorSpend;
// a game-package test may stub it.
var CatalogAnyColorSpend func(key string) []AnyColorSpendStatic

// spendsManaAsAnyColorLocked is THE reader: may `payer` spend mana as
// though it were mana of any color for the payment `ctx` describes?
//
// Caller must hold g.mu (read or write). Reads only.
func (g *Game) spendsManaAsAnyColorLocked(payer uuid.UUID, ctx ManaSpendContext) bool {
	if CatalogAnyColorSpend == nil || g.Battlefield == nil || payer == uuid.Nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, s := range CatalogAnyColorSpend(key) {
			switch s.Whose {
			case AnyColorSpendYou:
				if src.Controller != payer {
					continue
				}
			case AnyColorSpendEveryPlayer:
			default:
				continue
			}
			if s.Covers != nil && !s.Covers(ctx) {
				continue
			}
			return true
		}
	}
	return false
}

// SpendsManaAsAnyColorForEffect is the *ForEffect surface over the
// reader, for the view and card files. Caller must hold g.mu.
func (g *Game) SpendsManaAsAnyColorForEffect(payer uuid.UUID, ctx ManaSpendContext) bool {
	return g.spendsManaAsAnyColorLocked(payer, ctx)
}

// widenForAnyColorSpend is `cost` as a player under the grant may pay
// it: every requirement a coloured mana could pay is marked AnyMana,
// with its Options kept so the solvers still prefer the colour printed.
// A {C} (or a snow {S}, which the engine pays as colourless) is left
// alone: coloured mana cannot pay it (CR 106.1b), and widening it would
// let a Mountain pay Thought-Knot Seer's {C}. So is a symbol a
// spend-only fold left with no colour at all (spend_only.go): there is
// no colour to spend mana as though it were. Generic and {X} need no
// widening. Idempotent, and it returns `cost` itself when there is
// nothing to widen.
func widenForAnyColorSpend(cost ParsedCost) ParsedCost {
	need := false
	for _, r := range cost.Required {
		if !r.AnyMana && !requiresColorless(r) && !unpayableByAnyMana(r) {
			need = true
			break
		}
	}
	if !need {
		return cost
	}
	out := cost
	out.Required = make([]ColorRequirement, len(cost.Required))
	for i, r := range cost.Required {
		if !requiresColorless(r) && !unpayableByAnyMana(r) {
			r.AnyMana = true
		}
		out.Required[i] = r
	}
	return out
}

// costAsPaidByLocked is `cost` as `payer` may pay it for the payment
// `ctx` describes, with `x` the announced X: first the cost's own
// "spend only …" restriction folded into the symbols it allows
// (spend_only.go — it needs X, because "on X" restricts XSlots*x
// mana), then widened under a player-scoped spend grant. The argument
// itself when neither applies. THE one reading — see the file comment
// for where it is called and why there.
//
// The order is the rule: CR 609.4b's grant changes how a cost may be
// paid, and "spend only white mana" is a rule about how it may be paid,
// so the grant reaches the folded symbols exactly as it reaches a
// printed {W} (spend_only.go's file comment cites the rulings).
//
// A cost with no coloured requirement returns at once, before the
// battlefield walk: the enumerator asks this once per candidate, and
// most of what it prices is generic.
//
// Caller must hold g.mu (read or write).
func (g *Game) costAsPaidByLocked(payer uuid.UUID, ctx ManaSpendContext, cost ParsedCost, x int) ParsedCost {
	cost = cost.foldSpendOnly(x)
	if len(cost.Required) == 0 || !g.spendsManaAsAnyColorLocked(payer, ctx) {
		return cost
	}
	return widenForAnyColorSpend(cost)
}

// CostAsPaidByForEffect is costAsPaidByLocked for a caller that already
// holds g.mu — internal/legal's affordability probe, which runs inside
// ReadSnapshot and must solve the cost the payment will solve.
func (g *Game) CostAsPaidByForEffect(payer uuid.UUID, ctx ManaSpendContext, cost ParsedCost, x int) ParsedCost {
	return g.costAsPaidByLocked(payer, ctx, cost, x)
}

// CostAsPaidBy is costAsPaidByLocked under the read lock, for the
// auto-tap preview endpoint, so the plan and the missing-mana breakdown
// it shows are the ones the payment will make.
//
// Callers must NOT hold g.mu.
func (g *Game) CostAsPaidBy(payer uuid.UUID, ctx ManaSpendContext, cost ParsedCost, x int) ParsedCost {
	out := cost
	g.ReadSnapshot(func() {
		out = g.costAsPaidByLocked(payer, ctx, cost, x)
	})
	return out
}
