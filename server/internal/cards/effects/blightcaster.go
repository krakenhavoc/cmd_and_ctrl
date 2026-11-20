package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b40e2aae-3a1d-4bc2-bf67-e821ec800434",
		Name:         "Blightcaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Optional(Targeting(
				WheneverYouCast(Enchantment(), "Blightcaster — target creature gets -2/-2 until end of turn",
					boostTargetUntilEOT("Blightcaster — -2/-2 until end of turn", -2, -2)),
				TargetCreature("target creature")),
				"Blightcaster — have target creature get -2/-2 until end of turn?"),
				ForTargets(HarmUntilEOT(0, -2, -2))),
		},
	})
}
