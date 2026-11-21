package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Expedition Skulker — Creature — Vampire Rogue {1}{B}, 2/2 (Foundations
// Jumpstart):
//
//	"This creature has deathtouch as long as you control another Rogue."
//
// A layer 6 self grant gated on some other Rogue its controller controls,
// read through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b6e0d65-202e-4fd3-861e-8d2eaffe3269",
		Name:         "Expedition Skulker",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.BnsB29SelfWhileAnother("Rogue"), "deathtouch"),
		},
	})
}
