package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Melek, Reforged Researcher — Legendary Creature — Weird Detective
// {3}{U}{R}, */* (Murders at Karlov Manor, #4469):
//
//	"Melek's power and toughness are each equal to twice the number of
//	 instant and sorcery cards in your graveyard.
//	 The first instant or sorcery spell you cast each turn costs {3} less
//	 to cast."
//
// The power and toughness are Dragonfly Swarm's characteristic-defining
// ability (layer 7a), counting the instant and sorcery cards in its
// controller's graveyard live (fandanielInstantsAndSorceries). The
// discount is Artificer Class's "the first … spell you cast each turn":
// the caster's per-turn tally of instant and sorcery spells
// (game.CastTally.InstantSorcery) is still zero while the turn's first
// one is priced, because a cost is determined (CR 601.2f) before the
// spell is counted. An instant or sorcery cast earlier in the turn, even
// before Melek arrived, was the first, so nothing later that turn is
// discounted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23140c6f-c674-4d13-9b78-bce0ffcb981f",
		Name:         "Melek, Reforged Researcher",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Melek's power and toughness are each equal to twice the number of instant and sorcery cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 2 * fandanielInstantsAndSorceries(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		CostModifiers: []game.CostModifier{
			CostsLess(3, "The first instant or sorcery spell you cast each turn costs {3} less to cast.",
				YourSpell(), InstantOrSorcerySpell(), func(q game.CostQuery) bool {
					return q.Game == nil || q.Game.CastTallyFor(q.Controller).InstantSorcery == 0
				}),
		},
	})
}
