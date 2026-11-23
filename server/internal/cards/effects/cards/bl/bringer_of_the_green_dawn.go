package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bringer of the Green Dawn — Creature — Bringer {7}{G}{G}, 5/5 (Fifth Dawn, #5604):
//
//	"You may pay {W}{U}{B}{R}{G} rather than pay this spell's mana
//	 cost.
//	 Trample
//	 At the beginning of your upkeep, you may create a 3/3 green Beast
//	 creature token."
//
// The shared Bringer alternative cost (bringer_alternative_cost.go), and an
// optional upkeep trigger that creates the token table's 3/3 green Beast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "f0548af1-d1cc-44d1-b5a0-f013563f29de",
		Name:             "Bringer of the Green Dawn",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{effects.BringerAlternativeCost()},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Bringer of the Green Dawn — create a 3/3 green Beast",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: 1})),
				"Bringer of the Green Dawn — create a 3/3 green Beast creature token?"),
		},
	})
}
