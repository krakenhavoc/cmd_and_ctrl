package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tempest Harvester — Creature — Merfolk Wizard {1}{U}, 2/1:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 {T}, Pay {E}: Draw a card, then discard a card."
//
// ADR 0129 PR 1 (#1995). The discard is chosen from the hand after the
// draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c0ffe37-16ce-4788-a434-58b6fc34bd9c",
		Name:         "Tempest Harvester",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Tempest Harvester", 2),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay {E}: Draw a card, then discard a card.",
			Cost:    effects.Plus(effects.TapCost(), effects.PayEnergy(1)),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.LootOne(g, item, 1)
			},
		}},
	})
}
