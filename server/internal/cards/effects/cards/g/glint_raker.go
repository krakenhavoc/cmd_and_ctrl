package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glint Raker — Creature — Drake {3}{U}, 1/3 (The Brothers' War Commander, S82):
//
//	"Flying
//	 This creature gets +X/+0, where X is the greatest mana value among
//	 artifacts you control.
//	 Whenever this creature deals combat damage to a player, you may reveal
//	 that many cards from the top of your library. Put an artifact card
//	 revealed this way into your hand and the rest into your graveyard."
//
// The bonus is a layer 7c static (Carrion Grub's shape), read off the
// artifacts you control as they are now (a token with no mana cost is 0).
// "That many" is the combat damage the Raker dealt, read off the event; a
// library with fewer cards is revealed whole. The whole reveal is public and
// the artifact card you take is mandatory when there is one (a choice when
// there are several); every other revealed card, and all of them when none is
// an artifact, goes to the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e0a95a8e-5867-481b-9e85-e2d5386362b8",
		Name:            "Glint Raker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			Label:     "This creature gets +X/+0, where X is the greatest mana value among artifacts you control.",
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.Controller == source.Controller && p.IsArtifact() && p.ManaValue() > n {
						n = p.ManaValue()
					}
				}
				c.Power += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverThisDealsCombatDamageToAPlayer("Glint Raker — reveal that many cards from the top of your library; put an artifact card revealed this way into your hand and the rest into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := ctx.Trigger().Event.Amount
					if n <= 0 {
						return nil
					}
					player := ctx.Controller()
					revealed := g.RevealTopOfLibraryForEffect(player, ctx.Source(), n, "Glint Raker — reveals that many cards from the top of their library")
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  revealed,
						Match:  effects.Artifact(),
						Max:    1,
						Label:  "Glint Raker — put an artifact card revealed this way into your hand",
						Key:    effects.TakeRestGraveyard,
					}.Apply(ctx)
				}), "Glint Raker — reveal that many cards from the top of your library?"),
		},
	})
}
