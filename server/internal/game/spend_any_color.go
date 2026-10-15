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
// WIDENED, NOT FOLDED. A cast permission's grant used to fold a coloured
// symbol into the generic demand (#1928 made it widen too, ADR 0066's
// 2026-10-06 amendment). A fold throws away which colour the card printed,
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
// mana to pay. Before the Phyrexian strike, as the permission widening is,
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
	// AnyType makes it "as though it were mana of any TYPE" (Vizier of
	// the Menagerie): colourless is a type too, so coloured mana may also
	// pay a {C} symbol (CR 106.1b no longer stops it). False is the
	// colour grant, which leaves {C} alone.
	AnyType bool
	// From narrows the grant to mana of ONE colour (an uppercase WUBRG
	// letter): "You may spend BLUE mana as though it were mana of any
	// color" (Quicksilver Elemental; ADR 0155 Decision 7). Each coloured
	// symbol the payment reaches then admits that colour beside its own,
	// and no other mana. Empty is every colour of mana, which is every
	// grant written before it.
	From string
	// OwnAbilities narrows the grant to the activation costs of THIS
	// permanent's abilities — "to pay the activation costs of this
	// creature's abilities" (Quicksilver Elemental): a payment whose
	// ManaSpendContext.Object is the permanent the static is on, its own
	// rows and every row it was lent alike. Composes with Covers.
	OwnAbilities bool
}

// CatalogAnyColorSpend returns the spend statics a permanent with the
// given catalog key has. carddef.go sets it from CardDef.AnyColorSpend;
// a game-package test may stub it.
var CatalogAnyColorSpend func(key string) []AnyColorSpendStatic

// spendsManaAsAnyColorLocked is THE reader: may `payer` spend mana as
// though it were mana of any color for the payment `ctx` describes?
//
// A grant narrowed to one colour of mana (From) is not that, and answers
// false here; costAsPaidByLocked still applies it.
//
// Caller must hold g.mu (read or write). Reads only.
func (g *Game) spendsManaAsAnyColorLocked(payer uuid.UUID, ctx ManaSpendContext) bool {
	return g.spendGrantLocked(payer, ctx).any
}

// spendGrant is what the spend grants covering one payment allow.
type spendGrant struct {
	// any: every mana may be spent as though it were mana of any color;
	// anyType: of any type, so a {C} too.
	any, anyType bool
	// from lists the colours of mana a one-colour grant (From) lets
	// stand for any colour.
	from []string
}

// note folds one covering grant in.
func (s *spendGrant) note(anyType bool, from string) {
	if from == "" {
		s.any = true
		s.anyType = s.anyType || anyType
		return
	}
	for _, f := range s.from {
		if f == from {
			return
		}
	}
	s.from = append(s.from, from)
}

// spendGrantLocked is the reader behind spendsManaAsAnyColorLocked and
// costAsPaidByLocked: every grant covering the payment, folded.
//
// Two kinds of grant are read. A permanent's printed static
// (CatalogAnyColorSpend), live off the battlefield; and the rider a
// RESOLVED borrow wrote on the rows it lent (BorrowedAbility.SpendFrom,
// Grell Philosopher's "you may spend blue mana as though it were mana of
// any color to activate those abilities", ADR 0155 Decision 7), which
// lasts exactly as long as those rows do, whatever happens to the
// permanent whose trigger lent them.
func (g *Game) spendGrantLocked(payer uuid.UUID, ctx ManaSpendContext) spendGrant {
	var out spendGrant
	if payer == uuid.Nil {
		return out
	}
	if o := ctx.AbilityOrigin; o.Borrowed && o.Lent != nil && o.Lent.SpendFrom != "" && o.Lent.Activator == payer {
		out.note(false, o.Lent.SpendFrom)
	}
	if CatalogAnyColorSpend == nil || g.Battlefield == nil {
		return out
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
			if s.OwnAbilities && (ctx.Object == uuid.Nil || ctx.Object != src.InstanceID) {
				continue
			}
			if s.Covers != nil && !s.Covers(withPayer(ctx, payer)) {
				continue
			}
			out.note(s.AnyType, s.From)
			if out.anyType {
				return out
			}
		}
	}
	return out
}

// withPayer is ctx with its Payer filled in, for a Covers that compares
// the object's controller with the player paying.
func withPayer(ctx ManaSpendContext, payer uuid.UUID) ManaSpendContext {
	ctx.Payer = payer
	return ctx
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
func widenForAnyColorSpend(cost ParsedCost, anyType bool) ParsedCost {
	colourless := func(r ColorRequirement) bool { return !anyType && requiresColorless(r) }
	widens := func(r ColorRequirement) bool { return !colourless(r) && !unpayableByAnyMana(r) }
	need := cost.folded.n > 0 && !cost.folded.req.AnyMana && widens(cost.folded.req)
	for _, r := range cost.Required {
		if !r.AnyMana && widens(r) {
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
		if widens(r) {
			r.AnyMana = true
		}
		out.Required[i] = r
	}
	// The folded run is identical requirements, so it is widened
	// whole, as each of them was (#3909).
	if out.folded.n > 0 && widens(out.folded.req) {
		out.folded.req.AnyMana = true
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
// Before both, a life-for-mana grant marks the symbols it reaches
// (life_for_mana.go).
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
	// ADR 0131: K'rrik's "you may pay 2 life rather than pay that {B}"
	// marks the printed symbols FIRST, so the requirements the
	// spend-only fold adds below (Crypt Rats' "black mana on X") are not
	// marked: life never pays generic mana (the 2019-08-23 ruling).
	cost = g.grantLifeForManaLocked(payer, cost)
	cost = cost.foldSpendOnly(x)
	if cost.requiredCount() == 0 {
		return cost
	}
	grant := g.spendGrantLocked(payer, ctx)
	switch {
	case grant.any:
		return widenForAnyColorSpend(cost, grant.anyType)
	case len(grant.from) > 0:
		return widenFromColors(cost, grant.from)
	}
	return cost
}

// widenFromColors is widenForAnyColorSpend for a grant narrowed to some
// colours of mana (AnyColorSpendStatic.From): each coloured symbol also
// admits mana of those colours. The symbol keeps its own colours first,
// so the solvers still pay a {G} with green when there is green. A {C}
// is left alone for widenForAnyColorSpend's reason (blue mana is not
// colourless), and so is a symbol already widened to any mana. Returns
// `cost` itself when nothing changes.
func widenFromColors(cost ParsedCost, from []string) ParsedCost {
	// widen is r admitting `from` too, and whether that changed it.
	widen := func(r ColorRequirement) (ColorRequirement, bool) {
		if r.AnyMana || len(r.Options) == 0 || requiresColorless(r) || unpayableByAnyMana(r) {
			return r, false
		}
		var add []string
		for _, f := range from {
			if !matchColor(f, r.Options) {
				add = append(add, f)
			}
		}
		if len(add) == 0 {
			return r, false
		}
		r.Options = append(append([]string(nil), r.Options...), add...)
		return r, true
	}
	var out []ColorRequirement
	for i, r := range cost.Required {
		r, changed := widen(r)
		if !changed {
			continue
		}
		if out == nil {
			out = append([]ColorRequirement(nil), cost.Required...)
		}
		out[i] = r
	}
	// The folded run is widened whole, as each of its identical
	// requirements was (#3909).
	var tail ColorRequirement
	tailChanged := false
	if cost.folded.n > 0 {
		tail, tailChanged = widen(cost.folded.req)
	}
	if out == nil && !tailChanged {
		return cost
	}
	widened := cost
	if out != nil {
		widened.Required = out
	}
	if tailChanged {
		widened.folded.req = tail
	}
	return widened
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
