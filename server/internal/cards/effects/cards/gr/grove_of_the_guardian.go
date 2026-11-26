package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grove of the Guardian — Land (Return to Ravnica, #5566):
//
//	"{T}: Add {C}.
//	 {3}{G}{W}, {T}, Tap two untapped creatures you control, Sacrifice this
//	 land: Create an 8/8 green and white Elemental creature token with
//	 vigilance."
//
// The {T} is the symbol on the land itself; the two creatures are others, and
// may be ones that arrived this turn (Harmonized Trio's TapOthers cost, which
// does not target). The land is sacrificed as a cost, so the token is made
// even if the land is gone by resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f746612a-fbed-44ca-b2cc-5928e10cf4bb",
		Name:         "Grove of the Guardian",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label: "{3}{G}{W}, {T}, Tap two untapped creatures you control, Sacrifice this land: Create an 8/8 green and white Elemental creature token with vigilance.",
			Cost: effects.Plus(effects.ManaCost("{3}{G}{W}"), effects.TapCost(), effects.SacrificeThis(),
				game.AbilityCost{TapOthers: &game.TapOthersCost{
					Count:         2,
					Filter:        effects.TargetPermanent("two untapped creatures you control", effects.Creature()),
					ExcludeSource: true,
					Label:         "two untapped creatures you control",
				}}),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Effect: effects.Do(effects.CreateToken{
				Template: effects.TokenCard("8/8 green and white Elemental with vigilance"), N: 1}),
		}},
	})
}
