package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Janjeet Sentry — Creature — Vedalken Soldier {2}{U}, 2/3:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 {T}, Pay {E}{E}: You may tap or untap target artifact or creature."
//
// ADR 0129 PR 1 (#1995). "You may tap or untap" is a pick between the
// two made as the ability resolves (Gandalf the Grey's and Merrow
// Reejerey's reading); tapping a tapped permanent or untapping an
// untapped one does nothing, so the "may" needs no third option.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb818d43-8e5d-41aa-98f8-5826dfd5dfdb",
		Name:         "Janjeet Sentry",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Janjeet Sentry", 2),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay {E}{E}: You may tap or untap target artifact or creature.",
			Cost:    effects.Plus(effects.TapCost(), effects.PayEnergy(2)),
			Targets: effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			Effect:  effects.B38TapOrUntapTarget("Janjeet Sentry"),
		}},
	})
}
