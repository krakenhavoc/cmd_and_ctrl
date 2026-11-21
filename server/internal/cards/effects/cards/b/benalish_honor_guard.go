package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Benalish Honor Guard — Creature — Human Knight {1}{W}, 2/2 (Jumpstart 2022, #5948):
//
//	"This creature gets +1/+0 for each legendary creature you control."
//
// Academy Instructor's per-count layer 7c static, counting the legendary
// creatures its controller controls on every layer pass. The count is read
// from the board as the layers are computed, so it follows a legend
// entering, dying or changing hands.
//
// No simplification.
func init() {
	pump := effects.PumpSelfCreatureOrAttachedPer(1, 0, func(g *game.Game, source *game.Card) int {
		return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.IsCreature() && c.IsLegendary() })
	})
	pump.Label = "This creature gets +1/+0 for each legendary creature you control."
	effects.Register(effects.Spec{
		OracleID:     "0595b074-8f25-469d-9e82-ed60c4a7a2cd",
		Name:         "Benalish Honor Guard",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{pump},
	})
}
