package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Outcaster Greenblade — Creature — Human Mercenary {2}{G}, 1/2
// (Outlaws of Thunder Junction, #4468):
//
//	"When this creature enters, search your library for a basic land
//	 card or a Desert card, reveal it, put it into your hand, then
//	 shuffle.
//	 This creature gets +1/+1 for each Desert you control."
//
// The search is Map the Frontier's predicate, a basic land card or any
// card with the Desert land type, for one card to your hand. The bonus
// is Adelbert Steiner's layer 7c count, read live, so a Desert that
// enters or leaves changes it at once. Every Desert you control counts,
// creature or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "855a2579-8f78-46ce-a3a8-eddc808bd7b4",
		Name:         "Outcaster Greenblade",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Outcaster Greenblade gets +1/+1 for each Desert you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				desert := source.TextSubtype("Desert")
				n := effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtype(desert) })
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Outcaster Greenblade — search your library for a basic land card or a Desert card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.B34IsBasicLandOrDesertCard,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Outcaster Greenblade — a basic land card or a Desert card",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
