package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "82406175-63b3-4972-be8b-db0e70c908cb",
		Name:         "Krosan Tusker",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{2}{G}")},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventCycle, Self,
				"Krosan Tusker — cycled: you may search your library for a basic land card, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Krosan Tusker — a basic land card, to your hand",
					}.Apply(NewContext(g, item))
				})),
		},
	})
}
