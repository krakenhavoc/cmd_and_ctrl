package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Croakid Amphibonaut — Creature — Frog Guest {4}{U}, 4/3:
//
//	"This creature has flying as long as you control a stickered permanent."
//
// A permanent is stickered while any sticker is on it (CR 123.4, ADR
// 0208); the static reads the battlefield's sticker data, so the keyword
// comes the moment a sticker goes on and goes when the last one comes
// off.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b1796d2-5b19-48ab-8699-01f45d427498",
		Name:         "Croakid Amphibonaut",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(selfWhileYouControlA(Stickered()), "flying"),
		},
	})
}
