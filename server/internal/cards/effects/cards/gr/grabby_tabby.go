package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grabby Tabby — Creature — Cat Guest {3}{G}, 4/3:
//
//	"This creature has vigilance as long as you control a stickered permanent."
//
// A permanent is stickered while any sticker is on it (CR 123.4, ADR
// 0208); the static reads the battlefield's sticker data, so the keyword
// comes the moment a sticker goes on and goes when the last one comes
// off.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67ee367c-e346-48b6-8694-01c8d5f020cb",
		Name:         "Grabby Tabby",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.Stickered()), "vigilance"),
		},
	})
}
