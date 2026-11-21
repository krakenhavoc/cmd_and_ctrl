package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adversary by Algorithm — Sorcery {2}{G} (Star Trek, #4973):
//
//	"Create a 3/3 colorless Illusion artifact creature token.
//	 Beam me up {3}{G} (You may cast this card from your graveyard for
//	 {3}{G} if you also return a creature you control to its owner's
//	 hand. Then exile this spell.)"
//
// The token is the printed ttrk Illusion: a 3/3 colorless artifact
// creature with no abilities. Beam me up is BeamMeUp: a cast from the
// graveyard for {3}{G} plus the return of a creature you control as a
// cost, exiled as the spell leaves the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "37f3eda9-3242-4d62-846f-935543d8f4fd",
		Name:             "Adversary by Algorithm",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.BeamMeUp("{3}{G}")},
		OnResolve:        effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("3/3 colorless Illusion artifact"), N: 1}),
	})
}
