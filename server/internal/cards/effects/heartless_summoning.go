package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heartless Summoning — Enchantment {1}{B}:
//
//	"Creature spells you cast cost {2} less to cast.
//	 Creatures you control get -1/-1."
//
// The drawback is the point: a deck that plays this wants creatures
// whose value is in an ETB trigger rather than in a body, and the
// -1/-1 is what stops it from simply being a two-mana discount.
//
// Both halves are declarative — a cost modifier and a Layer 7c
// static — and they live in different engines for the reason ADR
// 0042 gives: the -1/-1 changes a characteristic of an object on the
// battlefield, the discount changes what someone pays for an object
// in their hand.
//
// Worth noticing on this card specifically: the {2} reduction spends
// GENERIC mana and stops at zero, so a {B} Carrion Feeder still
// costs {B}. Heartless Summoning does not make one-drops free.
func init() {
	Register(Spec{
		OracleID:     "fa1be67e-a07e-42ce-88be-446acd643dc6",
		Name:         "Heartless Summoning",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Creature spells you cast cost {2} less to cast.",
				YourSpell(), CreatureSpell()),
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power--
				c.Toughness--
			},
		}},
	})
}
