package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Raging Spirit — Creature — Spirit {3}{R}, 3/3 (Mirage, S82 OLD b808):
//
//	"{2}: This creature becomes colorless until end of turn."
//
// A layer 5 colour change (CR 613.1e) to no colour, locked to this object as
// the ability resolves (CR 611.2c): a Spirit that left and came back is a new
// object and stays coloured (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "137e0ad2-e10a-4327-86a3-5ac7bc7fc469",
		Name:         "Raging Spirit",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature becomes colorless until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.UntilEndOfTurn(effects.NewContext(g, item), item.SourceCardID, nil,
					"Raging Spirit — colorless", game.SetColorsMod())
			},
		}},
	})
}
