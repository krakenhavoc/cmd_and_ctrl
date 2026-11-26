package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mistcaller — Creature — Merfolk Wizard {U}, 1/1 (Core Set 2019, S82 OLD batch 199, #5527):
//
//	"Sacrifice this creature: Until end of turn, if a nontoken creature would
//	 enter and it wasn't cast, exile it instead."
//
// Containment Priest's replacement (a nontoken creature moving onto the
// battlefield from anywhere but the stack is exiled instead, CR 614.1, 614.6)
// made timed: a ModExileUncastCreatures scoped record created by the
// resolving ability, so it outlives the sacrificed Mistcaller and ends in the
// cleanup step. It applies to every player's creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "808d5a67-8c38-42f5-9413-8771d8b4ae38",
		Name:         "Mistcaller",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Until end of turn, if a nontoken creature would enter and it wasn't cast, exile it instead.",
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Cost:    effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ctx.Game.ExileUncastCreaturesForEffect(ctx.Source(), ctx.Controller(), effects.DurationUntilEndOfTurn(ctx),
					"Mistcaller — nontoken creatures that weren't cast are exiled instead of entering")
				return nil
			},
		}},
	})
}
