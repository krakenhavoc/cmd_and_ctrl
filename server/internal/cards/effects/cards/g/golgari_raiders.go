package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Golgari Raiders — Creature — Elf Warrior {3}{G}, 0/0 (Guilds of Ravnica,
// #5525):
//
//	"Haste
//	 Undergrowth — This creature enters with a +1/+1 counter on it for each
//	 creature card in your graveyard."
//
// A replacement effect on its own entry (CR 614.1c), counted as it enters. A
// 0/0 with an empty graveyard enters with no counters and is put into the
// graveyard by the next state-based check (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "abd12238-ea9d-4124-8229-3a479b8464d7",
		Name:            "Golgari Raiders",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				return effects.B11CreatureCardsInGraveyard(g, src.Controller)
			}, "Golgari Raiders: enters with a +1/+1 counter for each creature card in your graveyard"),
		},
	})
}
