package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sheriff of Safe Passage — Creature — Human Knight {2}{W}, 0/0
// (Outlaws of Thunder Junction, #4468):
//
//	"This creature enters with a +1/+1 counter on it plus an additional
//	 +1/+1 counter on it for each other creature you control.
//	 Plot {1}{W} (You may pay {1}{W} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// Golgari Grave-Troll's CR 614.1c self-entry replacement, counted as the
// Sheriff enters: one counter, plus one for each creature its controller
// already controls (the Sheriff is not on the battlefield yet, and is
// skipped by ID besides). The counters arrive as one batch, so a counter
// doubler sees all of them. Plot is the ordinary special action
// (CR 702.170).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "ec16008f-18d6-4461-b99c-0a504f211618",
		Name:           "Sheriff of Safe Passage",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Plot("{1}{W}")},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				return 1 + effects.CountPermanentsYouControl(g, src, func(c *game.Card) bool {
					return c.InstanceID != src.InstanceID && c.IsCreature()
				})
			}, "Sheriff of Safe Passage: enters with a +1/+1 counter plus one for each other creature you control"),
		},
	})
}
