package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elektra, Daughter of the Hand — Legendary Creature — Human Ninja
// Villain {2}{B}{B}, 3/3 (Marvel Super Heroes, #4455):
//
//	"Sneak {1}{B}{B} (You may cast this spell for {1}{B}{B} if you also
//	 return an unblocked attacker you control to hand during the declare
//	 blockers step. She enters tapped and attacking.)
//	 When Elektra enters, destroy target creature an opponent controls
//	 with power 3 or less."
//
// The first sneak card (#4586, CR 702.190). The keyword is Sneak(cost)
// (sneak.go); the enter trigger is Red Guardian's targeted destroy with a
// power ceiling (PowerLE) for its other clause. The trigger fires however
// she enters, cast for her mana cost or sneaked in.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "b46647a1-0c91-451b-8c0b-be8fe02faee2",
		Name:             "Elektra, Daughter of the Hand",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{1}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Elektra — destroy target creature an opponent controls with power 3 or less", destroyChosenPermanent),
				TargetCreature("target creature an opponent controls with power 3 or less", OpponentControls(), PowerLE(3))),
		},
	})
}
