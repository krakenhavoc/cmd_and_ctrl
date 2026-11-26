package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voracious Wurm — Creature — Wurm {1}{G}, 2/2 (Magic 2014, S82 old sets):
//
//	"This creature enters with X +1/+1 counters on it, where X is the amount of
//	 life you've gained this turn."
//
// Ignition Team's entry replacement (CR 614.1c): the count is read as the Wurm
// enters, against the controller it enters under, from the turn's life-gained
// tally (the sum of every gain, so a gain of 3 and a gain of 2 make 5, and
// life lost does not reduce it).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a249c4b-beaa-4e7f-b4f7-496d282327bd",
		Name:         "Voracious Wurm",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				return g.TurnTallyFor(src.Controller).LifeGained
			}, "Voracious Wurm enters with X +1/+1 counters, X the life you've gained this turn"),
		},
	})
}
