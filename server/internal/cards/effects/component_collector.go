package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Component Collector — {2}{U} Creature — Homunculus 1/4 (#2586, ADR
// 0132):
//
//	"If it's neither day nor night, it becomes day as this creature
//	 enters.
//	 Whenever day becomes night or night becomes day, you may tap or
//	 untap target nonland permanent."
//
// "You may tap or untap" is Janjeet Sentry's reading: a pick between the
// two made as the trigger resolves, and tapping a tapped permanent or
// untapping an untapped one does nothing, so the "may" needs no third
// option. The target is chosen as the trigger goes on the stack and
// judged again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6950ab0f-f448-4ed1-9dd4-868cc79a5982",
		Name:         "Component Collector",
		Completeness: CompletenessFull,
		AsEnters:     BecomesDayAsEnters(),
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverDayBecomesNightOrNightBecomesDay("Component Collector — you may tap or untap target nonland permanent",
					b38TapOrUntapTarget("Component Collector")),
				TargetPermanent("target nonland permanent", Nonland())),
		},
	})
}
