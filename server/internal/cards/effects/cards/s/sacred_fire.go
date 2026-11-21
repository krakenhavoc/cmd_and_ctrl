package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sacred Fire — Instant {R}{W} (Midnight Hunt, #5497):
//
//	"Sacred Fire deals 2 damage to any target and you gain 2 life.
//	 Flashback {4}{R}{W} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// The damage and then the life, as one resolution (the life is gained even
// when the damage is prevented); Flashback is the shared constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "812bf52c-37f5-41c9-8a69-009fddbafb3f",
		Name:             "Sacred Fire",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetAny(),
		Purpose:          effects.ForTargets(effects.DamageToTarget(0, 2)),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{4}{R}{W}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DamageToFirstTargetStep(2)(item, ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: 2}.Apply(ctx)
		},
	})
}
