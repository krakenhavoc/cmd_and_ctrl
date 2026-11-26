package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chancellor of the Tangle — Creature — Phyrexian Beast {4}{G}{G}{G}, 6/7:
//
//	"You may reveal this card from your opening hand. If you do, at the
//	 beginning of your first main phase of the game, add {G}.
//	 Reach, vigilance"
//
// The reveal is the Chancellor seam (RevealFromOpeningHand, CR 103.6): the
// armed trigger waits for the revealer's own first precombat main phase.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2a521811-7106-4531-b69c-631e207b8988",
		Name:            "Chancellor of the Tangle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		OpeningHand: effects.RevealFromOpeningHand(game.RevealAtYourFirstMainPhase,
			"Chancellor of the Tangle — add {G}", chancellorOfTheTangleFirstMain),
	})
}

var chancellorOfTheTangleFirstMain = game.SimpleDelayedBody("chancellor-of-the-tangle/first-main-mana", func(g *game.Game, item *game.StackItem) error {
	return effects.AddMana{Produced: "{G}"}.Apply(effects.NewContext(g, item))
})
