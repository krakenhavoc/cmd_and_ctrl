package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tandem Tactics — Instant {1}{W} (Ravnica: Clue Edition):
//
//	"Up to two target creatures each get +1/+2 until end of turn. You
//	 gain 2 life."
//
// Each still-legal target gets the boost (CR 608.2b); the life is gained
// even with no target chosen, but not when every chosen target became
// illegal (the spell is countered on resolution).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e42cd4da-f049-4304-a16f-911769985e6c",
		Name:         "Tandem Tactics",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 2)), LifeGain: 2},
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 2, Label: "Tandem Tactics — +1/+2"}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.GainLife{Amount: 2}.Apply(ctx)
		},
	})
}
