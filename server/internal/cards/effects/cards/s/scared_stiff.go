package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scared Stiff — Creature — Zombie Coward Guest {2}{B}, 3/2:
//
//	"This creature has menace as long as you control a stickered permanent."
//
// A permanent is stickered while any sticker is on it (CR 123.4, ADR
// 0208); the static reads the battlefield's sticker data, so the keyword
// comes the moment a sticker goes on and goes when the last one comes
// off.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a5d8acd-a460-43b2-816a-70f61d7831af",
		Name:         "Scared Stiff",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.Stickered()), "menace"),
		},
	})
}
