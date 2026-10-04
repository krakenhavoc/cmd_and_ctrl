package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wingrattle Scarecrow — Artifact Creature — Scarecrow {3}, 2/2:
//
//	"This creature has flying as long as you control a blue creature.
//	 This creature has persist as long as you control a black creature."
//
// Rattleblaze Scarecrow's shape (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "034264da-a404-4037-b215-b095bba77490",
		Name:         "Wingrattle Scarecrow",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(selfWhileYouControlA(And(Creature(), OfColor("U"))), "flying"),
			KeywordGrant(selfWhileYouControlA(And(Creature(), OfColor("B"))), game.KeywordPersist),
		},
	})
}
