package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kaleidostone — Artifact {2} (Conflux, #5587):
//
//	"When this artifact enters, draw a card.
//	 {5}, {T}, Sacrifice this artifact: Add {W}{U}{B}{R}{G}."
//
// Energy Refractor's enters draw and Crystal Quarry's five-colour
// output, behind a sacrifice: still a mana ability (CR 605.1a), since it
// has no target and is not a loyalty ability, so it does not use the
// stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "498ec48f-cb28-48d4-96f4-9d60bb4bb45c",
		Name:         "Kaleidostone",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{samiDrawOnETB("Kaleidostone")},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Sacrifice: true, Mana: "{5}"},
			Produced: "{W}{U}{B}{R}{G}",
			Label:    "{5}, {T}, Sacrifice this artifact: Add {W}{U}{B}{R}{G}",
		}},
	})
}
