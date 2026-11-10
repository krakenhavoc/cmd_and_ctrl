package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "29b537f5-4d9f-40ed-b38b-e8faa2b5af7d",
		Name:         "Splash Portal",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return ExileTarget{
					Target:   t.ID,
					ThenTail: withCarry(ctx.ItemTail(splashPortalReturn), t.ID),
				}.Apply(ctx)
			}
			return nil
		},
	})
}
