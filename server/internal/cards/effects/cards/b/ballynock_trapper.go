package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ballynock Trapper — Creature — Kithkin Soldier {3}{W}, 2/2 (Eventide, #5590):
//
//	"{T}: Tap target creature.
//	 Whenever you cast a white spell, you may untap this creature."
//
// Sinew Dancer's tap and the optional cast-trigger untap (eveB1UntapOnCast).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96232342-3ec7-4740-a114-969537b8f2fc",
		Name:         "Ballynock Trapper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Tap target creature.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.ThousandMoonsCrackshotTap,
		}},
		Triggered: []game.TriggeredAbility{effects.EveB1UntapOnCast("Ballynock Trapper", "W")},
	})
}
