package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Gaffer — Legendary Creature — Halfling Peasant {2}{W}, 2/3
// (EDHREC rank 1672):
//
//	"At the beginning of each end step, if you gained 3 or more life
//	 this turn, draw a card."
//
// The lifegain deck's card-a-turn — EACH end step, so on every
// player's turn the question is asked. An intervening-if (CR 603.4):
// the trigger goes on the stack only if the controller has gained
// three life this turn, and the resolution checks again. Nothing
// can un-gain life, so the second check is a formality; it is there
// because the printed clause says so.
//
// The count is the per-turn tally's LifeGained cell — every positive
// life change for the controller this turn (b15LifeGainedThisTurn). A
// lifelink hit, a Soul Warden trigger and a drain's gain half all
// count; a life LOSS does not offset a gain, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a177295e-3b58-4e46-a1cb-fc003a7a0848",
		Name:         "The Gaffer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B15EndStepBegan(ev) && effects.B15LifeGainedThisTurn(g, source.Controller) >= 3
			}, "The Gaffer — draw a card", effects.Do(effects.If{
				Cond: effects.CondLifeGained{AtLeast: 3},
				Then: []effects.Applier{effects.DrawCards{N: 1}},
			})),
		},
	})
}
