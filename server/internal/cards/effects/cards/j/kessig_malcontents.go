package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kessig Malcontents — Creature — Human Warrior {2}{R}, 3/1 (Avacyn
// Restored, S82 OLD, #5570):
//
//	"When this creature enters, it deals damage to target player or
//	 planeswalker equal to the number of Humans you control."
//
// The Humans are counted as the ability resolves (Horn of Gondor's count),
// through the source's text (CR 612.2), and the Malcontents is one of them.
// The damage is the creature's. A target gone by resolution means no damage
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eedaf405-4aa2-426a-af82-b0f92de7b6c3",
		Name:         "Kessig Malcontents",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Kessig Malcontents — damage to target player or planeswalker equal to the number of Humans you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := effects.FirstLegalTarget(ctx)
						if !ok {
							return nil
						}
						n := effects.B43CreaturesOfSubtypeControlled(g, item.Controller, ctx.TextSubtype("Human"))
						if n <= 0 {
							return nil
						}
						return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
					}),
				effects.TargetPlayerOrPlaneswalker(),
			),
		},
	})
}
