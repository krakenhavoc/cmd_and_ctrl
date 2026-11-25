package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spikewheel Acrobat — Creature — Human Rogue {3}{R}, 5/2 (Ravnica
// Allegiance):
//
//	"Spectacle {2}{R} (You may cast this spell for its spectacle cost rather
//	 than its mana cost if an opponent lost life this turn.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e9b403db-8981-440f-a451-aa77ae3cf006",
		Name:             "Spikewheel Acrobat",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Spectacle("{2}{R}")},
	})
}
