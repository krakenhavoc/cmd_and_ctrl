package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serra Avatar — Creature — Avatar {4}{W}{W}{W}, */*:
//
//	"Serra Avatar's power and toughness are each equal to your life
//	 total.
//	 When Serra Avatar is put into a graveyard from anywhere, shuffle it
//	 into its owner's library."
//
// The first line is a layer 7a characteristic-defining ability (CR
// 604.3), read on every layer pass, so it moves with the life total.
// It works everywhere (CR 113.6a): off the battlefield the card has no
// controller and "your" is its owner (CR 108.4a). The second line is
// the graveyard trigger (#2038,
// WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "c363725f-15cf-4955-aaf1-536b1f8daaed",
		Name:         "Serra Avatar",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Power and toughness are each equal to your life total.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := yourLifeTotal(g, source)
				c.Power, c.Toughness = n, n
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn("Serra Avatar — shuffle it into its owner's library"),
		},
	})
}

// yourLifeTotal is the life total of `source`'s controller, or of its
// owner when it has none (CR 108.4a); 0 when that player is gone.
func yourLifeTotal(g *game.Game, source *game.Card) int {
	you := source.Controller
	if you == uuid.Nil {
		you = source.Owner
	}
	if p := g.PlayerByIDForEffect(you); p != nil {
		return p.Life
	}
	return 0
}
