package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drowsing Tyrannodon — Creature — Dinosaur {1}{G}, 3/3:
//
//	"Defender
//	 As long as you control a creature with power 4 or greater, this
//	 creature can attack as though it didn't have defender."
//
// Bristlepack Sentry's waiver, read at declaration.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ee811356-ef6a-4972-b827-32c2d329b3c3",
		Name:            "Drowsing Tyrannodon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{effects.AvB18CanAttackAsThoughNoDefenderWhile(
			"As long as you control a creature with power 4 or greater, this creature can attack as though it didn't have defender.",
			func(g *game.Game, source *game.Card) bool {
				return effects.B24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
			})},
	})
}
