package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Souls of the Lost — Creature — Spirit {1}{B}, */*+1:
//
//	"As an additional cost to cast this spell, discard a card or
//	 sacrifice a permanent. Fathomless descent — Souls of the Lost's
//	 power is equal to the number of permanent cards in your graveyard
//	 and its toughness is equal to that number plus 1."
//
// The either/or cost (ADR 0100 §2), and a layer-7a characteristic-
// defining ability (CR 604.3, 613.4a) — Tarmogoyf's shape, counting
// descend's permanent cards (permanentCardsInGraveyard, CR 110.4) in
// its controller's graveyard. A permanent sacrificed or a card
// discarded to pay the cost is in the graveyard by the time the
// creature enters, and counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c05d5a51-8e39-4475-9743-1adc60283c5a",
		Name:         "Souls of the Lost",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			DiscardCost(1).Keyed("discard"),
			SacrificeCost("a permanent").Keyed("sacrifice"),
		),
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := permanentCardsInGraveyard(g, source.Controller)
				c.Power = n
				c.Toughness = n + 1
			},
		}},
	})
}
