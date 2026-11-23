package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serrated Scorpion — Creature — Scorpion {B}, 1/2 (Commander Masters):
//
//	"When this creature dies, it deals 2 damage to each opponent and you
//	 gain 2 life."
//
// A dies trigger: the Scorpion (as it last existed, CR 608.2h) deals 2 damage
// to each opponent, then its controller gains 2 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "abbdc656-02e9-43a0-8e2c-263d6c37c273",
		Name:         "Serrated Scorpion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Serrated Scorpion — it deals 2 damage to each opponent and you gain 2 life", serratedScorpionDrain),
		},
	})
}

func serratedScorpionDrain(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := g.DamageInstanceForEffect(func() error {
		for _, p := range ctx.Opponents() {
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: p, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return effects.GainLife{Amount: 2}.Apply(ctx)
}
