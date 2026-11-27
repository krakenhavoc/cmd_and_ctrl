package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// King Crab — Creature — Crab {4}{U}{U}, 4/5 (Urza's Legacy, S82 OLD b716, #5623):
//
//	"{1}{U}, {T}: Put target green creature on top of its owner's library."
//
// Roil Spout's tuck on a green creature (colour read through the layers, at
// announce and again as the ability resolves, CR 608.2b), the card going on
// top of its OWNER's library. A tap ability, so summoning sickness applies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24982bfe-5276-44a8-99a5-43b497ebf51f",
		Name:         "King Crab",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: Put target green creature on top of its owner's library.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target green creature", effects.OfColor("G")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if id := effects.FirstLegalBattlefieldTarget(effects.NewContext(g, item)); id != uuid.Nil {
					return g.TuckToLibraryForEffect(id, false)
				}
				return nil
			},
		}},
	})
}
