package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Expedite — Instant {R}:
//
//	"Target creature gains haste until end of turn.
//	 Draw a card."
//
// The haste grant is Act of Treason's own until-end-of-turn clause,
// GrantKeywordUntilEOT, pinned to the chosen target instead of a
// creature just stolen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3501a839-eef5-44e4-8637-b5754780454e",
		Name:         "Expedite",
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste"))},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (effects.GrantKeywordUntilEOT{
				Target:   item.Targets[0].ID,
				Keywords: []string{"haste"},
				Label:    "Expedite — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
