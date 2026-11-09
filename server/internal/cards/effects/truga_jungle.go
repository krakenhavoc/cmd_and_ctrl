package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Truga Jungle — Plane — Ergamon (#4870):
//
//	"All lands have "{T}: Add one mana of any color."
//	 Whenever chaos ensues, reveal the top three cards of your library. Put all
//	 land cards revealed this way into your hand and the rest on the bottom of
//	 your library in any order."
//
// The static grants the any-color mana ability to every land, whoever controls
// it (the grant is the one Chromatic Lantern's lands get). The chaos ability
// is RevealTopThenTakeToHand: the planar controller reveals, the land cards go
// to hand with no choice, and the rest go to the bottom in an order they pick.
// A library with fewer than three cards reveals what it has.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d858e6db-9842-41d1-8006-574c8dcd98ed",
		Name:         "Truga Jungle",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{AnyColorManaGrant(trugaJungleAnyColor)},
		Static: []game.StaticAbility{Labelled(GrantAbilities(
			func(target *game.Card, _ *game.Game, _ *game.Card) bool { return target.IsLand() },
			trugaJungleAnyColor), "All lands have \"{T}: Add one mana of any color.\"")},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Truga Jungle — reveal the top three cards of your library; put all land cards revealed this way into your hand and the rest on the bottom in any order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return RevealTopThenTakeToHand(ctx, item.Controller, 3,
						func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsLand() },
						"Truga Jungle — the top three cards of your library")
				}),
		},
	})
}

// trugaJungleAnyColor is the granted bundle's key. An on-disk key.
const trugaJungleAnyColor = "truga-jungle/any-color"
