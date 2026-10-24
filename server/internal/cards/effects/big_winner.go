package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Big Winner — Creature — Ogre Warrior Guest {3}{R}, 5/2:
//
//	"This creature has trample as long as you control a stickered permanent."
//
// A permanent is stickered while any sticker is on it (CR 123.4, ADR
// 0208); the static reads the battlefield's sticker data, so the keyword
// comes the moment a sticker goes on and goes when the last one comes
// off.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b181f5f7-d857-4ed0-996c-bb65b37bfb1b",
		Name:         "Big Winner",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(selfWhileYouControlA(Stickered()), "trample"),
		},
	})
}
