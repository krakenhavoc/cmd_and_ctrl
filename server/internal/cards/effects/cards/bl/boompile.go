package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boompile — Artifact {4} (Commander 2016, #5537):
//
//	"{T}: Flip a coin. If you win the flip, destroy all nonland
//	 permanents."
//
// Sorcerer's Strongbox's keyed coin flip: the activator calls it as the
// ability resolves, and only a won flip destroys every nonland permanent
// (Plague Boiler's wipe, Boompile itself included). A lost flip does
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85b8902f-2515-4ded-b28a-a83334f089f1",
		Name:         "Boompile",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Flip a coin. If you win the flip, destroy all nonland permanents.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy}},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				g.FlipCoinForEffect(game.CoinFlipSpec{
					Flipper:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Boompile — call the coin flip",
					ThenKey:  boompileFlipped,
					Asked:    game.NewAskedItem(item),
				})
				return nil
			},
		}},
	})
}

// boompileFlipped is "if you win the flip, destroy all nonland
// permanents". An on-disk key.
var boompileFlipped = effects.CoinFlipThen("coin-flip/boompile", func(ctx *effects.Context, r game.CoinFlipped) error {
	if !effects.WonFirstFlip(r) {
		return nil
	}
	return effects.DestroyAllMatching{Match: effects.Nonland()}.Apply(ctx)
})
