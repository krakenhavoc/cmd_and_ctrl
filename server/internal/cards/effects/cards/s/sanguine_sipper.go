package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanguine Sipper — Creature — Vampire Guest {1}{W}, 3/1:
//
//	"This creature has lifelink as long as you control a stickered permanent."
//
// A permanent is stickered while any sticker is on it (CR 123.4, ADR
// 0208); the static reads the battlefield's sticker data, so the keyword
// comes the moment a sticker goes on and goes when the last one comes
// off.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea787e29-5e24-4136-bdf6-bee338899e34",
		Name:         "Sanguine Sipper",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.Stickered()), "lifelink"),
		},
	})
}
