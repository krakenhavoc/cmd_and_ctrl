package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "b2223f1c-e607-43e0-86dd-5e3225330066",
		Name:             "Worm Harvest",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Retrace("{2}{B/G}{B/G}{B/G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := b17LandCardsInGraveyard(ctx.Game, item.Controller)
			if n <= 0 {
				return nil
			}
			return CreateToken{Template: TokenCard("1/1 black and green Worm"), N: n}.Apply(ctx)
		},
	})
}
