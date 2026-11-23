package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shambling Shell — Creature — Plant Zombie {1}{B}{G}, 3/1:
//
//	"Sacrifice this creature: Put a +1/+1 counter on target creature.
//	 Dredge 3 (If you would draw a card, you may mill three cards
//	 instead. If you do, return this card from your graveyard to your
//	 hand.)"
//
// The sacrifice is the cost, so the Shell is in the graveyard when the
// counter is placed, and dredge can bring it back the same turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c50f6b1-8850-4b08-a994-3ccc17f25b83",
		Name:         "Shambling Shell",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Dredge(3)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Put a +1/+1 counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return effects.AddCounter{Target: t.ID, Kind: "+1/+1", N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
