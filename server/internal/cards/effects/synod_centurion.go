package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Synod Centurion — Artifact Creature {4}, 4/4:
//
//	"When you control no other artifacts, sacrifice this creature."
//
// ADR 0107 §1 (#1858). A CR 603.8 state trigger; the Centurion itself
// does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe6337dc-df02-4baa-9fe0-b14e8d34186a",
		Name:         "Synod Centurion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouControlNoOther(QueryType("artifact"), "Synod Centurion — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
