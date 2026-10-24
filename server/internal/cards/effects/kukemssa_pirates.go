package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "280148c5-3a2f-48f7-a147-4578f2068073",
		Name:         "Kukemssa Pirates",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			withTargetsFrom(unblockedAttackerMay("Kukemssa Pirates", "gain control of target artifact defending player controls", "kukemssa-pirates-gain-control",
				func(ctx *Context, target uuid.UUID) error {
					return GainControl{Target: target, Duration: game.IndefiniteDuration(), Label: "Kukemssa Pirates — gain control"}.Apply(ctx)
				}), TargetOfDefendingPlayer("artifact", "target artifact defending player controls", Artifact())),
		},
	})
}
