package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blister Beetle — Creature — Insect {1}{B}, 1/1 (Shards of Alara, S82 OLD
// batch 517, #5588):
//
//	"When this creature enters, target creature gets -1/-1 until end of turn."
//
// Any creature, the Beetle itself included. The target is chosen as the
// trigger goes on the stack and the shrink applies only if it is still legal
// as the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ec6b0f1-8928-457c-adf0-c74a28edc86f",
		Name:         "Blister Beetle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Blister Beetle — target creature gets -1/-1 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						target := effects.FirstLegalBattlefieldTarget(ctx)
						if target == uuid.Nil {
							return nil
						}
						return effects.BoostUntilEOT{Target: target, Power: -1, Toughness: -1, Label: "Blister Beetle — -1/-1 until end of turn"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature")),
		},
	})
}
