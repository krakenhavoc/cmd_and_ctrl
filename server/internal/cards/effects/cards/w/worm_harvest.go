package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Worm Harvest — Sorcery {2}{B/G}{B/G}{B/G}:
//
//	"Create a 1/1 black and green Worm creature token for each land
//	 card in your graveyard.
//	 Retrace"
//
// Retrace (CR 702.81) is a graveyard cast for the PRINTED mana cost plus
// a land card discarded from hand: see Spitting Image. The land that pays
// retrace is discarded as a COST, with the spell already on the stack,
// so it is in the graveyard by resolution and counts here — the reason
// the card is a land-fuelled engine rather than a one-shot. The count is
// taken at resolution, not at cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "b2223f1c-e607-43e0-86dd-5e3225330066",
		Name:             "Worm Harvest",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Retrace("{2}{B/G}{B/G}{B/G}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.B17LandCardsInGraveyard(ctx.Game, item.Controller)
			if n <= 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 black and green Worm"), N: n}.Apply(ctx)
		},
	})
}
