package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lord of Extinction — Creature — Elemental {3}{B}{G}, */* (EDHREC
// rank 3694):
//
//	"Lord of Extinction's power and toughness are each equal to the
//	 number of cards in all graveyards."
//
// Tarmogoyf's bigger cousin: a Layer 7a characteristic-defining
// ability that SETS power and toughness to the card count across
// every seat's graveyard — eliminated players' included, since
// their cards are still in a graveyard — so an anthem (7c) and its
// own counters (7d) stack on top. The printed */* ships as 0/0 and
// the CDA overrides it at every recompute; with every graveyard
// empty it really is a 0/0 and dies to the toughness check, as
// printed.
//
// That last sentence was written before it was true. The toughness
// check skipped every `*` creature as the importer's stand-in, this
// one included — and skipped its damage checks with it, so an empty
// graveyard left an unkillable 0/0. #690 closed it: a layer 7a or 7b
// effect DEFINES the body, and a body the engine computed is not a
// stand-in (Characteristic.PTDefined, game.Card.ToughnessIsKnown).
//
// This carried a declared simplification until #1117: the layer cache
// was invalidated by battlefield motion, counters, attachments, taps
// and turn changes, not by a card reaching a graveyard from a hand, a
// library or the stack — so a mill, a discard or an instant resolving
// showed on the Lord's size at the next recompute rather than at
// once, while a creature DYING (a battlefield exit) counted
// immediately, which is what made the gap so hard to see. A graveyard
// crossing now bumps the layer version on its own, with no flag for
// this card to declare, and the caveat is gone.
func init() {
	Register(Spec{
		OracleID:     "ea5e3401-bd6c-47bb-a52a-8eec5f09455d",
		Name:         "Lord of Extinction",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := b35CardsInAllGraveyards(g)
				c.Power = n
				c.Toughness = n
			},
		}},
	})
}
