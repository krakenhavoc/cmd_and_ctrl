package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anguished Unmaking — Instant {1}{W}{B}:
//
//	"Exile target nonland permanent. You lose 3 life."
//
// Exile rather than destroy is why this sees play over cheaper
// removal — it answers indestructible, regeneration, and
// death-trigger payoffs alike.
//
// The life loss is NOT a cost: it is the spell's second sentence, so
// a single-target spell whose target has become illegal does not
// resolve at all and the caster keeps their 3 life (CR 608.2b). The
// target guard below therefore wraps only the exile, and the engine's
// fizzle check upstream is what skips the whole effect — see
// TestAnguishedUnmakingFizzlesEntirelyWhenTargetLeaves.
func init() {
	Register(Spec{
		OracleID: "ad09b3c3-c8e7-481c-8c45-e7f234935117",
		Name:     "Anguished Unmaking",
		Targets:  TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				if err := (ExileTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -3)
		},
	})
}
