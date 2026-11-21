package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Splash Portal — Instant {1}{U} (Bloomburrow, #4467):
//
//	"Exile target creature you control, then return it to the battlefield
//	 under its owner's control. If that creature is a Bird, Frog, Otter,
//	 or Rat, draw a card."
//
// The return is the exile's continuation and the question is asked of the
// card as it exists after returning (the ruling): a token never comes back
// and draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29b537f5-4d9f-40ed-b38b-e8faa2b5af7d",
		Name:         "Splash Portal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return effects.ExileTarget{
					Target:   t.ID,
					ThenTail: effects.WithCarry(ctx.ItemTail(effects.SplashPortalReturn), t.ID),
				}.Apply(ctx)
			}
			return nil
		},
	})
}
