package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kukemssa Pirates — Creature — Human Pirate {3}{U}, 2/2:
//
//	"Whenever this creature attacks and isn't blocked, you may gain control of target artifact
//	 defending player controls. If you do, this creature assigns no combat damage this turn."
//
// #3531: the control change names no duration, so it lasts indefinitely (CR 611.2a).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "280148c5-3a2f-48f7-a147-4578f2068073",
		Name:         "Kukemssa Pirates",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WithTargetsFrom(effects.UnblockedAttackerMay("Kukemssa Pirates", "gain control of target artifact defending player controls", "kukemssa-pirates-gain-control",
				func(ctx *effects.Context, target uuid.UUID) error {
					return effects.GainControl{Target: target, Duration: game.IndefiniteDuration(), Label: "Kukemssa Pirates — gain control"}.Apply(ctx)
				}), effects.TargetOfDefendingPlayer("artifact", "target artifact defending player controls", effects.Artifact())),
		},
	})
}
