package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rooftop Assassin — Creature — Vampire Assassin {3}{B}, 2/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"Flash
//	 Flying, lifelink
//	 When this creature enters, destroy target creature an opponent
//	 controls that was dealt damage this turn."
//
// Stingblade Assassin with lifelink: the entry trigger is the shared
// DealtDamageThisTurn predicate on a creature an opponent controls,
// destroyed on resolution. With no such creature the trigger has no
// target and is removed from the stack (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f3baef8c-f8d7-4cb2-b5cc-b889a991e4e8",
		Name:            "Rooftop Assassin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Rooftop Assassin — destroy target creature an opponent controls that was dealt damage this turn", destroyChosenPermanent),
				TargetCreature("target creature an opponent controls that was dealt damage this turn", OpponentControls(), DealtDamageThisTurn())),
		},
	})
}
