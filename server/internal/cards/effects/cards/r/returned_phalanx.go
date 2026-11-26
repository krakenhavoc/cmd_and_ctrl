package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Returned Phalanx — Creature — Zombie Soldier {1}{B}, 3/3 (Theros):
//
//	"Defender
//	 {1}{U}: This creature can attack this turn as though it didn't have
//	 defender."
//
// Roving Keep's defender waiver (the keyword stays). Nothing happens if the
// Phalanx has left the battlefield: a new object is not "this creature"
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f8d64a8e-1bff-4960-bc36-1fbb4fc18c8f",
		Name:            "Returned Phalanx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: This creature can attack this turn as though it didn't have defender.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.WaiveDefenderUntilEOT(effects.NewContext(g, item), item.SourceCardID, "Returned Phalanx — can attack as though it didn't have defender")
			},
		}},
	})
}
