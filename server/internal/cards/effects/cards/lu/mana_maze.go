package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mana Maze — Enchantment {1}{U} (Invasion):
//
//	"Players can't cast spells that share a color with the spell most
//	 recently cast this turn."
//
// The restriction is #6217's CantCastSharingAColorWithTheLastSpell, read
// against the colours of the spell most recently cast this turn as it was
// cast (CR 105.2) for every player. A colourless spell shares no colour, and
// before the turn's first spell nothing is refused.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b6f9d55-eae8-48fd-a1d1-d56939adaf3a",
		Name:         "Mana Maze",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			effects.CantCastSharingAColorWithTheLastSpell("Players can't cast spells that share a color with the spell most recently cast this turn."),
		},
	})
}
