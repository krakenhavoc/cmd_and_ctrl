package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pride Sovereign — Creature — Cat {2}{G}, 2/2:
//
//	"This creature gets +1/+1 for each other Cat you control.
//	 {W}, {T}, Exert this creature: Create two 1/1 white Cat creature
//	 tokens with lifelink. (An exerted creature won't untap during your
//	 next untap step.)"
//
// The anthem counts the OTHER Cats its controller controls as the layer
// pass reads them (layer 7c, after layer 4 typed them, so a changeling
// counts), the Coat of Arms pattern. The tokens are Cats, so each grows
// the Sovereign. The activation pays {W}, {T} and an exert (ADR 0130 §4,
// CR 701.43a) at announce.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a131c32d-2b4f-4ee4-aab2-ab05ab010978",
		Name:         "Pride Sovereign",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 for each other Cat you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := otherCatsControlled(g, source)
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}, Exert this creature: Create two 1/1 white Cat creature tokens with lifelink.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost(), effects.ExertThis()),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 2},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Cat with lifelink"), N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// otherCatsControlled counts the creatures other than `source` that its
// controller controls with the Cat subtype, reading the live battlefield
// inside the layer pass (sharedCreatureTypeCount says why).
func otherCatsControlled(g *game.Game, source *game.Card) int {
	if g == nil || g.Battlefield == nil || source == nil {
		return 0
	}
	// "Cat" is a word of the source's text (CR 612.2, #3984).
	cat := source.TextSubtype("Cat")
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != source.InstanceID && c.Controller == source.Controller && c.IsCreature() && c.HasSubtype(cat) {
			n++
		}
	}
	return n
}
