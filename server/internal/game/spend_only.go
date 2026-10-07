package game

import "strings"

// spend_only.go — #1600, ADR 0040's 2026-10-03 amendment: a COST that
// says which mana may pay it.
//
//	Throne of Eldraine  "{3}, {T}: Draw two cards. Spend only mana of the
//	                     chosen color to activate this ability."
//	Crypt Rats          "{X}: … Spend only black mana on X."
//	Crimson Hellkite    "{X}, {T}: … Spend only red mana on X."
//
// mana_restriction.go is the other direction: a restriction that rides
// the MANA ("spend this mana only to cast …") and is matched against
// what the mana pays for. This one rides the COST and is matched
// against the mana's colour. The two compose — a Throne's own mana is
// restricted to casting spells, so it cannot pay the Throne's draw
// ability whatever colour it is.
//
// # A fold into the symbols the solvers already read
//
// "Spend only white mana on {3}" pays exactly like {W}{W}{W}: three
// mana, each of which must be white. So the restriction is not taught
// to the pool solver, the auto-tapper, the plan executor and the
// missing-mana breakdown one by one — it is FOLDED into coloured
// requirements before any of them sees the cost, and each of them
// already pays a coloured requirement correctly. One fold, read at the
// one place every payment and every affordability probe already reads
// "how may this player pay this cost" (costAsPaidByLocked): the CR 602
// activation and its auto-tap, internal/legal's probe, and the auto-tap
// preview. So the view's legal_actions, the bot's moves and the payment
// cannot disagree about it (#544, and the lesson #1927 taught the
// any-colour grant).
//
// The fold needs the announced X, because "on X" restricts XSlots*x
// mana and nothing else. That is why costAsPaidByLocked takes it.
//
// The price SHOWN is untouched: the pricer stamps SpendOnly on the
// ParsedCost and String() ignores it, so the ability row still reads
// {3}. The cost is still {3} — its mana value, a cost reducer's generic
// to eat, a Trinisphere floor — and only the mana that may pay it
// changes.
//
// # Under "spend mana as though it were mana of any color"
//
// The fold runs BEFORE the any-colour widening, in the same function.
// CR 609.4b: such an effect "affects only how the player may pay a
// cost", and "spend only white mana" is a rule about how the cost may
// be paid — so a player under Chromatic Orrery may pay the Throne's
// {3} with any mana, spent as though it were the chosen colour. The
// Celestial Dawn rulings (2004-10-04) say so in as many words: white
// mana spent as though it were another colour may pay "an ability
// that can only be activated by spending another color of mana", and
// only an effect that checks the ACTUAL colour spent can tell the
// difference. Nothing here checks that, so the folded requirements are
// widened like any printed {W}: the solvers still spend real white
// first, and anything else only when no white is left.
//
// The mana-side restriction is the opposite case, and it still binds:
// "spend this mana only to cast monocolored spells of that color" is a
// restriction on the MANA, which CR 609.4b does not touch (the
// Mycosynth Lattice and Chromatic Orrery rulings). That half is a tag
// on the token (ManaRestrictMonocolored), not this.
//
// # Nothing may pay it
//
// A Throne whose colour was never chosen (a copy that entered without
// the prompt, or the moment before the controller answers) names no
// colour, and "spend only mana of the chosen color" with no chosen
// colour is a cost no mana can pay — the weaker direction, which is
// the rule every chosen-colour reader follows (#742). The fold marks
// each restricted symbol with noManaColor, which no mana has, and the
// any-colour widening leaves it alone: there is no colour to spend
// anything as though it were.

// ManaSpendOnly says which mana may pay a cost. Data, not a func: an
// AbilityCost is reachable from Game and the ADR 0041 closure ratchet
// admits no new func-typed route.
//
// Two forms. On an AbilityCost it is the DECLARATION, and may name the
// source's chosen colour (ChosenColor). On a ParsedCost it is RESOLVED:
// Colors is the whole set and ChosenColor is false. ResolveFor turns
// the first into the second.
type ManaSpendOnly struct {
	// Colors are the colours of mana that may pay, uppercase WUBRG —
	// {"B"} for "Spend only black mana on X". Empty in a resolved
	// restriction means nothing may pay.
	Colors []string `json:"colors,omitempty"`

	// ChosenColor is "of the chosen color": the colour stored on the
	// ability's source by its "as this enters, choose a color" prompt
	// (Card.ChosenColor, #742), read when the cost is priced.
	ChosenColor bool `json:"chosen_color,omitempty"`

	// XOnly scopes the restriction to the mana paid for {X} — "Spend
	// only black mana on X". Unset, it is the whole mana cost — "Spend
	// only mana of the chosen color to activate this ability".
	XOnly bool `json:"x_only,omitempty"`
}

// noManaColor is the option a folded symbol carries when no colour may
// pay it. No mana token has it, so neither solver can match it, and
// widenForAnyColorSpend leaves it alone. Not a colour anywhere else.
const noManaColor = "none"

// ResolveFor is the restriction as it applies to one activation of an
// ability printed on `source`: the declared colours plus, for "the
// chosen color", the colour stored on the source. Nil in, nil out.
//
// A copy with ChosenColor false — the resolved form — so the returned
// value can ride a ParsedCost without aliasing the catalog's
// process-lifetime declaration.
func (s *ManaSpendOnly) ResolveFor(source Card) *ManaSpendOnly {
	if s == nil {
		return nil
	}
	out := &ManaSpendOnly{XOnly: s.XOnly}
	for _, c := range s.Colors {
		out.Colors = appendColorOnce(out.Colors, strings.ToUpper(c))
	}
	if s.ChosenColor && source.ChosenColor != "" {
		out.Colors = appendColorOnce(out.Colors, strings.ToUpper(source.ChosenColor))
	}
	return out
}

// appendColorOnce appends a WUBRG colour that is not already present.
// Anything that is not one of the five colours is dropped: a resolved
// restriction names colours of mana, and colourless is not a colour
// (CR 105.1, CR 106.1b).
func appendColorOnce(xs []string, c string) []string {
	if !isColorSymbol(c) {
		return xs
	}
	for _, x := range xs {
		if x == c {
			return xs
		}
	}
	return append(xs, c)
}

// foldSpendOnly is `c` with its SpendOnly restriction folded into
// coloured requirements, for the announced X: every restricted mana
// becomes one requirement whose options are the restriction's colours.
// Returns `c` itself when it carries no restriction.
//
// The whole-cost form folds the generic component and narrows each
// coloured requirement to the colours it shares with the restriction
// (a {B} tax on a "spend only white mana" ability is a symbol nothing
// may pay). The "on X" form folds XSlots*x and nothing else. Either way
// the result has no XSlots left to multiply and no SpendOnly left to
// fold twice, so a caller that folds and then hands the cost and the
// same X to a solver gets the same answer.
//
// Pure.
func (c ParsedCost) foldSpendOnly(x int) ParsedCost {
	lim := c.SpendOnly
	if lim == nil {
		return c
	}
	if x < 0 {
		x = 0
	}
	allowed := lim.Colors
	if len(allowed) == 0 {
		allowed = []string{noManaColor}
	}
	out := c
	out.SpendOnly = nil
	n := c.XSlots * x
	out.XSlots = 0
	req := make([]ColorRequirement, 0, len(c.Required)+n+c.Generic)
	for _, r := range c.Required {
		if !lim.XOnly {
			r.Options = intersectColors(r.Options, allowed)
			if len(r.Options) == 0 {
				r.Options = []string{noManaColor}
			}
		}
		req = append(req, r)
	}
	if !lim.XOnly {
		n += c.Generic
		out.Generic = 0
	}
	for i := 0; i < n; i++ {
		req = append(req, ColorRequirement{Options: append([]string(nil), allowed...)})
	}
	out.Required = req
	return out
}

// unpayableByAnyMana reports whether a requirement is one the fold
// left with no colour that may pay it.
func unpayableByAnyMana(r ColorRequirement) bool {
	return len(r.Options) == 1 && r.Options[0] == noManaColor
}

// # The cast side (#2556)
//
// A SPELL prints the same clauses an ability does:
//
//	Drain Life    "Spend only black mana on X."
//	Soul Burn     "Spend only black and/or red mana on X."
//	Imperiosaur   "Spend only mana produced by basic lands to cast this spell."
//
// The colour half is the ability half's ManaSpendOnly, declared on the
// card (CardDef.SpendOnly), resolved against the card being cast and
// stamped onto the ParsedCost by THE cast pricer (printedCostLocked) —
// the way AbilityManaCostForTargetsForEffect stamps an ability's. From
// there nothing is spell-specific: costAsPaidByLocked folds it with the
// announced X, so the cast, its auto-tap, internal/legal's probe and
// the preview pay it identically, and the shown price stays printed.
//
// The source half is not a colour, so it is not folded into
// requirements: it is a property of the payment, carried by the
// ManaSpendContext every cast payment already threads
// (ManaSpendContext.SourceOnly). The pool solver skips a token whose
// recorded source does not qualify (allowsToken), and the auto-tapper
// leaves every permanent that could not make a qualifying mana out of
// its plan (autoTapTopUpLocked, excludeNonQualifyingSourcesLocked).

// CatalogSpellSpendOnly and CatalogSpellSpendOnlySources are the
// catalog hooks behind CardDef.SpendOnly and CardDef.SpendOnlySources.
var (
	CatalogSpellSpendOnly        func(key string) *ManaSpendOnly
	CatalogSpellSpendOnlySources func(key string) ManaSourceKinds
)

// SpellSpendOnlyFor is the card's own "spend only <colour> mana on X"
// declaration, or nil — nearly every card.
func SpellSpendOnlyFor(key string) *ManaSpendOnly {
	if CatalogSpellSpendOnly == nil || key == "" {
		return nil
	}
	return CatalogSpellSpendOnly(key)
}

// SpendOnlySourcesFor is the card's own "spend only mana produced by
// <sources> to cast this spell" declaration, or zero.
func SpendOnlySourcesFor(key string) ManaSourceKinds {
	if CatalogSpellSpendOnlySources == nil || key == "" {
		return 0
	}
	return CatalogSpellSpendOnlySources(key)
}
