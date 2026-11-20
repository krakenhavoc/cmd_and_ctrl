package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ballynock Trapper — Creature — Kithkin Soldier {3}{W}, 2/2 (Eventide, #5590):
//
//	"{T}: Tap target creature.
//	 Whenever you cast a white spell, you may untap this creature."
//
// Sinew Dancer's tap and the optional cast-trigger untap (eveB1UntapOnCast).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "96232342-3ec7-4740-a114-969537b8f2fc",
		Name:         "Ballynock Trapper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Tap target creature.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  thousandMoonsCrackshotTap,
		}},
		Triggered: []game.TriggeredAbility{eveB1UntapOnCast("Ballynock Trapper", "W")},
	})
}
