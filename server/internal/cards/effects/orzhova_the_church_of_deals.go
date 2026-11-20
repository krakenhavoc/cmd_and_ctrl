package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orzhova, the Church of Deals — Land (Guildpact, #5599):
//
//	"{T}: Add {C}.
//	 {3}{W}{B}, {T}: Target player loses 1 life and you gain 1 life."
//
// High Market's shape: a colourless mana ability and an activated drain
// (Gnawing Zombie's body, drainTargetOpponentOne). The target may be any
// player; one who has left the game in response counters the ability
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8551a9cf-c54b-42d4-92d6-550f4890a3d7",
		Name:         "Orzhova, the Church of Deals",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}{B}, {T}: Target player loses 1 life and you gain 1 life.",
			Purpose: game.Purpose{LifeGain: 1},
			Cost:    Plus(ManaCost("{3}{W}{B}"), TapCost()),
			Targets: TargetPlayer("target player"),
			Effect:  drainTargetOpponentOne,
		}},
	})
}
