package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Show of Confidence — Instant {1}{W} (Strixhaven: School of Mages,
// #5502):
//
//	"When you cast this spell, copy it for each other instant and sorcery
//	 spell you've cast this turn. You may choose new targets for the
//	 copies.
//	 Put a +1/+1 counter on target creature. It gains vigilance until end
//	 of turn."
//
// The cast trigger is storm's shape with its own count: the caster's
// instant and sorcery spells cast this turn, this one left out. A copy is
// not cast, so it triggers nothing and counts for nothing (CR 707.10).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f7c29ff-955c-496b-9300-2efb9e99ba2d",
		Name:         "Show of Confidence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "vigilance")),
		Triggered:    []game.TriggeredAbility{effects.S82OldB40CopyPerOtherInstantOrSorcery()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			id := item.Targets[0].ID
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"},
				Label: "Show of Confidence — vigilance until end of turn"}.Apply(ctx)
		},
	})
}
