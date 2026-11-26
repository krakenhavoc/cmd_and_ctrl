package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Divinity of Pride — Creature — Spirit Avatar {W/B}{W/B}{W/B}{W/B}{W/B}, 4/4
// (Commander 2013, S82 OLD b386):
//
//	"Flying, lifelink
//	 This creature gets +4/+4 as long as you have 25 or more life."
//
// Angel of Vitality's life-gated self pump at +4/+4, read live off the
// controller's life total.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c91c236-34d7-4454-a55a-784db7f68bde",
		Name:            "Divinity of Pride",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Static: []game.StaticAbility{
			effects.LifeGatedPump(effects.SelfWhileYourLifeAtLeast(25), 4, 4),
		},
	})
}
