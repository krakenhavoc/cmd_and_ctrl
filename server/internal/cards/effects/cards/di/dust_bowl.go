package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dust Bowl — Land:
//
//	"{T}: Add {C}.
//	 {3}, {T}, Sacrifice a land: Destroy target nonbasic land."
//
// Wasteland with the sacrifice moved off the source: any land you
// control pays it, this one included. The sacrifice is a cost, so it is
// gone before the ability resolves; the target is chosen at announce and
// can't be the land sacrificed for it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "d3df7128-31dd-4d71-90be-87e2e9ff51b4",
		Name:          "Dust Bowl",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.ColorlessTapRow()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice a land: Destroy target nonbasic land.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.B08SacrificeALand()),
			Targets: effects.TargetPermanent("target nonbasic land", effects.Land(), effects.B10NonbasicLand()),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
