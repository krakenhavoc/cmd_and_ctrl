package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx of Foresight — Creature — Sphinx {2}{U}{U}, 4/4 (Ravnica Allegiance):
//
//	"You may reveal this card from your opening hand. If you do, scry 3 at the
//	 beginning of your first upkeep.
//	 Flying
//	 At the beginning of your upkeep, scry 1."
//
// The opening-hand reveal is Devourer of Destiny's seam
// (RevealFromOpeningHandForYourFirstUpkeep, CR 103.6): the table is asked at
// the start of the game and a yes arms a delayed trigger for the first upkeep.
// The ordinary upkeep scry is Mystic Redaction's trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d1108bd1-87b2-4795-9690-54ff4504e04f",
		Name:            "Sphinx of Foresight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 1},
		OpeningHand: effects.RevealFromOpeningHandForYourFirstUpkeep(
			"Sphinx of Foresight — scry 3", sphinxOfForesightFirstUpkeep),
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Sphinx of Foresight — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}

var sphinxOfForesightFirstUpkeep = game.SimpleDelayedBody("sphinx-of-foresight/first-upkeep-scry", func(g *game.Game, item *game.StackItem) error {
	return effects.Scry{Player: item.Controller, N: 3}.Apply(effects.NewContext(g, item))
})
