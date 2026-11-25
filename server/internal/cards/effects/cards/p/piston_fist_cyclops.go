package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Piston-Fist Cyclops — Creature — Cyclops {1}{U/R}{U/R}, 4/3 (Guilds of
// Ravnica, #5525):
//
//	"Defender
//	 As long as you've cast an instant or sorcery spell this turn, this creature
//	 can attack as though it didn't have defender."
//
// Bristlepack Sentry's waiver of defender (CR 702.3b) while a condition holds,
// the condition being Sanar's "you've cast an instant or sorcery spell this
// turn" (the controller's cast tally, read live when attackers are declared).
// The keyword stays: the Cyclops still can't attack once the condition fails.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "790e3c9e-7d94-43f7-bc14-d83c7d4c889e",
		Name:            "Piston-Fist Cyclops",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{effects.AvB18CanAttackAsThoughNoDefenderWhile(
			"As long as you've cast an instant or sorcery spell this turn, this creature can attack as though it didn't have defender.",
			func(g *game.Game, source *game.Card) bool {
				return effects.SanarCastAnInstantOrSorceryThisTurn(g, source.Controller, source.InstanceID)
			})},
	})
}
