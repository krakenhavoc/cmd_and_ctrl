package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Commander Sofia Daguerre — Legendary Creature — Human Pilot {3}{W}, 1/3
// (Fallout):
//
//	"Flash
//	 Crash Landing — When Commander Sofia Daguerre enters, destroy up to one
//	 target legendary permanent. That permanent's controller creates a Junk
//	 token."
//
// The controller is read before the destroy (Generous Gift's order), so the
// Junk goes to whoever controlled the permanent, whether or not it was
// destroyed (an indestructible one survives and its controller still gets
// the Junk). With no target chosen, or the target gone, nothing happens
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a3fdb633-7fe9-4e72-a3a3-aa17b6f17864",
		Name:            "Commander Sofia Daguerre",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Commander Sofia Daguerre — Crash Landing: destroy up to one target legendary permanent; its controller creates a Junk token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
						return nil
					}
					target := item.Targets[0].ID
					controller, ok := effects.ControllerOfTarget(ctx, target)
					if !ok {
						return nil
					}
					if err := (effects.DestroyTarget{Target: target}).Apply(ctx); err != nil {
						return err
					}
					return effects.CreateToken{Controller: controller, Template: effects.JunkToken(), N: 1}.Apply(ctx)
				}),
				effects.TargetPermanent("up to one target legendary permanent", effects.Legendary()).WithCount(0, 1)),
		},
	})
}
