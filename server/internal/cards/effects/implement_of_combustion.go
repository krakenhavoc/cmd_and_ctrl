package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Implement of Combustion — Artifact {1} (Aether Revolt,
// #5536):
//
//	"{R}, Sacrifice this artifact: It deals 1 damage to target player or
//	 planeswalker.
//	 When this artifact is put into a graveyard from the battlefield,
//	 draw a card."
//
// Goblin Fireslinger's ping, paid with the Implement itself: the
// sacrificed Implement is still the source of the damage (CR 113.7a).
// The draw is Chromatic Star's leaves-for-the-graveyard trigger: it
// fires however the Implement got there, sacrificed to its own ability
// or destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f4bd26d5-4448-451f-9909-b6dea1551af7",
		Name:         "Implement of Combustion",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, Sacrifice this artifact: It deals 1 damage to target player or planeswalker.",
			Cost:    Plus(ManaCost("{R}"), SacrificeThis()),
			Targets: targetPlayerOrPlaneswalker(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsOneToFirstTarget,
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Implement of Combustion — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
