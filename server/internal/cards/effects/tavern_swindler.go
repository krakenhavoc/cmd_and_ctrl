package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tavern Swindler — Creature — Human Rogue {1}{B}, 2/2:
//
//	"{T}, Pay 3 life: Flip a coin. If you win the flip, you gain 6 life."
//
// The life is paid as part of the cost (CR 119.4), so it is gone even if
// the flip is lost. The controller calls the flip (CR 705.2) and a won
// call gains 6 life through the keyed continuation, so a table waiting on
// the call is a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81674138-08d0-4126-9e6d-27f6e7dd2403",
		Name:         "Tavern Swindler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Pay 3 life: Flip a coin. If you win the flip, you gain 6 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), PayLife(3)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				g.FlipCoinForEffect(game.CoinFlipSpec{
					Flipper:  ctx.Controller(),
					Source:   ctx.Source(),
					Question: "Tavern Swindler — call the coin flip",
					ThenKey:  tavernSwindlerFlipped,
				})
				return nil
			},
		}},
	})
}

// tavernSwindlerFlipped is the settled call: a won flip gains the
// flipper 6 life. An on-disk key: never renamed, never reused.
var tavernSwindlerFlipped = game.RegisterCoinFlipThen("coin-flip/tavern-swindler-gain", func(g *game.Game, r game.CoinFlipped) error {
	if !wonFirstFlip(r) {
		return nil
	}
	return g.ChangePlayerLifeForEffect(r.Source, r.Flipper, 6)
})
