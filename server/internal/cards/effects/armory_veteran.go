package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Armory Veteran — {1}{R} Creature — Orc Warrior 2/2 (Adventures in the
// Forgotten Realms, #5499):
//
//	"As long as this creature is equipped, it has menace."
//
// Armed Assailant's "as long as this creature is equipped" gate on a
// layer 6 keyword grant.
//
// No simplification.
func init() {
	menace := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfOnly(target, g, source) && isEquippedCreature(g, *target)
	}, "menace")
	menace.Label = "As long as this creature is equipped, it has menace."
	Register(Spec{
		OracleID:     "2b51bbb3-1514-4a79-861f-b9cec9f0ec1b",
		Name:         "Armory Veteran",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{menace},
	})
}
