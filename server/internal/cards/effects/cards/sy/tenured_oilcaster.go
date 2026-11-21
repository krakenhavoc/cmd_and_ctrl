package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tenured Oilcaster — Creature — Phyrexian Wizard {3}{B}, 2/4 (March of
// the Machine):
//
//	"Menace
//	 This creature gets +3/+0 as long as an opponent has eight or more
//	 cards in their graveyard.
//	 Whenever this creature attacks or blocks, each player mills a card."
//
// The static is Blackbloom Rogue's. The trigger is one ability watching
// both declarations (Carrion Rats); the mill is every live player's, the
// controller's included, in turn order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7fb837d7-e2c9-40f8-8a7d-156b3aa86830",
		Name:            "Tenured Oilcaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{{
			Label:    "This creature gets +3/+0 as long as an opponent has eight or more cards in their graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					effects.AnOpponentHasGraveyardOfAtLeast(g, source.Controller, 8)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 3
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
				}, "Tenured Oilcaster — each player mills a card",
				func(g *game.Game, item *game.StackItem) error { return effects.B02EachPlayerMills(g, item, 1) }),
		},
	})
}
