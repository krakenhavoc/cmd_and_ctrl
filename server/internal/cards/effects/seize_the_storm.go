package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seize the Storm — Sorcery {4}{R} (Innistrad: Midnight Hunt promo):
//
//	"Create a red Elemental creature token with trample and 'This token's
//	 power and toughness are each equal to the number of instant and sorcery
//	 cards in your graveyard plus the number of cards with flashback you own
//	 in exile.'
//	 Flashback {6}{R}"
//
// The Elemental sizes itself with a layer 7a static of its own template, so
// it keeps counting for its controller whatever happens to the spell. A
// flashed-back Seize the Storm is exiled as it leaves the stack and so is a
// card with flashback in exile, which the token counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d3455e93-3a05-42db-945d-c4c85f4ac827",
		Name:             "Seize the Storm",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{6}{R}")},
		Purpose:          game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: tokenFromCatalog(printedSeizeTheStormElementalToken), N: 1}.Apply(ctx)
		},
	})
}
