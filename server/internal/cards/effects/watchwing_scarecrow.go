package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Watchwing Scarecrow — Artifact Creature — Scarecrow {4}, 2/4
// (Shadowmoor, #5591):
//
//	"This creature has vigilance as long as you control a white
//	 creature.
//	 This creature has flying as long as you control a blue creature."
//
// Rattleblaze Scarecrow's two self grants (layer 6, CR 613.1f), each read
// off the board as it is: the colourless Scarecrow never satisfies its own
// condition.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24bd190c-04bf-44d3-97ba-9dafd9adf7ae",
		Name:         "Watchwing Scarecrow",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(selfWhileYouControlA(And(Creature(), OfColor("W"))), "vigilance"),
			KeywordGrant(selfWhileYouControlA(And(Creature(), OfColor("U"))), "flying"),
		},
	})
}
