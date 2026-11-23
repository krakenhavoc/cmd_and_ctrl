package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thrór's Map — Legendary Artifact {2} (The Hobbit, HOB):
//
//	"When Thrór's Map enters, search your library for a basic land
//	 card, reveal it, put it into your hand, then shuffle.
//	 {2}, {T}: Draw a card, then discard a card."
//
// The trigger is Mycosynth Wellspring's search without the "you may":
// the search is mandatory, so the prompt offers no decline, and the pick
// is revealed and goes to hand. The loot is the shared draw-then-discard
// (lootOne), drawing first and opening the discard once the draw has
// resolved (#4454).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "925cacdc-6d09-4fbd-9258-9f0539da180c",
		Name:         "Thrór's Map",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Thrór's Map — search for a basic land card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Thrór's Map — a basic land card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.LootOne(g, item, 1)
			},
		}},
	})
}
