package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Griffnaut Tracker — Creature — Human Detective {3}{W}, 2/3:
//
//	"Flying
//	 When this creature enters, exile up to two target cards from a
//	 single graveyard."
//
// #1807, ADR 0106 §5. The entry trigger targets as it goes on the
// stack, so the picks are made then and judged again as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6eff5e17-946b-4433-9f48-88f103844c42",
		Name:            "Griffnaut Tracker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersExileFromASingleGraveyard("Griffnaut Tracker", 2, ExileTargetCards),
		},
	})
}
