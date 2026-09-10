package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zombify — Sorcery {3}{B}:
//
//	"Return target creature card from your graveyard to the
//	 battlefield."
//
// The plain version of the effect, and the one that shows what the
// extra words on Reanimate are buying: no life loss, but also YOUR
// graveyard only, so it can never take an opponent's creature.
//
// Because the target is restricted to the caster's own pile, the
// creature's owner and its new controller are the same player, and
// the controller argument is redundant. It is passed explicitly
// anyway — the alternative is relying on a default that is only
// correct for half this family.
func init() {
	Register(Spec{
		OracleID: "bb95db4d-5017-4121-bf79-d68476602d8c",
		Name:     "Zombify",
		Targets:  targetCreatureInYourGraveyard(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			_, _ = reanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
