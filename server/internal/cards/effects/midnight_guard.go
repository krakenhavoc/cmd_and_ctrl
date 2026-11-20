package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Guard — Creature — Human Soldier {2}{W}, 2/3 (Dark Ascension, #5571):
//
//	"Whenever another creature enters, untap this creature."
//
// Soul Warden's trigger condition (any other creature, under any player's
// control) with Fain's untap of the source (b31UntapSelf), which does
// nothing once the Guard has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5430b0e-e77f-47b8-ae6d-30bcff4d7e35",
		Name:         "Midnight Guard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureEntered, "Midnight Guard — untap this creature", b31UntapSelf),
		},
	})
}
