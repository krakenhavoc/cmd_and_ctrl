package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fractured Sanity — Sorcery {U}{U}{U}:
//
//	"Each opponent mills fourteen cards.
//	 Cycling {1}{U} ({1}{U}, Discard this card: Draw a card.)
//	 When you cycle this card, each opponent mills four cards."
//
// The spell is Altar of the Brood's each-opponent mill
// (b12EachOpponentMills) at fourteen; a library with fewer mills what it
// has (CR 701.17b). Cycling is the engine's keyword ability (Cycling),
// and the cycle trigger fires from the graveyard, where the card already
// is when it triggers (CR 702.29c) — Agonasaur Rex's shape:
// InGraveyard(On(game.EventCycle, Self, …)). The trigger uses the stack,
// so it can be answered, and it mills each opponent of its controller
// as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf32524e-0df9-4af8-9371-84feeb0a86c6",
		Name:         "Fractured Sanity",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b12EachOpponentMills(ctx.Game, item, 14)
		},
		Activated: []ActivatedAbility{Cycling("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventCycle, Self,
				"Fractured Sanity — cycled: each opponent mills four cards",
				func(g *game.Game, item *game.StackItem) error {
					return b12EachOpponentMills(g, item, 4)
				})),
		},
	})
}
