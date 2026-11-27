package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Survivor of the Unseen — Creature — Human Wizard {2}{U}, 2/1 (Coldsnap, S82 OLD b559):
//
//	"Cumulative upkeep {2}
//	 {T}: Draw two cards, then put a card from your hand on top of your
//	 library."
//
// Enter the Infinite's put-back with the draw it follows: the card put back
// may be one just drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea96c8c6-b77b-4fc0-8b11-6b31e1ed9554",
		Name:         "Survivor of the Unseen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Survivor of the Unseen — cumulative upkeep {2}", "{2}"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Draw two cards, then put a card from your hand on top of your library.",
			Purpose: game.Purpose{Draws: 2, Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.OldB559DrawTwoThenPutOneBack,
		}},
	})
}
