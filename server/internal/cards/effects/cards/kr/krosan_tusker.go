package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krosan Tusker — Creature — Boar Beast {5}{G}{G}, 6/5:
//
//	"Cycling {2}{G}
//	 When you cycle this card, you may search your library for a basic
//	 land card, reveal that card, put it into your hand, then shuffle. (Do
//	 this before you draw.)"
//
// The cycling trigger fires from the graveyard (CR 702.29c) like every
// "when you cycle this card" row, and goes on the stack above the cycling
// ability, so it resolves, and the search finishes, before the draw. The
// search is an optional basic-land fetch to hand with a reveal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82406175-63b3-4972-be8b-db0e70c908cb",
		Name:         "Krosan Tusker",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventCycle, effects.Self,
				"Krosan Tusker — cycled: you may search your library for a basic land card, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Krosan Tusker — a basic land card, to your hand",
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
