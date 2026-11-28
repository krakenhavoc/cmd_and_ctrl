package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flooded Shoreline — Enchantment {U}{U} (Visions, S82 older sets batch 791,
// #5631):
//
//	"{U}{U}, Return two Islands you control to their owner's hand: Return
//	 target creature to its owner's hand."
//
// The two Islands are a cost, named as the ability is activated and returned
// before it goes on the stack (CR 601.2h); a board with fewer than two cannot
// activate it (CR 118.3). "Islands" is the land type, so any basic or nonbasic
// Island counts, tapped or not. The bounce is Unsummon's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01fd889f-1839-44b9-b936-0e62c8f8a730",
		Name:         "Flooded Shoreline",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{U}, Return two Islands you control to their owner's hand: Return target creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{U}{U}"), effects.ReturnNToHand(2, "two Islands you control", effects.HasSubtype("Island"))),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalCardTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				return effects.BounceToHand{Target: id}.Apply(ctx)
			},
		}},
	})
}
