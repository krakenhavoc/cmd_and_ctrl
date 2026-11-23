package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blightcaster — Creature — Human Wizard {3}{B}, 2/3 (Magic 2014, #5563):
//
//	"Whenever you cast an enchantment spell, you may have target creature get -2/-2 until end of turn."
//
// A cast trigger (CR 603.2) that resolves before the enchantment spell does.
// "You may have" is a CR 603.5 optional trigger: the target is chosen as it
// goes on the stack and the choice is asked as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b40e2aae-3a1d-4bc2-bf67-e821ec800434",
		Name:         "Blightcaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Optional(effects.Targeting(
				effects.WheneverYouCast(effects.Enchantment(), "Blightcaster — target creature gets -2/-2 until end of turn",
					effects.BoostTargetUntilEOT("Blightcaster — -2/-2 until end of turn", -2, -2)),
				effects.TargetCreature("target creature")),
				"Blightcaster — have target creature get -2/-2 until end of turn?"),
				effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
		},
	})
}
