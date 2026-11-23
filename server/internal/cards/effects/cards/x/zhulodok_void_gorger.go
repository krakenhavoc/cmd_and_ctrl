package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zhulodok, Void Gorger — Legendary Creature — Eldrazi {5}{C}, 7/4
// (Commander Masters):
//
//	"Colorless spells you cast from your hand with mana value 7 or greater
//	 have "Cascade, cascade.""
//
// Quandrix, the Proof's and Imoti's shape: the Eldrazi watches your casts
// and puts a cascade (cascadeEffect, CR 702.85) on the stack for each of the
// two granted instances, so a qualifying spell cascades twice, each its own
// trigger (CR 702.85c). The test is on the spell as cast: colorless
// (effective colours), cast from its owner's hand, mana value 7 or greater
// with an {X} counted at its chosen value.
//
// No simplification.
func init() {
	qualifies := func(spell game.Card, source *game.Card, g *game.Game) bool {
		mv, ok := g.ManaValueForEffect(spell)
		return ok && mv >= 7 && spell.IsColorless() &&
			spell.Owner == source.Controller &&
			effects.CastFromOwnersHand()(g, source.Controller, spell)
	}
	effects.Register(effects.Spec{
		OracleID:     "ec726c54-987b-48ed-8ffa-ec73a5e35333",
		Name:         "Zhulodok, Void Gorger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GrantsCascade("Zhulodok, Void Gorger — first cascade", qualifies),
			effects.GrantsCascade("Zhulodok, Void Gorger — second cascade", qualifies),
		},
	})
}
