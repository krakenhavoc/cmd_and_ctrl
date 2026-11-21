package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nantuko Husk — Creature — Zombie Insect {2}{B}, 2/2 (Foundations
// Jumpstart):
//
//	"Sacrifice a creature: This creature gets +2/+2 until end of turn."
//
// The Husk may sacrifice itself (Carrion Feeder's note): it is gone when
// the ability resolves and the pump has nowhere to land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0dcefc00-9425-43e3-bd47-1e34fff8b0e2",
		Name:         "Nantuko Husk",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    effects.SacrificeACreature(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 2, Label: "Nantuko Husk — +2/+2"}.Apply(ctx)
			},
		}},
	})
}
