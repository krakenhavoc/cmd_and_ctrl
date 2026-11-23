package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnawing Vermin — Creature — Rat {B}, 1/1 (The Brothers' War, #5482):
//
//	"When this creature enters, target player mills two cards.
//	 When this creature dies, target creature you don't control gets -1/-1
//	 until end of turn."
//
// Two printed triggers, two registered. Any player may be the first's
// target, the controller included; the second's creature must be one its
// controller does not control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33ef3d51-845d-460b-97a0-95e7591cf466",
		Name:         "Gnawing Vermin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Gnawing Vermin — target player mills two cards",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						who, ok := effects.FirstLegalPlayerTarget(ctx)
						if !ok {
							return nil
						}
						return effects.MillCards{Player: who, N: 2}.Apply(ctx)
					}),
				effects.TargetPlayer("target player"),
			),
			effects.Targeting(
				effects.WhenThisDies("Gnawing Vermin — target creature you don't control gets -1/-1 until end of turn",
					effects.BoostTargetUntilEOT("Gnawing Vermin", -1, -1)),
				effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
			),
		},
	})
}
