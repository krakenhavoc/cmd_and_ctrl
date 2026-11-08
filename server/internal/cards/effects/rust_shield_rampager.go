package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rust-Shield Rampager — Creature — Raccoon Warrior {3}{G}, 4/4
// (Bloomburrow, #4467):
//
//	"Offspring {2} (You may pay an additional {2} as you cast this
//	 spell. If you do, when this creature enters, create a 1/1 token
//	 copy of it.)
//	 This creature can't be blocked by creatures with power 2 or less."
//
// Offspring is the shared seam (offspring.go): the optional {2} and the
// entry trigger that makes a 1/1 token copy, which keeps the block rule.
// The block rule is Bristlebane Outrider's (CantBeBlockedBy over
// PowerLE(2)), each blocker's power read as blockers are declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "23878990-dcbd-4a2a-b4f7-ee7edb396c71",
		Name:          "Rust-Shield Rampager",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offspring("{2}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Rust-Shield Rampager"),
		},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
	})
}
