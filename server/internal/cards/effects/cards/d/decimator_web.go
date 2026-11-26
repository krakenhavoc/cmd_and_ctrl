package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Decimator Web — Artifact {4} (Mirrodin Besieged, S82 OLD b466, #5576):
//
//	"{4}, {T}: Target opponent loses 2 life, gets a poison counter, then mills
//	 six cards."
//
// Three instructions in printed order (CR 608.2c) on the one opponent
// target: life loss (not damage), a poison counter put by the controller
// (CR 122.1f), then a six-card mill (CR 701.17a). A target no longer legal
// fizzles the ability (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eeabf42a-b1ac-4579-b82a-aabf9052ec20",
		Name:         "Decimator Web",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Target opponent loses 2 life, gets a poison counter, then mills six cards.",
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, LifeLoss: 2}),
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					if err := g.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -2); err != nil {
						return err
					}
					if err := g.AddPlayerCounterByForEffect(ctx.Controller(), t.ID, game.CounterPoison, 1); err != nil {
						return err
					}
					return effects.MillCards{Player: t.ID, N: 6}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
