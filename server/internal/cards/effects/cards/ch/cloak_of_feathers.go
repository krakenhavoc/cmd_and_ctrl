package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloak of Feathers — Sorcery {U} (Portal, #5630):
//
//	"Target creature gains flying until end of turn.
//	Draw a card."
//
// Leap's text, word for word.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb4baf53-51ed-468b-a468-5d7d45a6dc26",
		Name:         "Cloak of Feathers",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying"))},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (effects.GrantKeywordUntilEOT{
				Target:   item.Targets[0].ID,
				Keywords: []string{"flying"},
				Label:    "Cloak of Feathers — flying until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
