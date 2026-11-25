package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// sacrifice_up_to_cost.go — #6121: an activated ability's sacrifice
// clause whose floor is zero.
//
//	Baba Lysaga, Night Witch   "{T}, Sacrifice up to three permanents: …"
//	Sword of the Ages          "{T}, Sacrifice this artifact and any number
//	                           of creatures you control: …"
//
// Both are the clause #1213 and ADR 0100 §3 already pay, at bounds that
// were refused on an ability because a cost that can be paid with nothing
// is free. Neither card's is: each prints a {T} beside it (CR 602.2b
// through 601.2b: the activator announces how many, zero included, and
// pays the {T} whatever the count). Register accepts the zero floor on an
// ability only beside such a component (AbilityCostPaysBesideSacrifice).
//
// The effect reads what was paid back off the stack item, never off the
// board: ctx.Sacrificed() for the count and SacrificedPermanents for the
// objects as they last existed on the battlefield (CR 608.2h).

// SacrificeUpTo is "Sacrifice up to N <permanents>": the activator names
// any number from zero to n. Combine it with the component that keeps the
// ability from being free, e.g.
//
//	Plus(TapCost(), SacrificeUpTo(3, "up to three permanents", Permanent()))
func SacrificeUpTo(n int, label string, preds ...CardPredicate) game.AbilityCost {
	return game.AbilityCost{SacrificeOther: SacrificeSpec(label, preds...).WithCount(0, n)}
}

// SacrificeAnyNumberOf is "Sacrifice any number of <permanents>" on an
// activated ability: no floor and no ceiling but the board. Sword of the
// Ages' "Sacrifice this artifact and any number of creatures you control"
// is
//
//	Plus(TapCost(), SacrificeThis(), SacrificeAnyNumberOf("any number of creatures you control", Creature()))
func SacrificeAnyNumberOf(label string, preds ...CardPredicate) game.AbilityCost {
	return game.AbilityCost{SacrificeOther: SacrificeSpec(label, preds...).WithCount(0, 0)}
}

// AbilityCostPaysBesideSacrifice reports whether an ability's cost has a
// component other than its sacrifice clause that always costs something —
// the {T} or {Q} symbol, sacrificing the source, or a mana cost — so a
// sacrifice clause with a floor of zero does not make the ability free
// (#6121). Register asks it before allowing that floor on an ability.
func AbilityCostPaysBesideSacrifice(c game.AbilityCost) bool {
	return c.Tap || c.UntapSelf || c.SacrificeSelf || c.Mana != ""
}
