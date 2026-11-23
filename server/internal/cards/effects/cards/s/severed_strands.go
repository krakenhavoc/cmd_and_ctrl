package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Severed Strands — Sorcery {1}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 You gain life equal to the sacrificed creature's toughness. Destroy
//	 target creature an opponent controls."
//
// The life is the sacrificed creature's toughness as it last existed on
// the battlefield (the 2018-10-05 ruling, CR 608.2h), read off the
// payment record (ADR 0113 §1). The rulings: an illegal target means the
// spell does not resolve and no life is gained (CR 608.2b); a legal
// target that can't be destroyed still gains the life.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "25f0527c-1340-4218-970f-9e4f84ef96e8",
		Name:           "Severed Strands",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		// The destruction waits for a prompt the life gain asks (a CR 616 order; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Amount: effects.SacrificedToughness(ctx)}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				legal := ctx.LegalTargets()
				if len(legal) == 0 || legal[0].Kind != game.TargetCard {
					return nil
				}
				return effects.DestroyTarget{Target: legal[0].ID}.Apply(ctx)
			}),
		),
	})
}
