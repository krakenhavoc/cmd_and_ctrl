package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anchor to Reality — Sorcery {2}{U}{U}:
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature.
//	 Search your library for an Equipment or Vehicle card, put that card
//	 onto the battlefield, then shuffle. If it has mana value less than
//	 the sacrificed permanent's mana value, scry 2."
//
// The comparison reads the sacrificed permanent's mana value as it last
// existed on the battlefield (CR 608.2h), from the payment record (ADR
// 0113 §1), against the card that was put onto the battlefield; a card
// with {X} in its cost has X = 0 (the 2022-02-18 ruling). Nothing found
// means nothing to compare, so no scry.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "6783f559-33d7-4b13-9a91-02b821e97163",
		Name:           "Anchor to Reality",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			sacrificedMV := sacrificedManaValue(ctx)
			return SearchLibrary{
				Player: item.Controller,
				Predicate: func(c game.Card) bool {
					return c.HasSubtype("Equipment") || c.HasSubtype("Vehicle")
				},
				Dest:    game.ZoneBattlefield,
				Limit:   1,
				Shuffle: true,
				Reason:  "Anchor to Reality — an Equipment or Vehicle card",
				Then: func(g *game.Game, found []uuid.UUID) error {
					if len(found) == 0 {
						return nil
					}
					c, ok := g.LookupCardForEffect(found[0])
					if !ok || c.ManaValue() >= sacrificedMV {
						return nil
					}
					return Scry{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
				},
			}.Apply(ctx)
		},
	})
}
