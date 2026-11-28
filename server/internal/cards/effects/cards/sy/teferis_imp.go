package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teferi's Imp — Creature — Imp {2}{U}, 1/1 (Mirage, S82 OLD batch 812, #5633):
//
//	"Flying
//	 Phasing (This phases in or out before you untap during each of your untap
//	 steps. While it's phased out, it's treated as though it doesn't exist.)
//	 Whenever this creature phases out, discard a card.
//	 Whenever this creature phases in, draw a card."
//
// The #6201 pair of triggers read off the permanent itself (CR 603.10b): the
// discard is the controller's own choice from their hand, the draw is a plain
// draw. Both belong to the creature's controller as the trigger goes on the
// stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e4826967-b2f9-4f32-8411-d243edf41082",
		Name:            "Teferi's Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordPhasing},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisPhasesOut("Teferi's Imp — discard a card", func(g *game.Game, item *game.StackItem) error {
				g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: item.Controller, Source: item.SourceCardID, N: 1})
				return nil
			}),
			effects.WheneverThisPhasesIn("Teferi's Imp — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
