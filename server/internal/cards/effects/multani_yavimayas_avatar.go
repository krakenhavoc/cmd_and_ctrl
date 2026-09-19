package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Multani, Yavimaya's Avatar — Legendary Creature — Elemental Avatar
// {4}{G}{G}, 0/0 (EDHREC rank 1849):
//
//	"Reach, trample
//	 Multani gets +1/+1 for each land you control and each land card
//	 in your graveyard.
//	 {1}{G}, Return two lands you control to their owner's hand:
//	 Return this card from your graveyard to your hand."
//
// The lands deck's big body. Two keywords on PrintedKeywords; the
// size is a layer 7c modify summing the controller's lands on the
// battlefield (b03LandsControlled) and the land cards in their
// graveyard (b17LandCardsInGraveyard), recomputed with the layer
// cache.
//
// Sandbox simplification, declared (the Ketria Triome posture, one
// whole ability omitted): the graveyard ability is not implemented.
// An ability activated from the graveyard has no shape (the
// Gravecrawler gap), and "return two lands you control to their
// owner's hand" as a cost has none either. Weaker than printed,
// never stronger — Multani is recast from the command zone or
// reanimated like any other creature.
//
// With no land on the battlefield and none in the graveyard Multani
// is the printed 0/0 he prints, and the next state-based check puts
// him into the graveyard (CR 704.5f), as in paper — the modify adds
// nothing to a body that is already zero. That was an engine gap
// until #691, when the toughness check stopped reading every printed
// 0/0 as the importer's stand-in (game.Card.ToughnessIsKnown).
//
// One engine gap left, shared with its neighbours and not the card's:
// the layer cache is invalidated by battlefield motion, counters,
// taps and turn changes, not by a land reaching a graveyard from a
// hand or a library, so a milled land shows on Multani's size at the
// next recompute.
func init() {
	Register(Spec{
		OracleID:        "4b8bf64b-4800-45ff-81c6-2857f34999b5",
		Name:            "Multani, Yavimaya's Avatar",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The last ability isn't implemented — Multani can't be returned from your graveyard by bouncing two lands."},
		PrintedKeywords: []string{"reach", "trample"},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b03LandsControlled(g, source.Controller) + b17LandCardsInGraveyard(g, source.Controller)
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
