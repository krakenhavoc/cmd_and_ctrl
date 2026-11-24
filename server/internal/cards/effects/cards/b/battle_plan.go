package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Battle Plan — Enchantment {3}{R} (Modern Horizons 2, #5501):
//
//	"At the beginning of combat on your turn, target creature you control
//	 gets +2/+0 until end of turn.
//	 Basic landcycling {1}{R}"
//
// The trigger targets as it goes on the stack and does nothing if the
// creature has gone by resolution (CR 608.2b). Basic landcycling is the
// typecycling activation from hand (CR 702.29e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09b4fba6-d943-49e4-9f9e-a74fcbabe83f",
		Name:         "Battle Plan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtBeginningOfYourCombat("Battle Plan — target creature you control gets +2/+0 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id := effects.FirstLegalBattlefieldTarget(ctx)
						if id == uuid.Nil {
							return nil
						}
						return effects.BoostUntilEOT{Target: id, Power: 2, Label: "Battle Plan — +2/+0 until end of turn"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{1}{R}")},
	})
}
