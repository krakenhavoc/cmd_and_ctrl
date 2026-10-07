package game

import "github.com/google/uuid"

// life_for_mana.go — "for each {B} in a cost, you may pay 2 life rather
// than pay that mana" as a PLAYER's static (ADR 0131, #2531): K'rrik,
// Son of Yawgmoth for its controller.
//
// THE RULE. K'rrik does not change a cost; it changes how the cost may
// be paid (the 2019-08-23 ruling). A {B} under K'rrik pays exactly as a
// {B/P} does (CR 107.4f): one black mana, or 2 life. It reaches the
// {B} half of a hybrid symbol ({B/G}, {2/B}; CR 107.4e), never generic
// mana, never {C}, and never a requirement Crypt Rats' "Spend only
// black mana on X" folded into black (the 2019-08-23 generic-mana
// ruling). Other players' costs are untouched.
//
// WHERE IT IS READ. Like the spend-as-though-any-colour grant
// (spend_any_color.go), at the PAYMENT: costAsPaidByLocked calls
// grantLifeForManaLocked first, before the spend-only fold, so every
// payment and every affordability probe sees the grant and none of the
// pricers do. The price a player is SHOWN stays printed, and mana value
// is untouched (CR 202.3) because the flag lives on a copy made at
// payment.
//
// WHAT THE FLAG MEANS. ColorRequirement.LifeGranted marks a requirement
// whose "or 2 life" half is a grant rather than a printed symbol.
// ColorRequirement.PaysWithLife() is true for either, and it is the one
// predicate the strike (PhyrexianLifePlan), the symbol count
// (PhyrexianSymbols) and the claim gate read, so the life half of a
// symbol has one owner.
//
// AUTO-TAP NEVER PAYS LIFE. The life is paid only when the announcement
// claims it (CastSpellParams.PhyrexianLife and its siblings, CR 601.2b),
// exactly as for a printed Phyrexian symbol. A mono-black K'rrik deck
// would otherwise lose life on almost every spell nobody chose to pay
// for. Nothing here touches the auto-tapper.

// LifeForManaStatic is one printed "for each {Color} in a cost, you may
// pay 2 life rather than pay that mana" static on a permanent. Catalog
// data, never stored: read live off the battlefield, keyed by
// CatalogAbilityKey, so two K'rriks grant the same colour once, one
// leaving cannot revoke the other's grant, and a K'rrik that has lost
// its abilities grants nothing.
type LifeForManaStatic struct {
	// Label is the clause as printed.
	Label string
	// Color is the colour symbol the grant reaches: "B" for K'rrik.
	Color string
}

// CatalogLifeForMana returns the life-for-mana statics a permanent with
// the given catalog key has. carddef.go sets it from CardDef.LifeForMana;
// a game-package test may stub it.
var CatalogLifeForMana func(key string) []LifeForManaStatic

// PaysWithLife reports whether one slot carries the CR 107.4f "or 2
// life" half: a printed Phyrexian symbol, or a symbol a life-for-mana
// grant marked at payment (LifeGranted).
func (r ColorRequirement) PaysWithLife() bool {
	return r.Phyrexian || r.LifeGranted
}

// paysLifeForColorsLocked returns the colours `payer` may pay 2 life
// for, each once. Nil when nothing grants it, which is nearly always.
//
// Caller must hold g.mu (read or write). Reads only.
func (g *Game) paysLifeForColorsLocked(payer uuid.UUID) []string {
	if CatalogLifeForMana == nil || g.Battlefield == nil || payer == uuid.Nil {
		return nil
	}
	var out []string
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		if src.Controller != payer {
			continue
		}
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, s := range CatalogLifeForMana(key) {
			out = appendColorOnce(out, s.Color)
		}
	}
	return out
}

// grantLifeForManaLocked is `cost` as `payer` may pay it under a
// life-for-mana grant: every requirement whose Options include a granted
// colour, and that does not already carry the life half, is marked
// LifeGranted. Options, String and the mana value are unchanged.
//
// Idempotent, and it returns `cost` itself when nothing is granted or
// nothing needs the flag. A requirement the spend-only fold left with no
// colour and a {C} are never marked: they hold no granted colour.
//
// Caller must hold g.mu (read or write).
func (g *Game) grantLifeForManaLocked(payer uuid.UUID, cost ParsedCost) ParsedCost {
	// Cheap negative first: the enumerator asks this once per candidate
	// and most costs print no coloured symbol the grant could reach.
	need := false
	for _, r := range cost.Required {
		if !r.PaysWithLife() && len(r.Options) > 0 && !requiresColorless(r) {
			need = true
			break
		}
	}
	if !need {
		return cost
	}
	colors := g.paysLifeForColorsLocked(payer)
	if len(colors) == 0 {
		return cost
	}
	out := cost
	copied := false
	for i, r := range cost.Required {
		if r.PaysWithLife() || !optionsIntersect(r.Options, colors) {
			continue
		}
		if !copied {
			out.Required = append([]ColorRequirement(nil), cost.Required...)
			copied = true
		}
		r.LifeGranted = true
		out.Required[i] = r
	}
	return out
}

// optionsIntersect reports whether any of `options` is one of `colors`.
func optionsIntersect(options, colors []string) bool {
	for _, o := range options {
		for _, c := range colors {
			if o == c {
				return true
			}
		}
	}
	return false
}

// LifeGrantedCostForEffect is grantLifeForManaLocked for a caller that
// already holds g.mu — internal/legal's life loops and the view's
// symbol count, which must count the symbols the payment will offer
// (ADR 0131 §1).
func (g *Game) LifeGrantedCostForEffect(payer uuid.UUID, cost ParsedCost) ParsedCost {
	return g.grantLifeForManaLocked(payer, cost)
}

// cardNameLocked is the name of the card `id` names, in whatever zone
// holds it, or "the cost" when it is nowhere. For an error message only
// (strikePhyrexianLifeLocked's source name). Caller must hold g.mu.
func (g *Game) cardNameLocked(id uuid.UUID) string {
	if c, _ := g.findCardAndZoneLocked(id); c != nil && c.Name != "" {
		return c.Name
	}
	return "the cost"
}

// LifeGrantedSymbols is how many of `c`'s symbols can be paid with life
// only because of a grant, not because they print a Phyrexian symbol.
// The view's `phyrexian_granted`. Pure.
func (c ParsedCost) LifeGrantedSymbols() int {
	n := 0
	for _, r := range c.Required {
		if r.LifeGranted && !r.Phyrexian {
			n++
		}
	}
	return n
}
