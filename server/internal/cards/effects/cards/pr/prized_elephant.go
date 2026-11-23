package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prized Elephant — Creature — Elephant {3}{W}, 3/3 (Magic 2013, #5568):
//
//	"This creature gets +1/+1 as long as you control a Forest.
//	 {G}: This creature gains trample until end of turn. (It can deal
//	 excess combat damage to the player or planeswalker it's attacking.)"
//
// Kird Ape's land-type pump and Thornling's trample for mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b1674b1-2176-493f-a463-936e715d51a2",
		Name:         "Prized Elephant",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control a Forest.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfWhileYouControlA(effects.Subtype("Forest")),
			Apply:     effects.PumpBy(1, 1),
		}},
		Activated: []effects.ActivatedAbility{effects.ConB1ThisGainsRow("{G}", "trample", "Prized Elephant", game.AnswerCombatGrant)},
	})
}
