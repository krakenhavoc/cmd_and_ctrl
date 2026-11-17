package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "33ef3d51-845d-460b-97a0-95e7591cf466",
		Name:         "Gnawing Vermin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Gnawing Vermin — target player mills two cards",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						who, ok := firstLegalPlayerTarget(ctx)
						if !ok {
							return nil
						}
						return MillCards{Player: who, N: 2}.Apply(ctx)
					}),
				TargetPlayer("target player"),
			),
			Targeting(
				WhenThisDies("Gnawing Vermin — target creature you don't control gets -1/-1 until end of turn",
					boostTargetUntilEOT("Gnawing Vermin", -1, -1)),
				TargetCreature("target creature you don't control", OpponentControls()),
			),
		},
	})
}
