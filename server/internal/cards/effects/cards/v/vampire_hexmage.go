package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire Hexmage — Creature — Vampire Shaman {B}{B} 2/1 (Commander 2014):
//
//	"First strike
//	 Sacrifice this creature: Remove all counters from target permanent."
//
// The sacrifice is the cost; the target permanent is judged again as the
// ability resolves (CR 608.2b). Every kind of counter goes: loyalty,
// +1/+1, charge, poison-free permanents alike.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fcb2a95b-cef0-4b29-8b2b-538f380aff2b",
		Name:            "Vampire Hexmage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Remove all counters from target permanent.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetPermanent("target permanent"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.RemoveEveryCounterFrom(ctx, id)
			},
		}},
	})
}
