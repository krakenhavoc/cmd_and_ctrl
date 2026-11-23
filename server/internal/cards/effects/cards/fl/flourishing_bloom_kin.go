package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flourishing Bloom-Kin — Creature — Plant Elemental {1}{G}, 0/0
// (Murders at Karlov Manor, #4469):
//
//	"This creature gets +1/+1 for each Forest you control.
//	 Disguise {4}{G}
//	 When this creature is turned face up, search your library for up to
//	 two Forest cards and reveal them. Put one of them onto the
//	 battlefield tapped and the other into your hand, then shuffle."
//
// The size is Blanchwood Armor's count (forestsControlledBy, the Forest
// land type read through the card's text, CR 612.2) on the Bloom-Kin
// itself, a layer 7c modify. Face down it is a plain 2/2 with ward {2}
// and none of this (CR 708.2a).
//
// The face-up trigger (CR 708.8) is Cultivate's two searches with any
// Forest card, not only a basic one: the first card goes onto the
// battlefield tapped, so a player who finds only one puts it there (the
// ruling), and the second goes to hand before the one shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "66e74e8f-036e-4b26-a4a8-2365891765bc",
		Name:             "Flourishing Bloom-Kin",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{4}{G}")},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 for each Forest you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.ForestsControlledBy(g, source)
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsTurnedFaceUp("Flourishing Bloom-Kin — search for up to two Forest cards: one onto the battlefield tapped, the other into your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.TwoLandSearch(effects.NewContext(g, item), "Flourishing Bloom-Kin", "Forest card", effects.IsLandWithSubtype("Forest"), bloomKinHandHalf)
				}),
		},
	})
}

// bloomKinHandHalf is the second search: the other Forest card into your
// hand, then the shuffle. An on-disk key.
var bloomKinHandHalf = effects.LandToHandTail("flourishing-bloom-kin/forest-to-hand", "Flourishing Bloom-Kin",
	"Forest card", effects.IsLandWithSubtype("Forest"))
