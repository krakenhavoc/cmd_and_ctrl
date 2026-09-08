package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// activated.go — S21 sub-PR 2: cost constructors for
// Spec.Activated, so a card file reads like its oracle text:
//
//	Cost: SacrificeACreature(),      // "Sacrifice a creature:"
//	Cost: TapCost(),                 // "{T}:"
//	Cost: SacrificeThis(),           // "Sacrifice this permanent:"
//	Cost: ManaCost("{1}{B}"),        // "{1}{B}:"
//
// Costs compose with Plus for the multi-part case ("{2}, {T}:").

// TapCost is "{T}" — the source must be untapped, and a creature
// source must be free of summoning sickness (CR 302.1).
func TapCost() game.AbilityCost { return game.AbilityCost{Tap: true} }

// SacrificeThis sacrifices the source as the cost.
func SacrificeThis() game.AbilityCost { return game.AbilityCost{SacrificeSelf: true} }

// SacrificeACreature is the sac-outlet cost: "Sacrifice a
// creature". The source itself qualifies when it's a creature —
// Carrion Feeder really can eat itself.
func SacrificeACreature() game.AbilityCost {
	return game.AbilityCost{SacrificeOther: sacrificeSpec("a creature", Creature())}
}

// SacrificeAPermanent is "Sacrifice a permanent" (Altar of
// Dementia's broader cousin).
func SacrificeAPermanent() game.AbilityCost {
	return game.AbilityCost{SacrificeOther: sacrificeSpec("a permanent")}
}

// ManaCost is a printed mana component, "{1}{B}".
func ManaCost(cost string) game.AbilityCost { return game.AbilityCost{Mana: cost} }

// PayLife is a life component (CR 118.8).
func PayLife(n int) game.AbilityCost { return game.AbilityCost{Life: n} }

// Plus merges cost components: Plus(ManaCost("{2}"), TapCost()) is
// "{2}, {T}". Later components win for scalar fields, which only
// matters if a caller passes two mana strings (they shouldn't).
func Plus(costs ...game.AbilityCost) game.AbilityCost {
	var out game.AbilityCost
	for _, c := range costs {
		if c.Tap {
			out.Tap = true
		}
		if c.SacrificeSelf {
			out.SacrificeSelf = true
		}
		if c.SacrificeOther != nil {
			out.SacrificeOther = c.SacrificeOther
		}
		if c.Mana != "" {
			out.Mana = c.Mana
		}
		if c.Life != 0 {
			out.Life = c.Life
		}
	}
	return out
}

// sacrificeSpec builds the "what may I sacrifice" clause. It reuses
// TargetSpec because the shape is identical — a predicate over
// battlefield cards — but it is NOT targeting: a sacrifice cost
// doesn't target, so hexproof and "can't be the target of" never
// apply to it. The controller filter lives in the engine's cost
// validation (CR 701.17b).
func sacrificeSpec(label string, preds ...CardPredicate) *game.TargetSpec {
	return TargetPermanent(label, preds...)
}
