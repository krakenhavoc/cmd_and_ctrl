package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hungering Yeti — {4}{R} Creature — Yeti 4/4 (Fate Reforged):
//
//	"As long as you control a green or blue permanent, you may cast this spell
//	 as though it had flash."
//
// Bard's Company's shape: the permission is judged for the caster as the cast
// begins (CR 601.3), on effective colors, and it is not an alternative cost.
// The Yeti is in the hand then, so it never counts itself.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c179fae-f767-4246-88d3-aa38da5bf6f4",
		Name:         "Hungering Yeti",
		Completeness: effects.CompletenessFull,
		FlashIf:      effects.FlashIfYou("as long as you control a green or blue permanent", hungeringYetiControlsGreenOrBlue),
	})
}

// hungeringYetiControlsGreenOrBlue is "you control a green or blue permanent".
func hungeringYetiControlsGreenOrBlue(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && (c.HasColor("G") || c.HasColor("U")) {
			return true
		}
	}
	return false
}
