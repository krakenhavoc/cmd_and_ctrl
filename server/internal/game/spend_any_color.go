package game

import (
	"slices"

	"github.com/google/uuid"
)

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
	// To narrows a From grant to symbols of ONE colour (an uppercase
	// WUBRG letter): "You may spend white mana as though it were RED
	// mana" (Sunglasses of Urza, #6202). The From colour then pays a
	// symbol only when that symbol admits To — a {R}, an {R/G} hybrid's
	// red half — and never a {U}. Empty is "as though it were mana of any
	// color". Meaningless without From.
	To string
	// OthersOnlyColorless is the second sentence of Celestial Dawn
	// (#6202): "You may spend other mana only as though it were
	// colorless mana." It is a restriction, not a grant: mana of every
	// colour but From pays only generic mana and {C} symbols, and a
	// coloured symbol admits only From. Colourless mana is unaffected
	// (it already is colourless). CR 609.4b: the costs and the mana are
	// unchanged; only how the player may pay is.
	OthersOnlyColorless bool
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
	// stand for another colour (spendPair.to) or any colour.
	from []spendPair
	// colorlessOnly is a Celestial Dawn restriction (OthersOnlyColorless)
	// covering the payment: mana of a colour not in exempt may be spent
	// only as though it were colourless.
	colorlessOnly bool
	exempt        []string
}

// spendPair is one one-colour grant: mana of colour `from` may be spent
// as though it were mana of colour `to`, or of any colour when `to` is
// empty.
type spendPair struct {
	from, to string
}

// note folds one covering grant in.
func (s *spendGrant) note(anyType bool, from string) {
	s.noteTo(anyType, from, "")
}

// noteTo is note for a grant narrowed to one target colour (To).
func (s *spendGrant) noteTo(anyType bool, from, to string) {
	if from == "" {
		s.any = true
		s.anyType = s.anyType || anyType
		return
	}
	for _, f := range s.from {
		if f.from == from && f.to == to {
			return
		}
	}
	s.from = append(s.from, spendPair{from: from, to: to})
}

// restrict folds one "other mana only as though it were colorless"
// restriction in: mana of colour `exempt` is spared. Two restrictions
// both bind, so only a colour both spare is spared.
func (s *spendGrant) restrict(exempt string) {
	if !s.colorlessOnly {
		s.colorlessOnly = true
		if exempt != "" {
			s.exempt = []string{exempt}
		}
		return
	}
	var kept []string
	for _, e := range s.exempt {
		if e == exempt {
			kept = append(kept, e)
		}
	}
	s.exempt = kept
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
			if s.OthersOnlyColorless {
				out.restrict(s.From)
			}
			out.noteTo(s.AnyType, s.From, s.To)
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
// A {C} is left alone: coloured mana cannot pay it (CR 106.1b), and
// widening it would let a Mountain pay Thought-Knot Seer's {C}. So is a
// snow {S}, even under an any-type grant: only mana from a snow source
// pays it (CR 107.4h). So is a symbol a
// spend-only fold left with no colour at all (spend_only.go): there is
// no colour to spend mana as though it were. Generic and {X} need no
// widening. Idempotent, and it returns `cost` itself when there is
// nothing to widen.
func widenForAnyColorSpend(cost ParsedCost, anyType bool) ParsedCost {
	colourless := func(r ColorRequirement) bool { return !anyType && requiresColorless(r) }
	// #5986: never a snow {S}, whose payer is a snow source and not a
	// colour or type (CR 107.4h).
	widens := func(r ColorRequirement) bool { return !r.Snow && !colourless(r) && !unpayableByAnyMana(r) }
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
	return g.costAsPaidByExcludingLocked(payer, ctx, cost, x, nil)
}

// costAsPaidByExcludingLocked is costAsPaidByLocked for a payer whose
// auto-tap may not reach for `excluded` (the cast's lock-taps and the
// sources its other costs spend, castAutoTapExclusionsLocked). The
// exclusions matter only to a "no more than one mana of each color"
// clause, whose colours are chosen by what the payer can pay
// (spend_only.go, "One mana of each colour"); every other cost reads
// exactly as costAsPaidByLocked reads it.
//
// Caller must hold g.mu (read or write).
func (g *Game) costAsPaidByExcludingLocked(payer uuid.UUID, ctx ManaSpendContext, cost ParsedCost, x int, excluded map[uuid.UUID]bool) ParsedCost {
	// ADR 0131: K'rrik's "you may pay 2 life rather than pay that {B}"
	// marks the printed symbols FIRST, so the requirements the
	// spend-only fold adds below (Crypt Rats' "black mana on X") are not
	// marked: life never pays generic mana (the 2019-08-23 ruling).
	cost = g.grantLifeForManaLocked(payer, cost)
	if choices := cost.distinctSpendChoices(x); len(choices) > 0 {
		return g.distinctColorsAsPaidLocked(payer, ctx, cost, x, excluded, choices)
	}
	return g.foldAndWidenLocked(payer, ctx, cost, x)
}

// distinctColorsAsPaidLocked is the cost as paid with the colour subset
// of a "no more than one mana of each color" clause chosen (#2575): the
// first of `choices` the floating pool pays alone, else the first the
// pool and an auto-tap plan pay together, else the first. Read-only.
//
// Caller must hold g.mu (read or write).
func (g *Game) distinctColorsAsPaidLocked(payer uuid.UUID, ctx ManaSpendContext, cost ParsedCost, x int, excluded map[uuid.UUID]bool, choices [][]string) ParsedCost {
	paid := make([]ParsedCost, len(choices))
	for i, colors := range choices {
		paid[i] = g.foldAndWidenLocked(payer, ctx, cost.withSpendColors(colors), x)
	}
	p := g.playerByIDLocked(payer)
	if p == nil {
		return paid[0]
	}
	for _, c := range paid {
		if p.ManaPool.CanPayFor(c, x, ctx) {
			return c
		}
	}
	for _, c := range paid {
		if _, _, ok := g.autoTapTopUpLocked(payer, c, x, ctx, excluded, 0); ok {
			return c
		}
	}
	return paid[0]
}

// foldAndWidenLocked is the rest of costAsPaidByLocked: the spend-only
// fold, then the spend grant's widening.
//
// Caller must hold g.mu (read or write).
func (g *Game) foldAndWidenLocked(payer uuid.UUID, ctx ManaSpendContext, cost ParsedCost, x int) ParsedCost {
	cost = cost.foldSpendOnly(x)
	if cost.requiredCount() == 0 {
		return cost
	}
	grant := g.spendGrantLocked(payer, ctx)
	switch {
	case grant.any:
		cost = widenForAnyColorSpend(cost, grant.anyType)
	case len(grant.from) > 0:
		cost = widenFromPairs(cost, grant.from)
	}
	if grant.colorlessOnly {
		cost = restrictToColorless(cost, grant.exempt)
	}
	return cost
}

// restrictToColorless is `cost` as a player under "you may spend other
// mana only as though it were colorless mana" (Celestial Dawn, #6202)
// may pay it: mana of a colour not in `exempt` pays only what
// colourless mana pays, which is generic mana (never a requirement)
// and a {C}. So a slot that admits {C} admits every non-exempt colour
// too, and a slot that admits a colour (or was widened to any mana)
// admits only the exempt colours. Snow {S} is paid by its source, not
// a colour (CR 107.4h), and is left alone, as is a slot no mana pays.
// Applied after the widening, so the grant the same card prints
// ("white mana as though it were mana of any color") is what lets
// white pay a {R}.
func restrictToColorless(cost ParsedCost, exempt []string) ParsedCost {
	isExempt := func(c string) bool {
		for _, e := range exempt {
			if e == c {
				return true
			}
		}
		return false
	}
	restrict := func(r ColorRequirement) ColorRequirement {
		if r.Snow || len(r.Options) == 0 || unpayableByAnyMana(r) {
			return r
		}
		colourless, coloured := r.AnyMana, r.AnyMana
		for _, o := range r.Options {
			if o == "C" {
				colourless = true
			} else {
				coloured = true
			}
		}
		var opts []string
		if coloured {
			opts = append(opts, exempt...)
		}
		if colourless {
			opts = append(opts, "C")
			for _, c := range []string{"W", "U", "B", "R", "G"} {
				if !isExempt(c) {
					opts = append(opts, c)
				}
			}
		}
		if len(opts) == 0 {
			opts = []string{noManaColor}
		}
		r.Options = opts
		r.AnyMana = false
		return r
	}
	out := cost
	out.Required = make([]ColorRequirement, len(cost.Required))
	for i, r := range cost.Required {
		out.Required[i] = restrict(r)
	}
	if out.folded.n > 0 {
		out.folded.req = restrict(out.folded.req)
	}
	return out
}

// widenFromColors is widenForAnyColorSpend for a grant narrowed to some
// colours of mana (AnyColorSpendStatic.From): each coloured symbol also
// admits mana of those colours. The symbol keeps its own colours first,
// so the solvers still pay a {G} with green when there is green. A {C}
// is left alone for widenForAnyColorSpend's reason (blue mana is not
// colourless), and so is a symbol already widened to any mana. Returns
// `cost` itself when nothing changes.
func widenFromColors(cost ParsedCost, from []string) ParsedCost {
	pairs := make([]spendPair, len(from))
	for i, f := range from {
		pairs[i] = spendPair{from: f}
	}
	return widenFromPairs(cost, pairs)
}

// widenFromPairs is widenFromColors with each colour's target: a pair
// with a `to` widens only the symbols that admit `to` (Sunglasses of
// Urza's white-as-red pays a {R} and an {R/G}, never a {U}).
func widenFromPairs(cost ParsedCost, from []spendPair) ParsedCost {
	// widen is r admitting `from` too, and whether that changed it.
	widen := func(r ColorRequirement) (ColorRequirement, bool) {
		if r.AnyMana || len(r.Options) == 0 || requiresColorless(r) || unpayableByAnyMana(r) {
			return r, false
		}
		var add []string
		for _, f := range from {
			if f.to != "" && !matchColor(f.to, r.Options) {
				continue
			}
			if !matchColor(f.from, r.Options) && !slices.Contains(add, f.from) {
				add = append(add, f.from)
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

// widenSnowAny is widenFromColors for a grant narrowed by the mana's
// source rather than its colour (CastPermission.AnyColorFromSnow, #6161):
// each coloured symbol also takes any mana a snow source produced
// (ColorRequirement.SnowAny). A {C} is left alone (colourless is not a
// colour, CR 106.1b), and so is an {S} (snow is neither a colour nor a
// type, CR 107.4h) and a symbol already widened to any mana. Returns
// `cost` itself when nothing changes.
func widenSnowAny(cost ParsedCost) ParsedCost {
	widen := func(r ColorRequirement) (ColorRequirement, bool) {
		if r.AnyMana || r.Snow || r.SnowAny || len(r.Options) == 0 || requiresColorless(r) || unpayableByAnyMana(r) {
			return r, false
		}
		r.SnowAny = true
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

// CostAsPaidByExcludingForEffect is costAsPaidByExcludingLocked for
// the same caller, which knows the sources the payment's auto-tap may
// not reach for (canPayExcluding), so a "no more than one mana of each
// color" clause is read with the exclusions the payment reads it with.
func (g *Game) CostAsPaidByExcludingForEffect(payer uuid.UUID, ctx ManaSpendContext, cost ParsedCost, x int, excluded map[uuid.UUID]bool) ParsedCost {
	return g.costAsPaidByExcludingLocked(payer, ctx, cost, x, excluded)
}

// CostAsPaidBy is costAsPaidByLocked under the read lock, for the
// auto-tap preview endpoint, so the plan and the missing-mana breakdown
// it shows are the ones the payment will make.
//
// Callers must NOT hold g.mu.
func (g *Game) CostAsPaidBy(payer uuid.UUID, ctx ManaSpendContext, cost ParsedCost, x int) ParsedCost {
	return g.CostAsPaidByExcluding(payer, ctx, cost, x, nil)
}

// CostAsPaidByExcluding is CostAsPaidBy with the preview's auto-tap
// exclusions, the ones CastSpell's auto-tap will plan around.
//
// Callers must NOT hold g.mu.
func (g *Game) CostAsPaidByExcluding(payer uuid.UUID, ctx ManaSpendContext, cost ParsedCost, x int, excluded map[uuid.UUID]bool) ParsedCost {
	out := cost
	g.ReadSnapshot(func() {
		out = g.costAsPaidByExcludingLocked(payer, ctx, cost, x, excluded)
	})
	return out
}
