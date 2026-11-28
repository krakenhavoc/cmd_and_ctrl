package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crumbling Sanctuary — Artifact {5} (Mercadian Masques, S82 OLD batch 700, #5619):
//
//	"If damage would be dealt to a player, that player exiles that many cards
//	 from the top of their library instead."
//
// Force Bubble's replacement widened to every player (CR 614.1a, "instead", so
// not a prevention): the damage never happens, and that many cards leave the
// top of the damaged player's library for exile. It replaces combat and
// noncombat damage alike and does nothing to damage dealt to a permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9269fda-e1e6-4e10-9a64-26e8c0f38b0c",
		Name:         "Crumbling Sanctuary",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.S82OldB700DamageToAPlayerExilesThatMany()},
	})
}
