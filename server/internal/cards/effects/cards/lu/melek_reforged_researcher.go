package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "23140c6f-c674-4d13-9b78-bce0ffcb981f",
		Name:         "Melek, Reforged Researcher",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Melek's power and toughness are each equal to twice the number of instant and sorcery cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 2 * effects.FandanielInstantsAndSorceries(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(3, "The first instant or sorcery spell you cast each turn costs {3} less to cast.",
				effects.YourSpell(), effects.InstantOrSorcerySpell(), func(q game.CostQuery) bool {
					return q.Game == nil || q.Game.CastTallyFor(q.Controller).InstantSorcery == 0
				}),
		},
	})
}
