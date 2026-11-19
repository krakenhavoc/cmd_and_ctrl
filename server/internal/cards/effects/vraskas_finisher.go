package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vraska's Finisher — Creature — Gorgon Assassin {2}{B}, 3/2 (Foundations
// Jumpstart, S82, #5929):
//
//	"When this creature enters, destroy target creature or planeswalker an
//	 opponent controls that was dealt damage this turn."
//
// The shared destroyFirstLegalTarget body, with Bitter Downfall's
// DealtDamageThisTurn predicate, so only a permanent the engine records as
// dealt damage this turn can be chosen (CR 115.1, CR 608.2b).
//
// No simplification.
func init() {
	const label = "Vraska's Finisher — destroy target creature or planeswalker an opponent controls that was dealt damage this turn"
	Register(Spec{
		OracleID:     "5df16529-6f1d-4978-b364-97d91abac441",
		Name:         "Vraska's Finisher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters(label, destroyFirstLegalTarget),
				TargetPermanent("target creature or planeswalker an opponent controls that was dealt damage this turn",
					Or(Creature(), Planeswalker()), OpponentControls(), DealtDamageThisTurn())),
		},
	})
}
