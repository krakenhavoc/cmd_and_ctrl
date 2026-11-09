package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Velis Vel — Plane — Lorwyn (Planechase Anthology, #4870):
//
//	"Each creature gets +1/+1 for each other creature on the battlefield
//	 that shares at least one creature type with it. (For example, if two
//	 Elemental Shamans and an Elemental Spirit are on the battlefield,
//	 each gets +2/+2.)
//	 Whenever chaos ensues, target creature gains all creature types until
//	 end of turn."
//
// A plane (ADR 0216 §6). The static is Coat of Arms' count (one per other
// creature, however many types they share, the ruling). The chaos ability
// is a layer-4 all-creature-types mod, which is not changeling (the ruling)
// and so grants no ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c4e174e-19fb-4c0a-b762-ef6d927392c9",
		Name:         "Velis Vel",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Each creature gets +1/+1 for each other creature on the battlefield that shares at least one creature type with it.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature()
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				n := sharedCreatureTypeCount(g, target)
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("Velis Vel — target creature gains all creature types until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return untilEndOfTurn(ctx, id, nil, "Velis Vel — all creature types until end of turn", game.AllCreatureTypesMod())
				}),
				TargetCreature("target creature")),
		},
	})
}
