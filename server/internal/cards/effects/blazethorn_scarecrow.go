package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blazethorn Scarecrow — Artifact Creature — Scarecrow {5}, 3/3
// (Shadowmoor, #5591):
//
//	"This creature has haste as long as you control a red creature.
//	 This creature has wither as long as you control a green creature.
//	 (It deals damage to creatures in the form of -1/-1 counters.)"
//
// Rattleblaze Scarecrow's two self grants (layer 6, CR 613.1f), each read
// off the board as it is: the colourless Scarecrow never satisfies its own
// condition.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f7a4ffb-83f7-46e5-b0a7-c8b068af6969",
		Name:         "Blazethorn Scarecrow",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(selfWhileYouControlA(And(Creature(), OfColor("R"))), "haste"),
			KeywordGrant(selfWhileYouControlA(And(Creature(), OfColor("G"))), "wither"),
		},
	})
}
