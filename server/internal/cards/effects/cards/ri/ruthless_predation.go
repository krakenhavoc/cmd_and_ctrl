package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruthless Predation — Sorcery {1}{G} (Phyrexia: All Will Be One):
//
//	"Target creature you control gets +1/+2 until end of turn. It fights
//	 target creature you don't control."
//
// Chelonian Tackle's shape with a required second target: slot 0 is your
// creature, slot 1 the opposing one. The pump is its own step, so the fight
// reads the boosted power (CR 608.2c). Both targets are re-checked as the
// spell resolves; if either is no longer legal the fight does not happen
// (CR 701.14b, CR 608.2b) but a still-legal first target keeps its pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "478e59b8-03ad-4f8e-8c18-6ab5ad77a6f8",
		Name:         "Ruthless Predation",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 2)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{
					Target:    t.ID,
					Power:     1,
					Toughness: 2,
					Label:     "Ruthless Predation — +1/+2 until end of turn",
				}.Apply(ctx)
			}),
			effects.Instruction(effects.FightTargetClauses),
		),
	})
}
